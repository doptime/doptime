---
slug: CallAt
title: 后端 - 定时调用CallAt
type:  docs
sidebar_position: 4
---


## CallAt: 定时运行你的代码
### 如何使用
doptime 提供了一个定时调用函数，以便在未来的某个时间运行你的代码。  
1. 你需要先定义好一个api, 或者是rpc函数。  
2. 然后通过CallAt来创建对该函数的定时调用函数。  
3. 像普通函数一样调用创建的函数。
```go   title="main.go"

package main

import (
	"github.com/doptime/doptime/api"
	"time"
	"fmt"
)

type InDemo struct {
	Text   string 
}
//create Api before callAt this api
var ApiDemo=api.Api(func(req *InDemo) (ret InDemo, err error) {
	fmt.Println("Demo api is called with InParam:" + req.Text + " run at " + time.Now().String())
    return  req, nil
})
//create Rpc before callAt this rpc
var DemoRpc = api.Rpc[*InDemo, string]()
var	callAtDemoApi := api.CallAt(ApiDemo)
var	callAtDemoRpc = api.CallAt(DemoRpc)
func main() {
	var param = &InDemo{Text: "TestCallAt "+time.Now().String()}	
	
	callAtDemoApi( time.Now().Add(time.Second*4),param)

	callAtDemoRpc(time.Now().Add(time.Second*2),param)
	//wait for the callAt to finish
	time.Sleep(time.Second*5)
}
```


### CallAt 补充说明
- CallAt的入参函数是要求是 api.Rpc 或 api.Api 创建的函数。 
- 这个调用后会立即返回。但是会在指定的时间后执行目标RPC或者是API。  
- 这个定时调用函数不受程序重启影响。    
    

## CallAtCancel: 取消已经计划好的定时调用
```go   title="main.go"

package main

import (
	"github.com/doptime/doptime/api"
	"time"
	"fmt"
)

type InDemo struct {
	Text   string 
}
//create Api before callAt this api
var ApiDemo=api.Api(func(req *InDemo) (ret InDemo, err error) {
	fmt.Println("Demo api is called with InParam:" + req.Text + " run at " + time.Now().String())
    return  req, nil
})
//create Rpc before callAt this rpc
var DemoRpc = api.Rpc[*InDemo, string]()
var	callAtDemoApi := api.CallAt(ApiDemo)
func main() {
	var param = &InDemo{Text: "TestCallAt "+time.Now().String()}	
	
	timeToRun:=time.Now().Add(time.Second*4)
	callAtDemoApi(timeToRun ,param)
	api.CallAtCancel(callAtDemoApi，timeToRun)
	time.Sleep(time.Second*5)
}
```
### CallAtCancel 说明
- 第一个参数是api.CallAt(...) 创建的e函数。
- 第二个参数是你想要取消的时间。
  时间是唯一的凭据，如果你要取消，就要先记住这个时间。