---
slug: doptime-flowchart
title: 后端框架： API + RPC + 数据
type:  docs
sidebar_position: 1
---
 
import Mermaid from '@theme/Mermaid';

:::tip 从http 到 doptime 到 redis
## doptime 后端服务交互流程图
:::
![doptime 后端服务交互流程图](image.png)


:::tip 如何理解
## 如何理解 doptime 
:::
### 1. [doptime是一个后端API框架](../getting-started/API)
- 在api.Api(...).Func中功能代码定义  
- 这个api可以响应http 请求。作为http服务端。(也可以不启用)

### 2. [doptime支持RPC](../getting-started/RPC)
- 上面创建的api函数作为RPC服务端用，响应RPC调用请求。不必另外定义。
- 通过在api.Rpc\[inType,outType\](...).Func 定义Rpc客户端函数，它可以调用RPC服务端。
- doptime通过RPC调用API时，使用redis 转发调用参数. 
- 我们把 redis 数据库看做是一个数据交换机，doptime进程则可以拔插到redis交换机上。
- 交换机模型大大简化了拓扑结构。插入和删除一个服务，不影响其他服务。这在动态添加服务，动态添加新服务上是非常有用的。 

### 3. [doptime 也支持传统点对点RPC](../getting-started/RpcOverHttp)
- doptime同样支持用Http直连而非redis转发，来提高效率
- 如果你不希望对外暴露redis服务器，可以使用http直连方案。
- 通过调用api.RpcOverHttp\[in,out\](Option.WithSourceHttp(...))，可以实现HTTP直连RPC调用

### 4. [作为数据库框架](../getting-started/Data)
- doptime 仅使用Redis 类型的数据库，不使用其它数据库
- doptime 自带了数据操作函数。你可以通过data.New\[keyType,valueType\]()来定义数据对象。


