# Mapper v2: 技术实现与架构文档

本文档详细描述了 Mapper v2 库的内部设计原理、核心流程及关键技术决策。

## 1. 架构概览

Mapper 基于 Go 语言的 `reflect` 包实现，采用 **"Normalize (归一化) -> Resolve (决议) -> Decode (解码)"** 的三段式处理管线。

### 处理流程图

```text
[Input Data]
     ↓
[Normalize Phase]
     -> Handle nil/ptr
     -> Optimization: Direct Struct Assignment (Sentinel Error Check)
     -> Build Case-Insensitive Map (inputWrapper)
     ↓
[Struct Iteration]
     ↓
   [Tag Parsing] <-------------------------------+
     -> Normalize Separators (Comma -> Space)    |
     -> Extract Name & @Directive                |
     ↓                                           |
   [Resolve Value Phase]                         |
     -> 1. Check Input (Direct Match)            |
     -> 2. Process @Directive:                   |
          -> a. Hybrid Check: Is it a Ref Key? --+ (Recursive Lookup)
          -> b. Fallback: Parse as Literal (Int/Bool/Float)
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

我们引入了 `inputWrapper` 结构作为中间层：

```go
type inputWrapper struct {
    val         reflect.Value // 实际数据值 (反射值)
    originalKey string        // 原始键名 (Case-Sensitive)
}

```

* **查找时**：使用 `strings.ToLower(key)` 在 `map[string]inputWrapper` 中查找。
* **填充 Remain 时**：读取 `originalKey` 写入结果 Map，确保输出与输入完全一致。

### 2.2 `refTracker`：防御性循环检测

为了防止用户配置 `json:"A @B"` 且 `json:"B @A"` 导致的无限递归与栈溢出，`Decoder` 维护了一个临时的 `map[string]bool`。

* **Scope**：每次 `Decode` 调用生命周期内有效（非并发安全，单次解码独享）。
* **逻辑**：在解析引用前检查 Key 是否存在。若存在，立即终止引用查找（视为未命中），实现非侵入式的防御。

## 3. 关键算法详解

### 3.1 智能值决议 (`resolveValue`)

这是 Mapper v2 的核心逻辑，采用了**混合决议策略 (Hybrid Resolution Strategy)**，统一了引用与字面量的处理入口。

**优先级逻辑：**

1. **Direct Lookup (O(1))**：
使用字段名（Tag Name 或 Struct Field Name）的小写形式在 Map 中查找。如有，直接返回。
2. **Directive Processing (`@`)**：
若第一步未命中，且存在 `@xxx` 指令，进入混合判定：
* **Ref Check（引用检查）**：检查 `xxx` 是否存在于输入 Map 中（忽略大小写）。
* **若存在**：视为引用。检查 `refTracker` 防循环，标记该 Key 为 `used`，递归返回其值。


* **Literal Fallback（字面量回退）**：
* **若不存在**：视为字面量。调用 `parseDefaultLiteral` 将字符串智能转换为 `bool`, `int64`, `float64` 或保持 `string`。





### 3.2 标签解析 (`parseTag`)

v2 版本对 Tag 解析进行了增强，以支持更自然的语法：

* **分隔符归一化**：首先将所有逗号 `,` 替换为空格，然后使用 `strings.Fields` 进行分词。这使得 `json:"id @field"` 和 `json:"id, @field"` 均被视为合法。
* **智能识别**：遍历分词结果，若某段以 `@` 开头，自动识别为默认值指令；否则识别为字段名。
* **空指令防御**：自动处理 `json:"@"` 这种空指令情况，防止切片越界 Panic。

### 3.3 预处理与哨兵错误 (`errAssignedDirectly`)

在 `prepareInputMap` 阶段，为了性能优化，如果 Input 结构体类型直接等同于 Output 结构体类型，我们会直接赋值。

* **v1 实现**：返回 `errors.New("assigned directly")`，依赖字符串匹配，脆弱。
* **v2 实现**：引入包级哨兵错误 `var errAssignedDirectly = errors.New("assigned directly")`。调用方使用 `errors.Is` 进行判断，符合 Go 最佳实践，更加健壮。

### 3.4 Remain 填充机制

Remain 的填充发生在所有显式字段解析之后：

1. 遍历内部的 `dataMap` (key 为小写)。
2. 检查 Key 是否在 `usedKeys` 集合中。
* **注意**：`usedKeys` 包含了**被直接映射**的 Key 以及**被 `@ref` 引用过**的 Key。


3. 如果未使用，提取 `inputWrapper.originalKey`，将值写入 Remain Map。

## 4. 类型系统与弱类型转换

Mapper 内置了一套轻量级转换逻辑 (`weaklyTypedHook` & `decodeBasic`)。

### 4.1 数值桥接 (`decodeNumericBridge`)

Go 的反射非常严格，`int` 无法直接赋值给 `float`。但在处理 `@` 字面量时，整数常被解析为 `int64`（如 `@0`）。

* **问题**：当目标字段是 `float64` 而默认值解析为 `int64` 时，直接 `Set` 会 Panic。
* **方案**：内置 `decodeNumericBridge`，检测到 int/float 类型不匹配时，自动进行强制类型转换。

### 4.2 特殊类型支持

* **JSON Number**：优先处理 `json.Number`，避免大整数转 float 导致的精度丢失。
* **Time Hook**：支持 RFC3339 及常见日期格式（`yyyy-MM-dd HH:mm:ss` 等）。

## 5. 性能与局限性

### 性能考量

* **反射开销**：大量使用 `reflect`，性能略低于代码生成方案。
* **分词开销**：v2 使用 `strings.Fields` 解析 Tag，虽然增加了字符串处理，但在典型结构体规模下（<100 字段）开销可忽略不计。
* **Map 重建**：`prepareInputMap` 会重建 Map 索引，对超大 Map 有一定内存压力。

### 已知局限

* **并发复用**：`Decoder` 实例包含状态（`refTracker`），因此**不是线程安全**的。请勿在并发环境下复用同一个 Decoder 实例。
* **复杂表达式**：不支持动态计算默认值（如 `@${ENV_VAR}`），仅支持静态字面量和同级引用。

## 6. 安全设计总结

1. **AssignableTo 前置检查**：赋值前必检查类型兼容性，杜绝 Panic。
2. **哨兵错误模式**：使用 `errAssignedDirectly` 控制内部流程，避免魔术字符串。
3. **Ref 循环防御**：通过 `refTracker` 防止配置层面的无限递归。
4. **Tag 解析防御**：对空 Tag、格式错误的 Tag 进行降级处理而非 Panic。
