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

func CallAtCancel[i any, o any](f func(InParam i) (ret o, err error), timeAt time.Time) (err error) {
	var (
		Rds     *redis.Client
		apiInfo *ApiInfo
		Values  []string
	)
	funcPtr := reflect.ValueOf(f).Pointer()
	if _apiInfo, ok := fun2ApiInfoMap.Load(funcPtr); !ok {
		log.Fatal().Str("service function should be defined By Api or Rpc before used in CallAt", specification.ApiNameByType((*i)(nil))).Send()
	} else {
		apiInfo = _apiInfo.(*ApiInfo)
	}
	if Rds, err = config.GetRdsClientByName(apiInfo.DataSource); err != nil {
		return err
	}
	Values = []string{"timeAt", strconv.FormatInt(timeAt.UnixNano(), 10), "data", ""}
	args := &redis.XAddArgs{Stream: apiInfo.Name, Values: Values, MaxLen: 4096}
	//use Rds.XAdd rather than Rds.HSet, to prevent Hset before receiing the result of  XAdd
	if cmd := Rds.XAdd(context.Background(), args); cmd.Err() != nil {
		log.Info().AnErr("Do XAdd", cmd.Err()).Send()
		return cmd.Err()
	}
	return nil
}
