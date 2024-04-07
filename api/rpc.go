package api

import (
	"context"
	"errors"
	"reflect"
	"time"

	"github.com/doptime/doptime/config"
	"github.com/doptime/doptime/specification"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"github.com/vmihailenco/msgpack/v5"
)

// create Api context.
// This New function is for the case the API is defined outside of this package.
// If the API is defined in this package, use Api() instead.
func Rpc[i any, o any](options ...*ApiOption) (rpc *Context[i, o]) {

	var option *ApiOption = mergeNewOptions(&ApiOption{ApiSourceRds: "default", Name: specification.ApiNameByType((*i)(nil))}, options...)

	rpc = &Context[i, o]{Name: option.Name, ApiSourceRds: option.ApiSourceRds, Ctx: context.Background(),
		WithHeader: HeaderFieldsUsed(reflect.TypeOf(new(i)).Elem()),
		Validate:   needValidate(reflect.TypeOf(new(i)).Elem()),
	}

	rpc.Fun = func(InParam i) (ret o, err error) {

		var (
			results []string
			cmd     *redis.StringCmd
			b       []byte
			db      *redis.Client
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
		if db, err = config.GetRdsClientByName(rpc.ApiSourceRds); err != nil {
			log.Info().Str("DataSource not defined in enviroment", rpc.ApiSourceRds).Send()
			return ret, err
		}
		args := &redis.XAddArgs{Stream: rpc.Name, Values: Values, MaxLen: 4096}
		if cmd = db.XAdd(rpc.Ctx, args); cmd.Err() != nil {
			log.Info().AnErr("Do XAdd", cmd.Err()).Send()
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

	funcPtr := reflect.ValueOf(rpc.Fun).Pointer()
	fun2Api.Set(funcPtr, rpc)
	return rpc
}
