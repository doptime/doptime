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
	"github.com/redis/go-redis/v9"
)

func CallAtCancel[i any](f func(timeAt time.Time, InParam i) (err error), timeAt time.Time) (err error) {
	var (
		Rds    *redis.Client
		api    ApiInterface
		Values []string
		exists bool
	)
	funcPtr := reflect.ValueOf(f).Pointer()
	if api, exists = callAtfun2Api.Get(funcPtr); !exists {
		dlog.Fatal().Str("service function should be defined By Api or Rpc before used in CallAt", specification.ApiNameByType((*i)(nil))).Send()
	}
	if Rds, exists = config.Rds[api.GetDataSource()]; !exists {
		dlog.Info().Str("DataSource not defined in enviroment", api.GetDataSource()).Send()
		return fmt.Errorf("DataSource not defined in enviroment")
	}
	Values = []string{"timeAt", strconv.FormatInt(timeAt.UnixNano(), 10), "data", ""}
	args := &redis.XAddArgs{Stream: api.GetName(), Values: Values, MaxLen: 4096}
	//use Rds.XAdd rather than Rds.HSet, to prevent Hset before receiing the result of  XAdd
	if cmd := Rds.XAdd(context.Background(), args); cmd.Err() != nil {
		dlog.Info().AnErr("Do XAdd", cmd.Err()).Send()
		return cmd.Err()
	}
	return nil
}
