package api

import (
	"context"
	"reflect"

	"github.com/doptime/doptime/httpserve/httpapi"
	"github.com/doptime/doptime/httpserve/httpdoc"
	"github.com/doptime/doptime/utils"
	"github.com/doptime/doptime/vars"
	"github.com/doptime/logger"
	"github.com/doptime/redisdb"
)

func Api[i any, o any](f func(InParameter i) (ret o, err error), options ...optionSetter) (out *ApiCtx[i, o]) {
	var targetType reflect.Type = reflect.TypeOf(new(i)).Elem()
	var option *Option = Option{ApiSourceRds: "default", ApiKey: utils.ApiNameByType(reflect.Zero(targetType).Interface())}.mergeNewOptions(options...)

	out = &ApiCtx[i, o]{Name: option.ApiKey, ApiSourceRds: option.ApiSourceRds, Ctx: context.Background(),
		Validate: redisdb.NeedValidate(reflect.TypeOf(new(i)).Elem()),
		Func:     f,
	}

	if len(out.Name) == 0 {
		logger.Debug().Msg("ApiNamed service created failed!")
		out.Func = func(InParameter i) (ret o, err error) {
			logger.Warn().Str("service misnamed", out.Name).Send()
			return ret, vars.ErrApiNameEmpty
		}
	}

	// Error handling: Check for naming conflicts
	if _, exists := httpapi.ApiViaHttp.Get(out.Name); exists {
		logger.Panic().Str("same service not allowed to defined twice!", out.Name).Send()
		return out
	}

	httpapi.ApiViaHttp.Set(out.Name, out)

	funcPtr := reflect.ValueOf(f).Pointer()
	httpapi.Fun2Api.Set(funcPtr, out)

	apis, _ := APIGroupByRdsToReceiveJob.Get(out.ApiSourceRds)
	apis = append(apis, out.Name)
	APIGroupByRdsToReceiveJob.Set(out.ApiSourceRds, apis)

	iType := reflect.TypeOf((*i)(nil)).Elem()
	oType := reflect.TypeOf((*o)(nil)).Elem()
	httpdoc.RegisterApi(out.Name, iType, oType)

	logger.Debug().Str("ApiNamed service created completed!", out.Name).Send()
	return out
}
