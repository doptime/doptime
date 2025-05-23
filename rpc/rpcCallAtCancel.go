package rpc

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/doptime/config/cfgredis"
	"github.com/doptime/doptime/httpserve/httpapi"
	"github.com/doptime/doptime/utils"
	"github.com/doptime/logger"
	"github.com/redis/go-redis/v9"
)

func CallAtCancel[i any](f func(timeAt time.Time, InParam i) (err error), timeAt time.Time) (err error) {
	var (
		Rds    *redis.Client
		api    httpapi.ApiInterface
		Values []string
		exists bool
	)
	funcPtr := reflect.ValueOf(f).Pointer()
	if api, exists = callAtfun2Api.Get(funcPtr); !exists {
		logger.Fatal().Str("service function should be defined By Api or Rpc before used in CallAt", utils.ApiNameByType((*i)(nil))).Send()
	}
	if Rds, exists = cfgredis.Servers.Get(api.GetDataSource()); !exists {
		logger.Info().Str("DataSource not defined in enviroment", api.GetDataSource()).Send()
		return fmt.Errorf("DataSource not defined in enviroment")
	}
	Values = []string{"timeAt", strconv.FormatInt(timeAt.UnixNano(), 10), "data", ""}
	args := &redis.XAddArgs{Stream: api.GetName(), Values: Values, MaxLen: 4096}
	//use Rds.XAdd rather than Rds.HSet, to prevent Hset before receiing the result of  XAdd
	if cmd := Rds.XAdd(context.Background(), args); cmd.Err() != nil {
		logger.Info().AnErr("Do XAdd", cmd.Err()).Send()
		return cmd.Err()
	}
	return nil
}
