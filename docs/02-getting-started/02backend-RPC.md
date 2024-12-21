---
slug: RPC
title: 后端 - 使用RPC
type:  docs
sidebar_position: 2
---


## **通过api.Rpc定义Rpc函数**
 
在上一节中，我们定义了一个api，现在我们将定义一个RPC函数来使用这个api。
要创建函数，你需要使用api.Rpc函数使用。  
```go   title="main.go"
package main

import (
	"github.com/doptime/doptime/api"
)

type InDemo struct {
	Id   string `mapstructure:"JwtId" validate:"required"`
}

//var DemoRpc = api.Rpc[*InDemo, string](options ...*api.ApiOption)
var DemoRpc = api.Rpc[*InDemo, string](...).Func
result, err := DemoRpc(&InDemo{Id:"123"})

```
- 如果你不需要用到多体模式。则不需要使用RPC 功能。  
  
## RPC 可选参数

与 API 相同，RPC 也支持可选参数来扩展功能。例如，可以指定使用不同的数据源或添加钩子函数。

示例：使用 WithSourceRds 参数
```go   title="main.go"
var DemoRpc = api.Rpc[*InDemo, string](
	api.Option.WithSourceRds("CustomRds"),
).Func
```

示例：添加 Hook 函数
```go   title="main.go"
var DemoRpc = api.Rpc[*InDemo, string](
	api.Option.WithSourceRds("CustomRds"),
).HookParamEnhancer(func(req *InDemo) (*InDemo, error) {
	// 参数增强逻辑
	return req, nil
}).Func
```
这些钩子函数可以在参数处理、结果保存、响应修改等方面提供额外的控制。


示例：完整的 RPC 函数定义和调用
```go   title="main.go"
package main

import (
	"github.com/doptime/doptime/api"
	"fmt"
)

type InDemo struct {
	Id   string `mapstructure:"JwtId" validate:"required"`
}

var DemoRpc = api.Rpc[*InDemo, string](
	api.Option.WithSourceRds("CustomRds"),
).HookParamEnhancer(func(req *InDemo) (*InDemo, error) {
	// 参数增强逻辑
	return req, nil
}).HookResultSaver(func(req *InDemo, ret string) error {
	// 结果保存逻辑
	return nil
}).HookResponseModifier(func(req *InDemo, ret string) (interface{}, error) {
	// 响应修改逻辑
	return ret, nil
}).Func

func main() {
	result, err := DemoRpc(&InDemo{Id: "123"})
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Result:", result)
	}
}
```
通过以上示例，我们可以清晰地看到如何定义和使用RPC函数，包括参数增强、结果保存和响应修改等功能。

