package api

import (
	"context"
	"errors"
	"reflect"

	"github.com/doptime/doptime/specification"
	"github.com/rs/zerolog/log"
)

// ApiOption is parameter to create an API, RPC, or CallAt
type Api[i any, o any] struct {
	Name       string
	DataSource string
	WithHeader bool
	IsRpc      bool
	Ctx        context.Context
	F          func(InParameter i) (ret o, err error)
	Validate   func(pIn interface{}) error
	// you can rewrite input parameter before excecute the service
	ParamEnhancer func(_mp map[string]interface{}, param i) (out i, err error)

	// you can save the result to db using paramMap
	ResultSaver func(param i, ret o, paramMap map[string]interface{}) (err error)

	// you can modify the result value to the web client.
	ResponseModifier func(param i, ret o, paramMap map[string]interface{}) (valueToWebclient interface{}, err error)
}

func New[i any, o any](f func(InParameter i) (ret o, err error), options ...*ApiOption) (out *Api[i, o]) {
	var option *ApiOption = mergeNewOptions(&ApiOption{DataSource: "default", Name: specification.ApiNameByType((*i)(nil))}, options...)

	out = &Api[i, o]{Name: option.Name, DataSource: option.DataSource, IsRpc: false, Ctx: context.Background(),
		WithHeader: HeaderFieldsUsed(reflect.TypeOf(new(i)).Elem()),
		Validate:   needValidate(reflect.TypeOf(new(i)).Elem()),
		F:          f,
	}

	if len(option.Name) == 0 {
		log.Debug().Msg("ApiNamed service created failed!")
		out.F = func(InParameter i) (ret o, err error) {
			err = errors.New("Api name is empty")
			log.Warn().Str("service misnamed", out.Name).Send()
			return ret, err
		}
	}

	ApiServices.Set(out.Name, out)

	funcPtr := reflect.ValueOf(f).Pointer()
	fun2Api.Set(funcPtr, out)

	apis, _ := APIGroupByDataSource.Get(out.DataSource)
	apis = append(apis, out.Name)
	APIGroupByDataSource.Set(out.DataSource, apis)

	log.Debug().Str("ApiNamed service created completed!", out.Name).Send()
	return out
}
