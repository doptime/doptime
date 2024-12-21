---
slug: RpcOverHttp
title: 后端 - 使用RpcOverHttp
type:  docs
sidebar_position: 2
---


:::tip 申明api
## **通过api.RpcOverHttp定义基于HTTP的RPC函数**
::: 
 和api.RPC 一样，api.RpcOverHttp是通过RPC 来调用远程的API，但不同的是通过http 连接，而不是Redis连接。  
```go   title="main.go"
package main

import (
	"github.com/doptime/doptime/api"
)

type InDemo struct {
	Id   string `mapstructure:"JwtId" validate:"required"`
	Other      map[string]interface{} `mapstructure:",remain" msgpack:"-" `
}
//RpcOverHttp 默认使用 doptime 作为目标服务器
var DemoRpc = api.RpcOverHttp[*InDemo, string]().Func
result, err := DemoRpc(&InDemo{Id:"123"})



var DemoRpc1 = api.RpcOverHttp[*InDemo, string](api.Option.ApiSourceHttp("HttpRPCName")).
HookParamEnhancer(func( req *InDemo) (r *InDemo, e error) {
	//load data from redis to complete req if you need
	return req, nil
}).HookResultSaver(func ( req *InDemo,ret string ) (e error) {
	//save result here if you need 	
	return nil
}).HookResponseModifier(func (req *InDemo,ret string ) (retvalue interface{},e error) {
	//furthur modify the response value to the client here
	return ret, nil
}).Func


```
- RpcOverHttp 有更高的响应速度。
- api.RpcOverHttp 用于连接使用 doptime 的api 服务 和 你自己的RPC服务。
- RpcOverHttp 设计用来避免Redis Server 向不被信任的终端暴露
- RpcOverHttp 有3个可选的Hook函数，用来增强参数，保存结果，修改返回值。
- 可以使用HookParamEnhancer 进一步补充参数
- 可以使用HookResultSaver 保存结果
- 可以使用HookResponseModifier 进一步修改给客户端的返回值，比如总是返回空值。