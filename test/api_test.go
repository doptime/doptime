package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/doptime/doptime/api"
	_ "github.com/doptime/doptime/httpserve"
	"github.com/doptime/doptime/rpc"
)

type Demo struct {
	Text string
}
type Demo1 struct {
	Text       string `json:"Text" validate:"required"`
	Attach     *Demo
	RemoteAddr string `json:"HeaderRemoteAddr" validate:"required"`
}

var ApiDemo = api.Api(func(InParam *Demo1) (ret string, err error) {
	now := time.Now()
	fmt.Println("Demo api is called with InParam:" + InParam.Text + " run at " + now.String() + " Attach:" + InParam.Attach.Text + " UserIP:" + InParam.RemoteAddr)
	return "hello world", nil
}).Func

func TestApiDemo(t *testing.T) {

	//create a http context
	var (
		result string
		err    error
	)
	result, err = ApiDemo(&Demo1{Text: "hello"})
	if err != nil {
		t.Error(err)
	} else if result != "hello world" {
		t.Error("result is not hello world")
	}
}

var DemoRpc = rpc.Rpc[*Demo1, string]().Func

func TestRPC(t *testing.T) {

	//create a http context
	var result string
	var err error
	result, err = DemoRpc(&Demo1{Text: "hello"})
	if err != nil {
		time.Sleep(10 * time.Second)
		t.Error(err)
	} else if result != "hello world" {
		t.Error("result is not hello world")
	}

}

var callDemoAt = rpc.CallAt(DemoRpc)

func TestCallAt(t *testing.T) {
	var (
		err   error
		now   time.Time = time.Now()
		param           = &Demo1{Text: "TestCallAt 10s later", Attach: &Demo{Text: "Attach"}}
	)

	fmt.Println("Demo api is calling with InParam:" + param.Text + " run at " + now.String())

	if err = callDemoAt(now.Add(time.Second*10), param); err != nil {
		t.Error(err)
	}
	time.Sleep(15 * time.Second)
}
func TestCallAtCancel(t *testing.T) {
	var (
		err   error
		now   time.Time = time.Now()
		param           = &Demo1{Text: "TestCallAt 10s later"}
	)
	fmt.Println("Demo api is calling with InParam:" + param.Text + " run at " + now.String())
	timeToRun := now.Add(time.Second * 10)
	if err = callDemoAt(timeToRun, param); err != nil {
		t.Error(err)
	}
	if err = rpc.CallAtCancel(callDemoAt, timeToRun); err != nil {
		t.Error("cancel failed")
	}
	time.Sleep(30 * time.Second)

}
