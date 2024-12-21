---
slug: Data
title: 后端 - 使用Data
type:  docs
sidebar_position: 3
---

:::tip dopTime的数据对象
## doptime可仅作为数据库框架使用。
::: 
 你可以独立地使用doptime的数据操作功能。就像它只是一个数据库操作框架（如同GORM那样）。

:::tip 使用doptime定义数据对象:
## **定义doptime的数据对象**
::: 
### 使用data.New\[keyType,valueType\](...)
```go   title="main.go"
package main

import (
 "github.com/doptime/doptime/data"
)

type Demo struct {    
    Foo   string `msgpack:"alias:foo"`
    Name  string `msgpack:"alias:name"`
}
// 如果你需要自定义主键名称，或redis datasource, 请参看可选参数
var keyDemo = data.New[string,*Demo]()
keyDemo.HSet("foo", &Demo{Foo:"bar"})
if _demo, err := keyDemo.HGet("foo"); err == nil {
    fmt.Println("demo:", _demo.Foo)
}
```
在上面这个例子中，我们定义了一个redis data context,对应 主键名称是demo.   
我们在其中添加一个field名为"foo", value 值为Demo\{Foo:"bar"\}的数据。


:::tip data.New的可选参数
## **通过data.New的可选参数**   
:::
```go   title="main.go"
var keyDemo =data.New[string,*Demo](data.Option.WithRds("dragonflydb").WithKey("myKeyName"))
```
### data.Option.WithRds(...)
使用WithRds，data context可以指定其它的redis datasource。 如果不使用，作为默认，redis datasource值是"default" .   
可以通过这样的方式来创建迁移数据的新源。
### data.Option.WithKey(...)
主键名称默认是根据 valueType 的 struct name 来生成的。
你可以通过WithKey来指定其它的主键名称。
许多时候，valueType 可能是string 等类型。这时候你需要指定主键名称。因为在doptime 中系统默认类型是不被允许作为主键名称的。
### 可选参数可以单独或组合使用
&nbsp;
:::tip data中，如何定义数据序列化行为
## **通过msgpack定义data的序列化**   
::: 
```go   title="main.go"
type Demo struct {    
    Foo   string `msgpack:"alias:foo"`
    Name  string `msgpack:"alias:name"`
}
```
### **说明**
- doptime 使用 **msgpack**  来序列化data对象的值   
- 通过msgpack 标签，你可以定义数据的序列化方式。比如不保存某些字段等。  
- [点击查看更多msgpack文档](https://msgpack.uptrace.dev/guide/#quickstart:~:text=%23-,Struct%20tags,-msgpack%20supports%20following) 。  
