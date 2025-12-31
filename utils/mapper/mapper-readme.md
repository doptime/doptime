# Mapper v2: 工业级 Go 结构体解码库

**Mapper** 是一个轻量级但功能强大的 Go 语言库，用于将非结构化数据（如 `map[string]interface{}`、JSON 解析结果）解码为强类型的 Go 结构体。

它的设计理念是 **"JSON + Mapstructure + Remain + Smart Defaults"**，完美解决了“固定字段 + 动态扩展”以及“配置合并”等复杂场景下的痛点。

> ⚠️ **v2 版本重要变更**：v2 版本引入了全新的 `@` 语法并移除了 `default=` 写法。请参考文末的 [迁移指南](#7-v1-到-v2-迁移指南)。

## 1. 核心特性

* **零配置使用**：通过 `mapper.Decode` 一键直达。
* **极简语法 (Unified @ Syntax)**：
    * **字面量默认**：支持 `json:"port @8080"`，自动推断并转换为目标类型（支持 int/bool/float）。
    * **字段引用**：支持 `json:"host @ip"`，若当前字段缺失，自动引用同级字段 `ip` 的值。
    * **灵活分隔**：推荐使用空格分隔 Tag，如 `json:"name @default"`，同时也兼容传统的逗号分隔。
* **Remain 机制**：
    * 自动将未被定义的字段捕获到 `map` 中。
    * **保留原始 Key 大小写**：输入是什么 Key，Remain 中就是什么 Key（解决了标准库转小写的问题）。
* **弱类型转换 (Weakly Typed)**：
    * `string` ↔ `int/float/bool` 自动互转。
    * `string` → `time.Time` 自动解析（支持 RFC3339 等多种格式）。
    * 支持 `json.Number` 的无损处理。
* **工业级健壮性**：
    * **循环引用检测**：防止配置错误（A 引用 B，B 引用 A）导致栈溢出。
    * **Nil 安全**：输入为 `nil` 时依然能正确应用默认值。

## 3. 快速开始

### 3.1 基础映射与智能默认值

Mapper 使用 `json` 标签定义字段名，直接使用 `@` 符号定义默认策略。

```go
package main

import (
	"fmt"
	"your_project/pkg/mapper"
)

type Config struct {
	// 1. 基础映射
	Name string `json:"name"`

	// 2. 字面量默认值 (自动解析为 int 类型)
	// 语法：字段名 "port", 默认值 "@8080"
	Port int `json:"port @8080"`

	// 3. 引用默认值 (强力特性)
	// 如果 'host' 没传，尝试使用 'ip' 字段的值
	// 语法：字段名 "host", 默认引用 "@ip"
	Host string `json:"host @ip"`

	// 4. 仅定义默认值，字段名回退 (Fallback)
	// 语法：无显式字段名，字段名为 "Debug"，默认值为 "true"
	Debug bool `json:"@true"`
}

func main() {
	input := map[string]interface{}{
		"name": "MyApp",
		"ip":   "192.168.1.1", // 将被 Host 引用
		// "port" 缺失 -> 使用 8080
		// "host" 缺失 -> 引用 ip
		// "Debug" 缺失 -> 使用 true
	}

	var cfg Config
	if err := mapper.Decode(input, &cfg); err != nil {
		panic(err)
	}

	fmt.Printf("Config: %+v\n", cfg)
	// Output: Config: {Name:MyApp Port:8080 Host:192.168.1.1 Debug:true}
}

```

### 3.2 Remain 机制（处理动态字段）

当输入包含结构体未定义的字段时，`Remain` 字段会自动“兜底”。

**规则：**

1. 字段名必须为 `Remain` 或标签指定 `json:"remain"`。
2. 类型必须是 `map[string]interface{}`。
3. **特性**：Mapper 会保留输入数据的原始大小写（例如输入 "UserAge"，Remain 中 key 也是 "UserAge"）。
4. **注意**：被 `@` 引用的字段（如上例中的 `ip`）也会被标记为“已使用”，**不会**再次进入 Remain。

```go
type Product struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	// 捕获所有未被 ID 和 Name 消耗的字段
	Remain map[string]interface{} `json:"-"`
}

func main() {
	input := map[string]interface{}{
		"id":     100,
		"name":   "Phone",
		"Color":  "Red",  // 动态字段，保留大写 C
		"Weight": "200g", // 动态字段
	}

	var p Product
	mapper.Decode(input, &p)

	fmt.Println(p.Remain)
	// Output: map[Color:Red Weight:200g]
}

```

## 4. 核心逻辑详解

### 4.1 值决议优先级 (Resolution Logic)

对于任何字段，Mapper 按照以下优先级决定最终值：

1. **输入优先**：输入数据中存在该字段（忽略大小写匹配） → 使用输入值。
2. **引用默认 (`@ref`)**：标签定义了 `@other` 且 `other` 存在于输入中 → 使用 `other` 的值。
3. **字面默认 (`@val`)**：标签定义了 `@100` 或 `@true` → 使用解析后的字面量值。
4. **零值**：以上都不满足 → 保持 Go 零值。

> **注意**：如果通过 `@ref` 引用了某个字段，该被引用字段也会被标记为“已使用”，**不会**再出现在 `Remain` 中（避免数据重复）。

### 4.2 弱类型支持矩阵

`mapper` 会尽最大努力尝试转换类型，而不是直接报错。

| 输入类型 (Input) | 目标类型 (Struct Field) | 行为 |
| --- | --- | --- |
| `string` ("123") | `int` / `uint` / `float` | 自动解析数值 |
| `string` ("true") | `bool` | 转为 `true` |
| `json.Number` | `int/float/string` | 智能拆箱，保留精度 |
| `string` (RFC3339) | `time.Time` | 自动尝试多种格式解析 |
| `interface{}` | `interface{}` | 直接赋值 |

### 4.3 安全机制

1. **循环引用保护**：
如果配置错误导致循环引用（如 `A @B`, `B @A`），Mapper 会自动中断引用链并返回 `nil`，防止程序崩溃。
2. **Nil 输入处理**：
`mapper.Decode(nil, &out)` 是合法的。此时结构体字段将全部尝试应用默认值（`@` 指令）。

## 5. API 参考

### `func Decode(input interface{}, output interface{}) error`

最常用的入口函数。

* `input`: 输入数据，通常是 `map[string]interface{}`，也可以是 `nil`。
* `output`: 必须是结构体指针（`*struct`）。

### `type DecoderConfig`

如果需要更精细的控制，可以使用 `NewDecoder`。

* `TagName`: 默认为 "json"。
* `WeaklyTypedInput`: 默认为 `true`。

---

## 6. 注意事项与限制

1. **Tag 分隔符**：v2 版本推荐使用**空格**分隔字段名与默认值（如 `json:"name @default"`）。为了兼容性，逗号也会被视作分隔符处理，因此 `json:"name, @default"` 也是合法的。
2. **结构体指针**：`output` 参数必须是指针，且必须是可寻址的。
3. **并发安全**：`Decoder` 实例在 Decode 过程中是有状态的（用于循环检测），因此**不可**在多个 Goroutine 中并发复用同一个 `Decoder` 实例。请尽量使用 `mapper.Decode` 静态方法，或每次 `NewDecoder`。

## 7. v1 到 v2 迁移指南

v2 版本引入了破坏性变更（Breaking Changes），移除了 `default=` 语法。

**1. 语法替换**

* **旧 (v1)**: `json:"port, default=8080"`
* **新 (v2)**: `json:"port @8080"`
* **旧 (v1)**: `json:"host, default=@ip"`
* **新 (v2)**: `json:"host @ip"`

**2. 行为变化**

* v2 对 Tag 的解析更加宽松，支持空格分隔。
* 逗号不再是必须的分隔符，但会被兼容处理。
