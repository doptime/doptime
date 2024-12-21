---
slug: using-libapi
title: 快速开发：使用libapi
type:  docs
sidebar_position: 32
---

##  libapi  功能
libapi 提供一些重要而常用的api. 比如：是对象时间锁，JWT加密，解密等

## 调用第三方RPC函数示例
###  JWT加密

```go   title="Jwt-encoding-demo.go"
package main

import (
	"github.com/doptime/doptime/config"
	"github.com/doptime/doptime/libapi"
)

// type JwtEncodingOut struct {
// 	Token string `json:"token"`
// }
//func ApiJwtSign(in *JwtEncodingIn) (out *JwtEncodingOut, err error) 
ret,err:=libapi.ApiJwtSign(&libapi.JwtEncodingIn{
	Params:     map[string]interface{}{"id": "15303344054097909085", "name": "hpovzbd" },
	JwtSecret:  config.Cfg.Http.JwtSecret,
	SignMethod: libapi.HS256,
	Duration:   3600*24*30,
})

```

### key-时间锁定
```go   title="time-locker-demo.go"
package main

import (
	"github.com/doptime/doptime/libapi"
)

lockedBool,_:=libapi.ApiLockKey(&libapi.InLockKey{
	Key:     "user15303344054097909085",
	DurationMs:   3600*24*30,
})

if lockedBool {
	console.log("only 10 free trial allowed in one day!")
}
```


### [查看有哪些第三方函数可用](https://github.com/doptime/doptime/tree/master/libapi)

## [欢迎提交libapi](https://github.com/doptime/doptime/pulls)
 
