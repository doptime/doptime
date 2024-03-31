package api

import (
	"context"
	"errors"
	"reflect"

	"github.com/doptime/doptime/specification"
	"github.com/rs/zerolog/log"
)

func New[i any, o any](f func(InParameter i) (ret o, err error), options ...*ApiOption) (out *Api[i, o]) {
	var option *ApiOption = mergeNewOptions(&ApiOption{DataSource: "default", Name: specification.ApiNameByType((*i)(nil))}, options...)

	out = &Api[i, o]{Name: option.Name, DataSource: option.DataSource, IsRpc: false, Ctx: context.Background(),
		WithHeader: HeaderFieldsUsed(reflect.TypeOf(new(i)).Elem()),
		WithJwt:    WithJwtFields(reflect.TypeOf(new(i)).Elem()),
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
