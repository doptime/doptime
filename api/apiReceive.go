package api

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/doptime/config/cfgredis"
	"github.com/doptime/doptime/httpserve/httpapi"
	"github.com/doptime/logger"
	"github.com/redis/go-redis/v9"
	"github.com/vmihailenco/msgpack/v5"
)

// ensure all apis can be called by rpc
// because rpc receive needs to know all api names to create stream reading
var ApiStartingWaiter func() = func() func() {
	LastApiCnt := -1
	//if the count of apis is not changing, then all apis are loaded
	Checker := func() {
		//if ApiServices.Count() no longer changed, then all apis are loaded
		for _cnt := httpapi.ApiViaHttp.Count(); _cnt == 0 || LastApiCnt != _cnt; _cnt = httpapi.ApiViaHttp.Count() {
			time.Sleep(time.Millisecond * 30)
			LastApiCnt = _cnt
		}
	}
	return Checker
}()

func rpcReceive() {
	var (
		rds      *redis.Client
		services []string
		exists   bool
	)
	for _, dataSource := range APIGroupByRdsToReceiveJob.Keys() {
		if services, exists = APIGroupByRdsToReceiveJob.Get(dataSource); !exists {
			logger.Error().Str("dataSource missing in APIGroupByRdsToReceiveJob", dataSource).Send()
			continue
		}

		if rds, exists = cfgredis.Servers.Get(dataSource); !exists {
			logger.Error().Str("dataSource missing in rpcReceive", dataSource).Send()
			continue
		}
		go rpcReceiveOneDatasource(services, rds)
	}
}
func rpcReceiveOneDatasource(serviceNames []string, rds *redis.Client) {
	var (
		apiName, data string
		cmd           *redis.XStreamSliceCmd
	)

	//wait for all rpc services ready, so that rpc results can be received
	ApiStartingWaiter()

	c := context.Background()

	//deprecate using list command LRange, to avoid continually query consumption
	//use xreadgroup to receive data ,2023-01-31
	for args := defaultXReadGroupArgs(serviceNames); ; {
		if cmd = rds.XReadGroup(c, args); cmd.Err() == redis.Nil {
			continue
		} else if cmd.Err() != nil {
			logger.Error().AnErr("rpcReceiveError", cmd.Err()).Send()
			//ensure the stream is created
			if items := strings.Split(cmd.Err().Error(), "api:"); len(items) > 1 {
				if items2 := strings.Split("api:"+items[1], "'"); len(items2) > 1 {
					logger.Info().Str("starting XGroupEnsureCreatedOneGroup", items2[0]).Send()
					go XGroupEnsureCreatedOneGroup(c, items2[0], rds)
				}
			} else {
				logger.Error().AnErr("No API name Captured between No such key 'xxx'", cmd.Err()).Send()
			}

			time.Sleep(time.Second)
		}

		for _, stream := range cmd.Val() {
			apiName = stream.Stream
			for _, message := range stream.Messages {
				timeAtStr, atOk := message.Values["timeAt"]
				//skip case of placeholder stream while not atOk
				//but if timeAt is setted, then empty data is allowed, used to clear the task
				if data = message.Values["data"].(string); len(data) == 0 && !atOk {
					continue
				}
				//the delay calling will lost if the app is down
				if atOk {
					if len(data) == 0 {
						rpcCallAtTaskRemoveOne(apiName, timeAtStr.(string))
					} else {
						rpcCallAtTaskAddOne(apiName, timeAtStr.(string), data)
					}
				} else {
					go CallApiLocallyAndSendBackResult(apiName, message.ID, []byte(data))
				}
				httpapi.ApiCounter.Add(apiName, 1)
			}
		}
	}
}
func CallApiLocallyAndSendBackResult(apiName, BackToID string, s []byte) (err error) {
	var (
		msgPackResult []byte
		ret           interface{}
		service       httpapi.ApiInterface
		rds           *redis.Client
		exists        bool
	)
	if service, exists = httpapi.ApiViaHttp.Get(apiName); !exists {
		return fmt.Errorf("service %s not found", apiName)
	}
	var _map = map[string]interface{}{}
	var msgpackNonstruct []byte
	if err = msgpack.Unmarshal(s, &_map); err != nil {
		msgpackNonstruct = s
	}
	if ret, err = service.CallByMap(_map, msgpackNonstruct, nil); err != nil {
		return err
	}
	if msgPackResult, err = msgpack.Marshal(ret); err != nil {
		return
	}
	ctx := context.Background()
	DataSource := service.GetDataSource()
	if rds, exists = cfgredis.Servers.Get(DataSource); !exists {
		logger.Error().Str("DataSource not defined in enviroment while CallApiLocallyAndSendBackResult", DataSource).Send()
		return fmt.Errorf("DataSource not defined in enviroment %s", DataSource)
	}
	pipline := rds.Pipeline()
	pipline.RPush(ctx, BackToID, msgPackResult)
	pipline.Expire(ctx, BackToID, time.Second*20)
	_, err = pipline.Exec(ctx)
	return err
}
