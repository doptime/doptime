---
slug: Http
title: 后端 - http
type:  docs
sidebar_position: 5
---
:::tip 使用doptime框架 :
## 启用HTTP服务
::: 
### 需要启用http服务情形：
1. 通过http, 进行redis 查询、修改、删除等操作。
2. 通过http, 调用api函数、RPC函数。
  
### 如何启用http服务：   
你可以通过匿名导入doptime/httpserve 模块来启用http模块的例子：
   
```go   title="main.go"

import (
	_ "github.com/doptime/doptime/httpserve"
)
// 现在，api.Api定义的函数自动启用http访问。你可以通过配置进一步了解启用和禁用的方法。
```
## 在前端中使用HTTP服务
	[这部分见doptime-client文档](./frontend-Usage)

##  HTTP API 注意事项

### 支持哪些数据格式的http请求 （Content-Type）
1. 支持json格式的数据。  
发情的请求header 的Content-Type 为"application/json" 时，要求body 采用 json 格式 编码.
2. 支持msgpack格式的数据。  
发情的请求header 的Content-Type 为"application/octet-stream" 时，要求body 采用 msgpack 格式 编码.
3. 默认采用msgpack编码发送的数据。


### doptime API 接受HTTP数据后的响应流程

1. 合并 form + body（ json body / msgpack body ） 成为一个大的 map[string]interface\{\} 
2. 继续合并 JWT 数据 + Header 数据 到该map中。  
3. 如果有 HookParamEnhancer，调用它，通过查询数据库等操作进一步补足参数。
4. 通过validator验证数据。
5. 调用API/RPC。 
6. 如果有 HookResultSaver，调用它，把结构存储到redis数据库中
7. 如果有 HookResponseModifier，调用它，进一步修改给客户端的返回值。

