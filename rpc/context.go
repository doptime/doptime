package rpc

import (
	"context"

	"github.com/doptime/config/cfgapi"
)

// ApiOption is parameter to create an API, RPC, or CallAt
type Context[i any, o any] struct {
	Name          string
	ApiSourceRds  string
	ApiSourceHttp *cfgapi.ApiSourceHttp
	Ctx           context.Context
	Func          func(InParameter i) (ret o, err error)
	Validate      func(pIn interface{}) error
	// you can rewrite input parameter before excecute the service
	ParamEnhancer func(param i) (out i, err error)

	// you can save the result to db using paramMap
	ResultSaver func(param i, ret o) (err error)

	// you can modify the result value to the web client.
	ResponseModifier func(param i, ret o) (valueToWebclient interface{}, err error)
}
