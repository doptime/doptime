package api

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/doptime/doptime/config"
	"github.com/doptime/doptime/dlog"
	"github.com/doptime/doptime/specification"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/redis/go-redis/v9"
)

// create Api context.
// This New function is for the case the API is defined outside of this package.
// If the API is defined in this package, use Api() instead.
// timeAt is ID of the task. if you want's to cancel the task, you should provide the same timeAt
func CallAt[i any, o any](f func(InParam i) (ret o, err error)) (callAtFun func(timeAt time.Time, InParam i) (err error)) {
	var (
		db      *redis.Client
		exists  bool
		ctx     = context.Background()
		apiInfo ApiInterface
	)
	funcPtr := reflect.ValueOf(f).Pointer()
	if apiInfo, exists = GetApiByFunc(funcPtr); !exists {
		dlog.Fatal().Str("service function should be defined By Api or Rpc before used in CallAt", specification.ApiNameByType((*i)(nil))).Send()
	}
	dataSource, apiName := apiInfo.GetDataSource(), apiInfo.GetName()
	if db, exists = config.Rds[dataSource]; !exists {
		dlog.Info().Str("DataSource not defined in enviroment", dataSource).Send()
		return nil
	}

	callAtFun = func(timeAt time.Time, InParam i) (err error) {
		var (
			b      []byte
			cmd    *redis.StringCmd
			Values []string
		)
		if b, err = specification.MarshalApiInput(InParam); err != nil {
			return err
		}
		fmt.Println("CallAt", apiName, timeAt.UnixNano())
		Values = []string{"timeAt", strconv.FormatInt(timeAt.UnixNano(), 10), "data", string(b)}
		args := &redis.XAddArgs{Stream: apiName, Values: Values, MaxLen: 4096}
		if cmd = db.XAdd(ctx, args); cmd.Err() != nil {
			dlog.Info().AnErr("Do XAdd", cmd.Err()).Send()
			return cmd.Err()
		}
		return nil

	}
	callAtfun2Api.Set(reflect.ValueOf(callAtFun).Pointer(), apiInfo)
	return callAtFun
}

var callAtfun2Api cmap.ConcurrentMap[uintptr, ApiInterface] = cmap.NewWithCustomShardingFunction[uintptr, ApiInterface](func(key uintptr) uint32 {
	hash := uint32(2166136261)
	const prime32 = uint32(16777619)
	return ((hash*prime32)^uint32(key))*prime32 ^ uint32(key>>32)
})
