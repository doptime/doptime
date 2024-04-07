package api

import (
	"context"
	"reflect"

	"github.com/doptime/doptime/specification"
	"github.com/doptime/doptime/vars"
	"github.com/rs/zerolog/log"
)

func Api[i any, o any](f func(InParameter i) (ret o, err error), options ...*ApiOption) (out *Context[i, o]) {
	var option *ApiOption = mergeNewOptions(&ApiOption{ApiSourceRds: "default", Name: specification.ApiNameByType((*i)(nil))}, options...)

	out = &Context[i, o]{Name: option.Name, ApiSourceRds: option.ApiSourceRds, Ctx: context.Background(),
		WithHeader: HeaderFieldsUsed(reflect.TypeOf(new(i)).Elem()),
		Validate:   needValidate(reflect.TypeOf(new(i)).Elem()),
		Func:       f,
	}

	if len(option.Name) == 0 {
		log.Debug().Msg("ApiNamed service created failed!")
		out.Func = func(InParameter i) (ret o, err error) {
			log.Warn().Str("service misnamed", out.Name).Send()
			return ret, vars.ErrApiNameEmpty
		}
	}

	ApiServices.Set(out.Name, out)

	funcPtr := reflect.ValueOf(f).Pointer()
	fun2Api.Set(funcPtr, out)

	apis, _ := APIGroupByRdsToReceiveJob.Get(out.ApiSourceRds)
	apis = append(apis, out.Name)
	APIGroupByRdsToReceiveJob.Set(out.ApiSourceRds, apis)

	log.Debug().Str("ApiNamed service created completed!", out.Name).Send()
	return out
}
