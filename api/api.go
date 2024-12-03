package api

import (
	"context"
	"reflect"

	"github.com/doptime/doptime/httpdoc"
	"github.com/doptime/doptime/specification"
	"github.com/doptime/doptime/vars"
	"github.com/doptime/logger"
)

func Api[i any, o any](f func(InParameter i) (ret o, err error), options ...optionSetter) (out *Context[i, o]) {
	var option *Option = Option{ApiSourceRds: "default"}.mergeNewOptions(options...)

	out = &Context[i, o]{Name: specification.ApiNameByType((*i)(nil)), ApiSourceRds: option.ApiSourceRds, Ctx: context.Background(),
		WithHeader: HeaderFieldsUsed(reflect.TypeOf(new(i)).Elem()),
		Validate:   needValidate(reflect.TypeOf(new(i)).Elem()),
		Func:       f,
	}

	if len(out.Name) == 0 {
		logger.Debug().Msg("ApiNamed service created failed!")
		out.Func = func(InParameter i) (ret o, err error) {
			logger.Warn().Str("service misnamed", out.Name).Send()
			return ret, vars.ErrApiNameEmpty
		}
	}

	// Error handling: Check for naming conflicts
	if _, exists := ApiServices.Get(out.Name); exists {
		logger.Panic().Str("same service not allowed to defined twice!", out.Name).Send()
		return out
	}

	ApiServices.Set(out.Name, out)

	funcPtr := reflect.ValueOf(f).Pointer()
	fun2Api.Set(funcPtr, out)

	apis, _ := APIGroupByRdsToReceiveJob.Get(out.ApiSourceRds)
	apis = append(apis, out.Name)
	APIGroupByRdsToReceiveJob.Set(out.ApiSourceRds, apis)

	iType := reflect.TypeOf((*i)(nil)).Elem()
	oType := reflect.TypeOf((*o)(nil)).Elem()
	httpdoc.RegisterApi(out.Name, iType, oType)

	logger.Debug().Str("ApiNamed service created completed!", out.Name).Send()
	return out
}
