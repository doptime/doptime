# Mapper v2: 工业级 Go 结构体解码库

**Mapper** 是一个轻量级但功能强大的 Go 语言库，用于将非结构化数据（如 `map[string]interface{}`、JSON 解析结果）解码为强类型的 Go 结构体。

它的设计理念是 **"JSON + Mapstructure + Remain + Defaults"**，完美解决了“固定字段 + 动态扩展”以及“配置合并”等复杂场景下的痛点。

> ⚠️ **v2 版本重要变更**：
> 1. **语法升级**：全面采用 `@` 定义默认值，移除 `default=` 写法。
> 2. **分隔符变更**：强制使用**空格**分隔字段名与指令（如 `json:"name @val"`），**不再支持逗号分隔**。
> 
> 

## 1. 核心特性

* **隐式映射 (Implicit Mapping)**：
* **零标签**：无需编写 `json:"name"`，自动根据字段名（忽略大小写）匹配输入数据。
* **零配置**：开箱即用，通过 `mapper.Decode` 一键直达。


* **统一 `@` 语法 (Unified @ Syntax)**：
* **字面默认**：`json:"port @8080"`，字段缺失时自动使用 8080。
* **引用默认**：`json:"host @ip"`，字段缺失时引用同级 `ip` 字段的值。
* **特殊键引用**：`json:"uid @@sub"`，引用 Map 中名为 `@sub` 的特殊键（常用于框架层注入）。


* **Remain 机制**：
* 自动将结构体中未定义的字段捕获到 `Remain` map 中。
* **保留原始 Key**：输入是什么 Key，Remain 中就是什么 Key（区分大小写）。


* **弱类型转换 (Weakly Typed)**：
* `string` ↔ `int/float/bool` 自动互转。
* `bool` → `int/uint/float` (true=1, false=0)。
* `string` → `time.Time` 自动解析（支持 RFC3339 等多种格式）。
* 支持 `json.Number` 的无损处理。


* **工业级健壮性**：
* **循环引用检测**：防止配置错误（A 引用 B，B 引用 A）导致栈溢出。
* **Nil 安全**：输入为 `nil` 时依然能正确应用默认值。



## 2. 快速开始

### 2.1 基础映射与默认值

绝大多数情况下，你不需要写 Tag。只有在需要**重命名**或**指定默认值**时才使用 Tag。

```go
package main

import (
	"fmt"
	"your_project/pkg/mapper"
)

type Config struct {
	// 1. 隐式映射 (无需 Tag)
	// 自动匹配输入 Map 中的 "Name", "name", "NAME" 等
	Name string

	// 2. 字面量默认值
	// 字段名自动推断为 "Port"；若输入缺失，默认 8080
	Port int `json:"@8080"`

	// 3. 引用默认值
	// 字段名自动推断为 "Host"；若输入缺失，引用 "ip" 字段的值
	Host string `json:"@ip"`

	// 4. 显式重命名 + 默认值
	// 强制绑定 Map 中的 "is_debug"；若缺失，默认 true
	Debug bool `json:"is_debug @true"`
}

func main() {
	input := map[string]interface{}{
		"name": "MyApp",
		"ip":   "192.168.1.1",
		// "Port" 缺失 -> 使用 8080
		// "Host" 缺失 -> 引用 ip
		// "is_debug" 缺失 -> 使用 true
	}

	var cfg Config
	if err := mapper.Decode(input, &cfg); err != nil {
		panic(err)
	}

	fmt.Printf("Config: %+v\n", cfg)
	// Output: Config: {Name:MyApp Port:8080 Host:192.168.1.1 Debug:true}
}

```

### 2.2 进阶模式：引用特殊键名 (`@@`)

配合上层框架（如 Doptime），Mapper 可以读取 Map 中以 `@` 开头的特殊键。

* **场景**：框架在 HTTP 层将 JWT Claims (如 `sub`) 注入到 Map 中，Key 命名为 `@sub`。
* **语法**：`json:"@@sub"` 实际上是 `json:"@sub"` 的引用简写（即查找名为 `@sub` 的 Key）。

```go
type UserProfile struct {
    // 自动映射 Map 中的 "@sub" 键
    // 实现了无需手动提取 Context 的身份绑定
    UserID string `json:"@@sub"` 

    Name string 
}

// 模拟框架构造的输入数据
input := map[string]interface{}{
    "name": "Alice",
    "@sub": "user_123456", // 框架注入的安全字段
}

```

### 2.3 Remain 机制（动态字段兜底）

```go
type Product struct {
	ID   int    `json:"id"`
	Name string // 隐式映射
	// 捕获所有未被 ID 和 Name 消耗的字段
	Remain map[string]interface{} `json:"-"`
}

func main() {
	input := map[string]interface{}{
		"id":     100,
		"name":   "Phone",
		"Color":  "Red",  // 动态字段
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

对于任何结构体字段，Mapper 按照以下优先级决定最终值：

1. **输入优先**：输入 Map 中存在该字段（忽略大小写） → 使用输入值。
2. **引用默认**：
* 标签定义了 `@Key` (例如 `@ip` 或特殊键 `@@sub`)。
* Map 中存在该 `Key` (区分大小写) → 使用该 Key 的值。


3. **字面默认**：
* 标签定义了 `@Value` (例如 `@8080`) → 解析字面量。
* *解析顺序*：Int -> Float -> Bool -> String。


4. **零值**：以上都不满足 → 保持 Go 零值。

> **注意**：被引用的字段（如 `@ip`）会被标记为“已使用”，**不会**再出现在 `Remain` 中。

### 3.2 弱类型支持矩阵

`mapper` 尽最大努力进行类型转换，而不是直接报错。

| 输入类型 (Input) | 目标类型 (Struct Field) | 行为 |
| --- | --- | --- |
| `string` ("123") | `int` / `uint` / `float` | 自动解析数值 |
| `string` ("true") | `bool` | 转为 `true` |
| **`bool` (true)** | **`int` / `float**` | **转为 `1` / `1.0**` (v2 新特性) |
| `json.Number` | `int/float/string` | 智能拆箱，保留精度 |
| `string` (Time) | `time.Time` | 自动尝试 RFC3339 等格式 |

## 4. API 参考

### `func Decode(input interface{}, output interface{}) error`

最常用的入口函数。

* `input`: 输入数据，通常是 `map[string]interface{}`，也可以是 `nil`。
* `output`: 必须是结构体指针（`*struct`）。

---

## 5. v1 到 v2 迁移指南

v2 版本引入了破坏性变更（Breaking Changes）。

**1. 语法替换**

* ❌ 旧 (v1): `json:"port, default=8080"`
* ✅ 新 (v2): `json:"port @8080"` (或仅 `json:"@8080"`)

**2. 分隔符变更**

* v2 **严格使用空格**进行分词。
* **禁止使用逗号**分隔字段名和指令。
* ❌ 错误：`json:"id, @val"`
* ✅ 正确：`json:"id @val"`