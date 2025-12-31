# Mapper v2: 工业级 Go 结构体解码库

**Mapper** 是一个轻量级但功能强大的 Go 语言库，用于将非结构化数据（如 `map[string]interface{}`、JSON 解析结果）解码为强类型的 Go 结构体。

它的设计理念是 **"JSON + Mapstructure + Remain + Smart Defaults"**，完美解决了“固定字段 + 动态扩展”以及“配置合并”等复杂场景下的痛点。

> ⚠️ **v2 版本重要变更**：
> 1.  **语法升级**：全面采用 `@` 定义默认值，移除 `default=` 写法。
> 2.  **分隔符变更**：强制使用**空格**分隔字段名与指令（如 `json:"name @val"`），**不再支持逗号分隔**。

## 1. 核心特性

* **零配置使用**：通过 `mapper.Decode` 一键直达。
* **极简语法 (Unified @ Syntax)**：
    * **字面量默认**：支持 `json:"port @8080"`，自动推断并转换为目标类型（优先解析为 int/float，最后尝试 bool）。
    * **字段引用**：支持 `json:"host @ip"`，若当前字段缺失，自动引用同级字段 `ip` 的值。
    * **纯空格分隔**：`json:"name @default"`，符合 Go Struct Tag 惯例，通过 IDE 语法检查。
* **Remain 机制**：
    * 自动将未被定义的字段捕获到 `map` 中。
    * **保留原始 Key 大小写**：输入是什么 Key，Remain 中就是什么 Key。
* **弱类型转换 (Weakly Typed)**：
    * `string` ↔ `int/float/bool` 自动互转。
    * `bool` → `int/uint/float` (true=1, false=0)。
    * `string` → `time.Time` 自动解析（支持 RFC3339 等多种格式）。
    * 支持 `json.Number` 的无损处理。
* **工业级健壮性**：
    * **循环引用检测**：防止配置错误（A 引用 B，B 引用 A）导致栈溢出。
    * **Nil 安全**：输入为 `nil` 时依然能正确应用默认值。

## 2. 快速开始

### 2.1 基础映射与智能默认值

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

	// 2. 字面量默认值 (空格分隔)
	// 语法：字段名 "port", 默认值 "@8080" (解析为 int)
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

### 2.2 Remain 机制（处理动态字段）

当输入包含结构体未定义的字段时，`Remain` 字段会自动“兜底”。

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

## 3. 核心逻辑详解

### 3.1 值决议优先级 (Resolution Logic)

对于任何字段，Mapper 按照以下优先级决定最终值：

1. **输入优先**：输入数据中存在该字段（忽略大小写匹配） → 使用输入值。
2. **引用默认 (`@ref`)**：标签定义了 `@other` 且 `other` 存在于输入中 → 使用 `other` 的值。
3. **字面默认 (`@val`)**：标签定义了 `@100` 或 `@true` → 使用解析后的字面量值。
* *解析顺序*：Int -> Float -> Bool -> String (避免 "1" 被误判为 true)。


4. **零值**：以上都不满足 → 保持 Go 零值。

> **注意**：如果通过 `@ref` 引用了某个字段，该被引用字段也会被标记为“已使用”，**不会**再出现在 `Remain` 中（避免数据重复）。

### 3.2 弱类型支持矩阵

`mapper` 会尽最大努力尝试转换类型，而不是直接报错。

| 输入类型 (Input) | 目标类型 (Struct Field) | 行为 |
| --- | --- | --- |
| `string` ("123") | `int` / `uint` / `float` | 自动解析数值 |
| `string` ("true") | `bool` | 转为 `true` |
| **`bool` (true)** | **`int` / `float**` | **转为 `1` / `1.0**` (v2 新特性) |
| `json.Number` | `int/float/string` | 智能拆箱，保留精度 |
| `string` (RFC3339) | `time.Time` | 自动尝试多种格式解析 |

## 4. API 参考

### `func Decode(input interface{}, output interface{}) error`

最常用的入口函数。

* `input`: 输入数据，通常是 `map[string]interface{}`，也可以是 `nil`。
* `output`: 必须是结构体指针（`*struct`）。

### `type DecoderConfig`

如果需要更精细的控制，可以使用 `NewDecoder`。

* `TagName`: 默认为 "json"。
* `WeaklyTypedInput`: 默认为 `true`。

---

## 5. v1 到 v2 迁移指南

v2 版本引入了破坏性变更（Breaking Changes）。

**1. 语法替换**

* ❌ 旧 (v1): `json:"port, default=8080"`
* ✅ 新 (v2): `json:"port @8080"`
* ❌ 旧 (v1): `json:"host, default=@ip"`
* ✅ 新 (v2): `json:"host @ip"`

**2. 分隔符变更**

* v2 **严格使用空格**进行分词。
* **禁止使用逗号**分隔字段名和指令。
* ❌ 错误：`json:"id, @val"` （会被解析为字段名是 `"id,"`，导致匹配失败）
* ✅ 正确：`json:"id @val"`