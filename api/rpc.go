package api

import (
	"context"
	"errors"
	"reflect"
	"time"

	"github.com/doptime/config/cfgredis"
	"github.com/doptime/doptime/specification"
	"github.com/doptime/logger"
	"github.com/redis/go-redis/v9"
	"github.com/vmihailenco/msgpack/v5"
)

// create Api context.
// This New function is for the case the API is defined outside of this package.
// If the API is defined in this package, use Api() instead.
func Rpc[i any, o any](options ...optionSetter) (rpc *Context[i, o]) {

	var option *Option = Option{ApiSourceRds: "default"}.mergeNewOptions(options...)

	rpc = &Context[i, o]{Name: specification.ApiNameByType((*i)(nil)), ApiSourceRds: option.ApiSourceRds, Ctx: context.Background(),
		WithHeader: HeaderFieldsUsed(reflect.TypeOf(new(i)).Elem()),
		Validate:   needValidate(reflect.TypeOf(new(i)).Elem()),
	}

	rpc.Func = func(InParam i) (ret o, err error) {

		var (
			results []string
			cmd     *redis.StringCmd
			b       []byte
			db      *redis.Client
			exists  bool
		)
		if b, err = specification.MarshalApiInput(InParam); err != nil {
			return ret, err
		}
		var Values = []string{"data", string(b)}

		// if hashCallAt {
		// 	Values = []string{"timeAt", strconv.FormatInt(ops.CallAt.UnixMilli(), 10), "data", string(b)}
		// } else {
		// 	Values = []string{"data", string(b)}
		// }
		if db, exists = cfgredis.Servers.Get(rpc.ApiSourceRds); !exists {
			logger.Info().Str("DataSource not defined in enviroment", rpc.ApiSourceRds).Send()
			return ret, err
		}
		args := &redis.XAddArgs{Stream: rpc.Name, Values: Values, MaxLen: 4096}
		if cmd = db.XAdd(rpc.Ctx, args); cmd.Err() != nil {
			logger.Info().AnErr("Do XAdd", cmd.Err()).Send()
			return ret, cmd.Err()
		}
		// if hashCallAt {
		// 	return out, nil
		// }

		//BLPop 返回结果 [key1,value1,key2,value2]
		//cmd.Val() is the stream id, the result will be poped from the list with this id
		if results, err = db.BLPop(context.Background(), time.Second*6, cmd.Val()).Result(); err != nil {
			return ret, err
		}

		if len(results) != 2 {
			return ret, errors.New("BLPop result length error")
		}
		b = []byte(results[1])
		oType := reflect.TypeOf((*o)(nil)).Elem()
		//if o type is a pointer, use reflect.New to create a new pointer
		if oType.Kind() == reflect.Ptr {
			ret = reflect.New(oType.Elem()).Interface().(o)
			return ret, msgpack.Unmarshal(b, ret)
		}
		oValueWithPointer := reflect.New(oType).Interface().(*o)
		return *oValueWithPointer, msgpack.Unmarshal(b, oValueWithPointer)
	}

	ApiServices.Set(rpc.Name, rpc)

	funcPtr := reflect.ValueOf(rpc.Func).Pointer()
	fun2Api.Set(funcPtr, rpc)
	return rpc
}
