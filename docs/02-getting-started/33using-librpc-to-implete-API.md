---
slug: using-librpc-to-implete-API
title: 快速开发：使用librpc
type:  docs
sidebar_position: 33
---

##  librpc  设计原则
librpc 包，位于：github.com/doptime/doptime/librpc
- 用极小，0.1M，的代码体积，覆盖尽可能多的重要后端功能。
- librpc长期目标是覆盖大多数的后端功能。
- 不需要使用第三方库，以非常简单的方式使用。
- 成本计价原则。努力使得RPC的成本低于本机运行成本。

- 欢迎加入社群，提交接口需求。
  <!-- - 模块进化原则。doptime 反复迭代优化一千种RPC来满足大部分RPC需求。 -->
<!-- - 零配置，透明的API交易。doptime 像股市一样，把API调用价格自动成交在市场供需平衡点。 -->
<!-- RPC的逻辑代码，涉及不同的语言，不同的依赖，编译配置等。把它们放到一个项目实际上是不可行的。 -->

## 拆解RPC的调用流程
1. 根据上下文数据，预备RPC入参。这包括来自web client + server db中的数据。
2. 调用RPC，完成API的逻辑操作，并得到结果。
3. 存储相关结果到数据库。

4. 返回数据给客户端。

## 实现流程拆解的3个Hook

### 1. HookParamEnhancer
使用 HookParamEnhancer ，完成 .1 功能 来补足缺失的入参（比如从数据库中读取数据等）。
```go
func (ctx *Context[i, o]) HookParamEnhancer(
	paramEnhancer func(param i) (out i, err error),
) *Context[i, o] {
	ctx.ParamEnhancer = paramEnhancer
	return ctx
}
```
   paramEnhancer 是一个修正 param i的函数。

   很多时候，入参的param已经初步准备好了，但是可能不完整，部分field可能需要查询数据库或是计算才能得到。这个hook用以修正。  

   
### 2. HookResultSaver
使用HookResultFinalizer， 完成 .3(存储RPC结果到数据库)。    
```go 
func (ctx *Context[i, o]) HookResultSaver(
	resultSaver func(param i, ret o) (err error),
) *Context[i, o] {
	ctx.ResultSaver = resultSaver
	return ctx
}
```
   resultSaver 是一个保存RPC结果到数据库的函数。 你可以在这里保存RPC的结果到数据库中。 通常是redis数据库。  
### 3. HookResponseModifier
使用 HookResponseModifier ，完成 .4 (修改给web client的响应数据).

```go 
// HookResponseModifier is a hook function to modify the response  value to the web client.
func (ctx *Context[i, o]) HookResponseModifier(
	ResponseModifier func(param i, ret o) (valueToWebclient interface{}, err error),
) *Context[i, o] {
	ctx.ResponseModifier = ResponseModifier
	return ctx
}
```
   需要说明的是 修改给客户端的响应数据 通常是不需要的，因为rpc 中通常已经考虑了 json tag ,来避免返回数据中存在敏感数据的泄露。如果你需要返回空值等，可以在这里修改。

### 使用Other字段
doptime librpc 中的 入参，都存在 Other map[string]interface{} 的字段，用来存放web-client的其它请求参数。  

那些并没有在入参中出现的参数，如果你又需要使用，可以从这个map中取出来。  

如果多个hook函数之间要通信，可以通过 Other 字段 来传递。

## 调用第三方RPC函数示例
```go   title="main.go"
package main

import (
	"github.com/doptime/doptime/api"
	. "github.com/doptime/doptime/librpc"
	"github.com/doptime/doptime/data"
)


//use default: api.Option.ApiSourceHttp("doptime")
keyText2Audio:=data.New("keyText2Audio")
var Text2Audio = api.RpcOverHttp[*Text2AudioIn, *Text2AudioOut]().
HookParamEnhancer(func( req *Text2AudioIn) (r *InDemo, e error) {
	req.Other["TraceId"] = "123456"
	//load data from redis to complete req if you need
	return req, nil
}).HookResultSaver(func ( req *InDemo,rlt *Text2AudioOut ) (e error) {
	keyText2Audio.HSet(xxhash.Sum64String(req.Text),rlt)
	//save result here if you need 	
	return nil
}).HookResponseModifier(func (req *InDemo,rlt *Text2AudioOut ) (ret interface{},e error) {
	//furthur modify the response value to the client here
	return rlt, nil
}).Func


```


### 如何查看有哪些第三方函数可用。
1. 通常，你引用 doptime/librpc 包, 在vscode 中输入 你的关键, 来查看可以创建的RPC函数。
```go 
package main

import (
	. "github.com/doptime/doptime/librpc"
)
api.RpcOverHttp[*yourKeywordsHere...
```

2. 参加社群。提交你的请求。
 
