package api

import (
	"context"
	"reflect"
	"strconv"
	"time"

	"github.com/doptime/doptime/config"
	"github.com/doptime/doptime/specification"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

func CallAtCancel[i any](f func(timeAt time.Time, InParam i) (err error), timeAt time.Time) (err error) {
	var (
		Rds    *redis.Client
		api    ApiInterface
		Values []string
		ok     bool
	)
	funcPtr := reflect.ValueOf(f).Pointer()
	if api, ok = callAtfun2Api.Get(funcPtr); !ok {
		log.Fatal().Str("service function should be defined By Api or Rpc before used in CallAt", specification.ApiNameByType((*i)(nil))).Send()
	}
	if Rds, err = config.GetRdsClientByName(api.GetDataSource()); err != nil {
		return err
	}
	Values = []string{"timeAt", strconv.FormatInt(timeAt.UnixNano(), 10), "data", ""}
	args := &redis.XAddArgs{Stream: api.GetName(), Values: Values, MaxLen: 4096}
	//use Rds.XAdd rather than Rds.HSet, to prevent Hset before receiing the result of  XAdd
	if cmd := Rds.XAdd(context.Background(), args); cmd.Err() != nil {
		log.Info().AnErr("Do XAdd", cmd.Err()).Send()
		return cmd.Err()
	}
	return nil
}
