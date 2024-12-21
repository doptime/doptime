---
slug: replacing-CRUD-APIs-with-redis-commands
title: 快速开发：安息吧 CURD API
type:  docs
sidebar_position: 30
---
### doptime 如何让CURD操作消失。
- CURD API的作用可以抽象为3点：1.实现上下文绑定，2.实现权限控制. 3.实现数据操作。

- doptime 已经实现了2: dopTime已经通过黑白名单`command::key::on/off` 解决了权限问题。  
- doptime 已经实现了3: doptime-client 可以直接操作redis数据。  
- 所以，最后我们只需要解决上下文绑定问题，就可以让CURD操作消失。
 
## 绑定上下文和数据操作
通过以下两种方式，CURD API可以被统一为几个标准的redis函数，不需要另写 
### 绑定JWT上下文和数据操作
- 我们在doptime-client 的 **主键 / key**和**字段 / fields**  中引入绑定标识符约定：@xxx。 比如主键名：userInfo@id。
- redis key, redis fields 中的 @标记，会被doptime 自动替换成jwt中相应的字段。   
  比如 @id 会被替换成jwt中的id字段。  像是userInfo@id被替换为: userInfo2802151341134135268

- 通过这种方式，我们实现了JWT上下文绑定。 可以使用上下文相关的redis CURD。
### 绑定其它的上下文 
- 对于上下文信息不在JWT中的情况，可以在客户端JS中通过查询等方式得到key fields等信息，然后再进行CURD操作。  

- 一个非常常见的做法是是计算哈希值并查询。比如文本转语音。可以计算文本的xxhash值，然后hget("text2voice",xxhash(text))来查询。


### 注意事项
- 建议JWT 的id 等主键哈希，都使用string类型, 而不采用int32或int64类型。  
- 因为jwt字段不区分类型，所以int会被理解为float64。@标记 会被表示成科学计数法的情况。引发查询失败和非预期键写入。

- string 类型可以避免类型转换开销。
