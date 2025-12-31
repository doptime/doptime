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
     -> Tokenize by Space (strings.Fields)       |
     -> Extract Name & @Directive                |
     ↓                                           |
   [Resolve Value Phase]                         |
     -> 1. Check Input (Direct Match)            |
     -> 2. Process @Directive:                   |
          -> a. Hybrid Check: Is it a Ref Key? --+ (Recursive Lookup)
          -> b. Fallback: Parse as Literal       |
               (Priority: Int -> Float -> Bool)  |
     ↓
[Decode Phase]
     -> Type Assertion / Weakly Typed Hook
     -> Numeric/Bool Bridge (Robustness)
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
* **逻辑**：在解析引用前检查 Key 是否存在。若存在，立即终止引用查找（视为未命中，回退到字面量或零值），实现非侵入式的防御。

## 3. 关键算法详解

### 3.1 标签解析 (`parseTag`)

v2 版本为了解决编辑器 JSON 校验问题及统一代码风格，严格采用 **空格分隔 (Space Separator)**。

* **实现**：使用 `strings.Fields(tag)` 进行分词。
* **逻辑**：
* 遍历分词结果。
* 以 `@` 开头的 Token 被识别为 `defaultDirective`。
* 第一个非 `@` 开头的 Token 被识别为 `name`。


* **示例**：
* `json:"id @uuid"` → Name="id", Directive="@uuid"
* `json:"@true"` → Name="" (Fallback to field name), Directive="@true"


* **注意**：不再支持逗号分隔。`json:"id, @val"` 会被解析为 Name="id,"，导致无法正确匹配字段。

### 3.2 智能字面量解析 (`parseDefaultLiteral`)

为了避免类型推断歧义（例如字符串 "1" 被 Go 的 `strconv.ParseBool` 误判为 true），v2 调整了解析优先级：

1. **Try Int64**：优先尝试解析为整数。若成功，返回 `int64`。（"1" → 1）
2. **Try Float64**：尝试解析为浮点数。（"1.5" → 1.5）
3. **Try Bool**：最后尝试解析为布尔值。（"true" → true）
4. **Fallback**：保持字符串原值。

### 3.3 混合值决议 (`resolveValue`)

这是 Mapper v2 的核心逻辑，统一了引用与字面量的处理入口。

**优先级逻辑：**

1. **Direct Lookup**：使用字段名的小写形式在 Map 中查找。
2. **Directive Processing (`@`)**：
* **Ref Check（引用检查）**：去掉 `@` 后，检查剩余字符串是否对应输入 Map 中的 Key。
* **若存在**：视为引用。递归获取值。


* **Literal Fallback（字面量回退）**：
* **若不存在**：视为字面量。调用 `parseDefaultLiteral` 进行转换。





## 4. 类型系统与弱类型转换

Mapper 内置了一套轻量级转换逻辑。

### 4.1 数值与布尔值的桥接 (`decodeNumericBridge` / `decodeInt`)

在处理默认值（解析为 `int64`/`bool`）赋值给结构体字段时，Mapper 增强了鲁棒性：

* **问题**：目标字段为 `int`，但默认值解析出来是 `bool(true)`；或者目标是 `float64`，默认值是 `int64(1)`。
* **方案**：
* **Int ↔ Float**：自动进行类型转换。
* **Bool → Numeric**：
* `true` 转换为 `1` (int) 或 `1.0` (float)。
* `false` 转换为 `0` (int) 或 `0.0` (float)。





### 4.2 特殊类型支持

* **JSON Number**：优先处理 `json.Number`，避免大整数转 float 导致的精度丢失。
* **Time Hook**：支持 RFC3339 及常见日期格式（`yyyy-MM-dd HH:mm:ss` 等）。

## 5. 安全设计总结

1. **哨兵错误模式**：使用 `errAssignedDirectly` 控制“直接赋值优化”的流程，避免脆弱的字符串匹配。
2. **AssignableTo 前置检查**：赋值前必检查类型兼容性，杜绝 Panic。
3. **Ref 循环防御**：通过 `refTracker` 防止配置层面的无限递归。
4. **空指令防御**：能够正确处理 `json:"@"` 等边缘 Case。

## 6. 性能与局限性

* **分词开销**：`strings.Fields` 在典型结构体规模下开销极低，且比手写的逗号解析器更健壮。
* **并发复用**：`Decoder` 实例包含状态（`refTracker`），因此**不是线程安全**的。请勿在并发环境下复用同一个 Decoder 实例。