package api

import (
	"context"
	"reflect"

	"github.com/doptime/doptime/dlog"
	"github.com/doptime/doptime/specification"
	"github.com/doptime/doptime/vars"
)

func Api[i any, o any](
	f func(InParameter i) (ret o, err error),
	options ...*Option,
) (out *Context[i, o]) {
	var option *Option = mergeNewOptions(&Option{ApiSourceRds: "default"}, options...)

	out = &Context[i, o]{Name: specification.ApiNameByType((*i)(nil)), ApiSourceRds: option.ApiSourceRds, Ctx: context.Background(),
		WithHeader: HeaderFieldsUsed(reflect.TypeOf(new(i)).Elem()),
		Validate:   needValidate(reflect.TypeOf(new(i)).Elem()),
		Func:       f,
	}

	if len(out.Name) == 0 {
		dlog.Debug().Msg("ApiNamed service created failed!")
		out.Func = func(InParameter i) (ret o, err error) {
			dlog.Warn().Str("service misnamed", out.Name).Send()
			return ret, vars.ErrApiNameEmpty
		}
	}
	// Error handling: Check for naming conflicts
	if _, exists := ApiServices.Get(out.Name); exists {
		dlog.Panic().Str("same service not allowed to defined twice!", out.Name).Send()
		return out
	}

	ApiServices.Set(out.Name, out)

	funcPtr := reflect.ValueOf(f).Pointer()
	fun2Api.Set(funcPtr, out)

	apis, _ := APIGroupByRdsToReceiveJob.Get(out.ApiSourceRds)
	apis = append(apis, out.Name)
	APIGroupByRdsToReceiveJob.Set(out.ApiSourceRds, apis)

	out.RegisterApi(option.PublishInfo)
	dlog.Debug().Str("ApiNamed service created completed!", out.Name).Send()
	return out
}
