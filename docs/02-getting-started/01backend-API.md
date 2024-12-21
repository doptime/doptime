---
slug: API
title: 后端 - 使用API
type: docs
sidebar_position: 1
---
instruction: doptime 项目中的api 接口参数已经发生了变化，现在需要根据相关的文档重新定义api接口文档。

:::tip 申明api
## **通过api.Api定义doptime函数**
:::
```go title="/api/api.go"
package api

// api.Api 定义如下:
func Api[i any, o any](
	f func(InParameter i) (ret o, err error),
	options ...*ApiOption,
) (out *Context[i, o]) {
    ...
}
```

`api.Api(...)` 要求传入一个函数。
	• 函数的入参是一个结构体指针；
	• 函数的出参有两个：1）任意的数据类型 2）error。  
`api.Api(...).Func` 函数Func，就是传入函数。

这是定义 api.Api 的例子：

```go   title="main.go"
package main

import (
	"github.com/doptime/doptime/api"
	"github.com/doptime/doptime/data"
)

type InDemo struct {
	Id   string `mapstructure:"JwtId"`
}

var keyInDemo = data.New[string, *InDemo]()
// define api with input/output data structure
ApiDemo := api.Api(func(req *InDemo) (ret string, err error) {
    keyInDemo.HSET(req.Id, req)
    return `{data:"ok"}`, nil
}).Func
```


### API 命名规则

	1.	默认的 API 名称是 “api:” + 结构体类型名称。且类型名首字母会改写为小写。
	2.	结构体名称会移除常见的前缀和后缀。  
移除前缀： “api:”, “input”, “in”, “req”, “arg”, “param”, “src”, “data”  
移除后缀： “input”, “in”, “req”, “arg”, “param”, “src”, “data”  

举例来说，上述例子中，ApiDemo 的 API 名称是 “api:demo”。它对入参 InDemo 附加了一个前缀 “api:”，移除了 “InDemo” 中的 “In”，并且首字母小写。

### API 的调用方式

1.	作为普通的函数调用
2.	通过 HTTP 调用（通过 doptime-client，在 web 客户端调用）
3.	通过 RPC 调用
4.	通过 RPCOverHttp 调用（在服务端）

### API 被 RPC 调用出现异常的诊断：

1.	在 Redis 中，查看 RPC 调用传参  
•	当发起 RPC 调用后，可以在 Redis 下 API 名称（如 api:demo）的主键下，看到传参列表  
•	注意：如果是本地参数调用，不会有传输参数。  
2.	使用 httpserve.Debug()，在源码中的 httpStart() 函数下断点。查看传参处理过程。  

## api.Api 的可选参数

### api.Option.WithSourceRds(ApiSourceRds string)

例如在微服务的拆分隔离中，通过配置选项，你可以指定其它的 Redis datasource 作为参数交换网关。


```go   title="main.go"
// define api with input/output data structure
ApiDemo := api.Api(func(req *InDemo) (ret string, err error) {
    keyInDemo.HSET(req.Id, req)
    return `{data:"ok"}`, nil
}, api.Option.WithSourceRds("ServiceQ")).Func
```

上面的例子中，api.Option.WithSourceRds("ServiceQ") 是可选参数，用来指定 Redis 服务器中转函数调用。


### api.Option.WithSourceHttp(ApiSourceHttp string)
创建基于 HTTP 的 API 调用。
通过ApiSourceHttp string参数，指定 HTTP 服务器的地址。
    
```go   title="main.go"
// define api with input/output data structure
ApiDemo := api.Api(func(req *InDemo) (ret string, err error) {
    keyInDemo.HSET(req.Id, req)
    return `{data:"ok"}`, nil
}, api.Option.WithSourceHttp("http://localhost:8080")).Func
```