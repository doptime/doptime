package api

import (
	"context"

	"github.com/doptime/doptime/config"
)

// ApiOption is parameter to create an API, RPC, or CallAt
type Context[i any, o any] struct {
	Name          string
	ApiSourceRds  string
	ApiSourceHttp *config.ApiSourceHttp
	WithHeader    bool
	Ctx           context.Context
	F             func(InParameter i) (ret o, err error)
	Validate      func(pIn interface{}) error
	// you can rewrite input parameter before excecute the service
	ParamEnhancer func(_mp map[string]interface{}, param i) (out i, err error)

	// you can save the result to db using paramMap
	ResultSaver func(param i, ret o, paramMap map[string]interface{}) (err error)

	// you can modify the result value to the web client.
	ResponseModifier func(param i, ret o, paramMap map[string]interface{}) (valueToWebclient interface{}, err error)
}
