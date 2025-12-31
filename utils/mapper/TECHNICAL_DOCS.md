# Mapper 技术实现与架构文档

本文档详细描述了 Mapper 库的内部设计原理、核心流程及关键技术决策。

## 1. 架构概览

Mapper 基于 Go 语言的 `reflect` 包实现，采用 **"Normalize (归一化) -> Resolve (决议) -> Decode (解码)"** 的三段式处理管线。

### 处理流程图

```text
[Input Data] 
     ↓ 
[Normalize Phase] 
     -> Handle nil/ptr 
     -> Build Case-Insensitive Map (inputWrapper)
     ↓
[Struct Iteration]
     ↓
   [Resolve Value Phase] <-----------------------+
     -> 1. Check Input (Direct Match)            |
     -> 2. Check Reference Default (@ref) -------+ (Recursive Lookup)
     -> 3. Check Literal Default (Smart Parse)
     ↓
[Decode Phase]
     -> Type Assertion / Weakly Typed Hook
     -> AssignableTo Check
     -> Recursive Decode (for structs)
     ↓
[Finalize]
     -> Fill Remain Field (Unused Keys)

```

## 2. 核心数据结构设计

### 2.1 `inputWrapper`：解决大小写悖论

Mapper 面临一个经典矛盾：**字段查找需要忽略大小写**（配置灵活性），但 **Remain 字段需要保留原始大小写**（数据完整性）。

为此，我们引入了 `inputWrapper` 结构作为中间层：

```go
type inputWrapper struct {
    val         reflect.Value // 实际数据值 (反射值)
    originalKey string        // 原始键名 (Case-Sensitive)
}

```

* **查找时**：使用 `strings.ToLower(key)` 在 `map[string]inputWrapper` 中查找。
* **填充 Remain 时**：读取 `originalKey` 写入结果 Map，确保输出与输入完全一致。

### 2.2 `refTracker`：防御性循环检测

为了防止用户配置 `A default=@B` 且 `B default=@A` 导致的栈溢出，`Decoder` 维护了一个临时的 `map[string]bool`。

* **Scope**：每次 `Decode` 调用生命周期内有效。
* **逻辑**：在解析 `@ref` 前检查 Key 是否存在。若存在，立即终止引用查找（返回 nil），实现非侵入式的防御。

## 3. 关键算法详解

### 3.1 智能值决议 (`resolveValue`)

这是 Mapper 的“大脑”，负责决定一个字段到底使用什么值。优先级逻辑如下：

1. **Direct Lookup (O(1))**：
使用字段名（tag 或 struct field name）的小写形式在 Map 中查找。如有，直接返回，并返回 `keyUsed` 以便后续排除。
2. **Reference Default (`@`)**：
* 解析 tag 中的 `@field`。
* **Cycle Check**：检查 `refTracker`。
* **Lookup**：去 Map 中查找引用的字段。
* **关键点**：如果引用成功，**被引用的 Key** 也会被标记为 `used`，防止数据重复进入 Remain。


3. **Literal Default (Smart Parse)**：
* 解析字面量（如 `"100"`、`"true"`）。
* **类型推断**：尝试将字符串预解析为 `bool`、`int64` 或 `float64`。
* **解决痛点**：解决了 `interface{}` 类型字段接收默认值时变成 `string` 类型的问题，确保 `default="100"` 赋给 `interface{}` 时是数字类型。



### 3.2 预处理阶段 (`prepareInputMap`)

* **Nil 安全**：如果输入为 `nil`，返回一个空的 `map` 而非报错。这允许结构体即使没有输入数据，也能根据 `default` 标签正确初始化。
* **结构体直连优化**：如果 Input 是结构体且类型与 Output 完全一致，直接 `out.Set(in)`，跳过繁琐的反射字段遍历。

### 3.3 Remain 填充机制

Remain 的填充发生在所有显式字段解析之后：

1. 遍历内部的 `dataMap` (key 为小写)。
2. 检查 Key 是否在 `usedKeys` 集合中。
* **注意**：`usedKeys` 包含了**被直接映射**的 Key 以及**被 `@ref` 引用过**的 Key。


3. 如果未使用，提取 `inputWrapper.originalKey`，将值写入 Remain Map。

## 4. 类型系统与弱类型转换

Mapper 不依赖第三方转换库，而是内置了一套轻量级转换逻辑 (`weaklyTypedHook` & `decodeBasic`)。

### 4.1 数值桥接 (`decodeNumericBridge`)

Go 的反射非常严格，`int` 无法直接赋值给 `float`。但在处理默认值时，字面量往往被解析为 `int64`（如 `default="0"`）。

* **问题**：当目标字段是 `float64` 而默认值解析为 `int64` 时，直接 `Set` 会 Panic。
* **方案**：内置 `decodeNumericBridge`，检测到 int/float 类型不匹配时，自动进行强制类型转换。

### 4.2 特殊类型支持

* **JSON Number**：优先处理 `json.Number`，避免大整数转 float 导致的精度丢失。
* **Time Hook**：支持 RFC3339 及常见日期格式（`yyyy-MM-dd HH:mm:ss` 等）。

## 5. 性能与局限性

### 性能考量

* **反射开销**：大量使用 `reflect`，性能略低于代码生成方案（如 easyjson）。
* **内存分配**：`prepareInputMap` 会重建 Map 索引，对超大 Map (10k+ keys) 有一定内存压力。
* **优化策略**：通过 `originalKey` 保存策略，避免了在 Remain 阶段重新反射 Key 的类型。

### 已知局限

* **Default 语法**：`default` 值中不能包含逗号 `,`，因为逗号是 struct tag 的分隔符。
* **复杂表达式**：不支持动态计算默认值（如 `default=${ENV_VAR}`），仅支持静态字面量和同级引用。

## 6. 安全设计总结

1. **AssignableTo 前置检查**：赋值前必检查类型兼容性，杜绝 Panic。
2. **Nil 指针保护**：所有反射操作前检查 `IsValid`。
3. **栈溢出保护**：通过 `refTracker` 防止配置层面的无限递归。