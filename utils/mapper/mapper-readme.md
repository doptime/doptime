# Mapper: 工业级 Go 结构体解码库

**Mapper** 是一个轻量级但功能强大的 Go 语言库，用于将非结构化数据（如 `map[string]interface{}`、JSON 解析结果）解码为强类型的 Go 结构体。

它的设计理念是 **"JSON + Mapstructure + Remain + Smart Defaults"**，完美解决了“固定字段 + 动态扩展”以及“配置合并”等复杂场景下的痛点。

## 1. 核心特性

* **零配置使用**：通过 `mapper.Decode` 一键直达。
* **智能默认值 (Smart Defaults)**：
* **字面量默认**：支持 `default="8080"`，自动推断并转换为目标类型（支持 int/bool/float）。
* **字段引用**：支持 `default="@host"`，若当前字段缺失，自动引用同级字段 `host` 的值。


* **Remain 机制**：
* 自动将未被定义的字段捕获到 `map` 中。
* **保留原始 Key 大小写**：输入是什么 Key，Remain 中就是什么 Key（解决了标准库转小写的问题）。


* **弱类型转换 (Weakly Typed)**：
* `string`  `int/float/bool` 自动互转。
* `string`  `time.Time` 自动解析（支持 RFC3339 等多种格式）。
* 支持 `json.Number` 的无损处理。


* **工业级健壮性**：
* **循环引用检测**：防止配置错误（A 引用 B，B 引用 A）导致栈溢出。
* **Nil 安全**：输入为 `nil` 时依然能正确应用默认值。


## 3. 快速开始

### 3.1 基础映射与智能默认值

Mapper 使用 `json` 标签定义字段名，使用 `default` 定义默认策略。

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
	Port int `json:"port, default=8080"`

	// 3. 引用默认值 (强力特性)
	// 如果 'host' 没传，尝试使用 'ip' 字段的值
	Host string `json:"host, default=@ip"`

	// 4. 弱类型转换 ("true" -> bool)
	Debug bool `json:"debug"`
}

func main() {
	input := map[string]interface{}{
		"name":  "MyApp",
		"ip":    "192.168.1.1", // 将被 Host 引用
		"debug": "true",        // 字符串自动转 bool
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

1. **输入优先**：输入数据中存在该字段（忽略大小写匹配）  使用输入值。
2. **引用默认 (`@ref`)**：标签定义了 `default=@other` 且 `other` 存在于输入中  使用 `other` 的值。
3. **字面默认**：标签定义了 `default=100`  使用 `100` (智能类型转换)。
4. **零值**：以上都不满足  保持 Go 零值。

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
如果配置错误导致循环引用（如 `A default=@B`, `B default=@A`），Mapper 会自动中断引用链并返回 `nil`，防止程序崩溃。
2. **Nil 输入处理**：
`mapper.Decode(nil, &out)` 是合法的。此时结构体字段将全部尝试应用 `default` 值。

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

## 6. 注意事项

1. **Default 语法限制**：`default` 标签的值中目前**不能包含逗号**（`,`），因为逗号是 struct tag 的分隔符。
2. **结构体指针**：`output` 参数必须是指针，且必须是可寻址的。