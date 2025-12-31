package api

import (
	"context"

	cmap "github.com/orcaman/concurrent-map/v2"
)

var ApiServiceBatchSize = cmap.New[int64]()

// ApiOption is parameter to create an API, RPC, or CallAt
type ApiCtx[i any, o any] struct {
	Name         string
	ApiSourceRds string
	Ctx          context.Context
	Func         func(InParameter i) (ret o, err error)
	Validate     func(pIn interface{}) error
	// you can rewrite input parameter before excecute the service
	ParamEnhancer func(param i) (out i, err error)

	// you can save the result to db using paramMap
	ResultSaver func(param i, ret o) (err error)

	// you can modify the result value to the web client.
	ResponseModifier func(param i, ret o) (valueToWebclient interface{}, err error)
}
