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
func Rpc[i any, o any](options ...*ApiOption) (retf func(InParam i) (ret o, err error)) {
	var (
		db     *redis.Client
		err    error
		ctx               = context.Background()
		option *ApiOption = &ApiOption{DataSource: "default"}
	)
	if len(options) > 0 {
		option = options[0]
	}

	if len(option.Name) > 0 {
		option.Name = specification.ApiName(option.Name)
	}
	if len(option.Name) == 0 {
		option.Name = specification.ApiNameByType((*i)(nil))
	}
	if len(option.Name) == 0 {
		log.Error().Str("service misnamed", option.Name).Send()
	}

	if db, err = config.GetRdsClientByName(option.DataSource); err != nil {
		log.Info().Str("DataSource not defined in enviroment", option.DataSource).Send()
		return nil
	}
	ProcessOneJob := func(s []byte) (ret interface{}, err error) {
		var (
			Values  = []string{"data", string(s)}
			results []string
			cmd     *redis.StringCmd
			b       []byte
		)
		// if hashCallAt {
		// 	Values = []string{"timeAt", strconv.FormatInt(ops.CallAt.UnixMilli(), 10), "data", string(b)}
		// } else {
		// 	Values = []string{"data", string(b)}
		// }
		args := &redis.XAddArgs{Stream: option.Name, Values: Values, MaxLen: 4096}
		if cmd = db.XAdd(ctx, args); cmd.Err() != nil {
			log.Info().AnErr("Do XAdd", cmd.Err()).Send()
			return "", cmd.Err()
		}
		// if hashCallAt {
		// 	return out, nil
		// }

		//BLPop 返回结果 [key1,value1,key2,value2]
		//cmd.Val() is the stream id, the result will be poped from the list with this id
		if results, err = db.BLPop(ctx, time.Second*6, cmd.Val()).Result(); err != nil {
			return "", err
		}

		if len(results) != 2 {
			return "", errors.New("BLPop result length error")
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

	retf = func(InParam i) (out o, err error) {
		var (
			b []byte
		)
		if b, err = specification.MarshalApiInput(InParam); err != nil {
			return out, err
		}
		ret, er := ProcessOneJob(b)
		return ret.(o), er
	}
	rpcInfo := &ApiInfo{
		DataSource:                option.DataSource,
		Name:                      option.Name,
		WithHeader:                HeaderFieldsUsed(reflect.TypeOf(new(i)).Elem()),
		WithJwt:                   WithJwtFields(reflect.TypeOf(new(i)).Elem()),
		ApiFuncWithMsgpackedParam: ProcessOneJob,
	}
	funcPtr := reflect.ValueOf(retf).Pointer()
	fun2ApiInfoMap.Store(funcPtr, rpcInfo)
	APIGroupByDataSource.Upsert(option.DataSource, []string{}, func(exist bool, valueInMap, newValue []string) []string {
		return append(valueInMap, option.Name)
	})
	return retf
}
