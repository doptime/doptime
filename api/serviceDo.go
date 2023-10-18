package api

import (
	"context"
	"fmt"
	"regexp"
	"time"

	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"github.com/vmihailenco/msgpack/v5"
	"github.com/yangkequn/saavuu/config"
)

type ApiInfo struct {
	// ApiName is the name of the service
	ApiName string
	// ApiFuncWithMsgpackedParam is the function of the service
	ApiFuncWithMsgpackedParam func(s []byte) (ret interface{}, err error)
}

var ApiServices cmap.ConcurrentMap[string, *ApiInfo] = cmap.New[*ApiInfo]()

func apiServiceNames() (serviceNames []string) {
	for _, serviceInfo := range ApiServices.Items() {
		serviceNames = append(serviceNames, serviceInfo.ApiName)
	}
	return serviceNames
}
func receiveJobs() {
	var (
		cmd     *redis.XStreamSliceCmd
		apiName string
		err     error
	)
	c := context.Background()
	//create group if none exists, with consumer saavuu
	for _, serviceName := range apiServiceNames() {
		if err = XGroupCreateOne(c, serviceName); err != nil {
			time.Sleep(time.Second)
		}
	}

	//deprecate using list command LRange, to avoid continually query consumption
	//use xreadgroup to receive data ,2023-01-31
	for args := defaultXReadGroupArgs(); ; {
		if cmd = config.Rds.XReadGroup(c, args); cmd.Err() == redis.Nil {
			continue
		} else if cmd.Err() != nil {
			log.Info().AnErr("receiveApiJobs", cmd.Err()).Send()
			//2023-10-18T05:39:41Z INF receiveApiJobs=NOGROUP No such key 'api:skillSearch' or consumer group 'group0' in XREADGROUP with GROUP option
			matchServiceName := regexp.MustCompile(`No such key '(api:.*)' or consumer `)
			if matchServiceName.MatchString(cmd.Err().Error()) {
				GroupName := matchServiceName.FindStringSubmatch(cmd.Err().Error())[1]
				log.Info().Str("Try recreate group", GroupName).Send()
				if cmd := config.Rds.Del(c, GroupName); cmd.Err() == nil {
					XGroupCreateOne(c, GroupName)
				}
			}
			time.Sleep(time.Second)
			continue
		}

		for _, stream := range cmd.Val() {
			apiName = stream.Stream
			for _, message := range stream.Messages {
				bytesValue := message.Values["data"].(string)
				//the delay calling will lost if the app is down
				if dueTimeStr, ok := message.Values["dueTime"]; ok {
					go delayTaskAddOne(apiName, dueTimeStr.(string), bytesValue)
				} else {
					go DoOneJob(apiName, apiName, []byte(bytesValue))
				}
				apiCounter.Add(apiName, 1)
			}
		}
	}
}
func DoOneJob(apiName, BackToID string, s []byte) (err error) {
	var (
		msgPackResult []byte
		ret           interface{}
		service       *ApiInfo
		ok            bool
	)
	if service, ok = ApiServices.Get(apiName); !ok {
		return fmt.Errorf("service %s not found", apiName)
	}
	if ret, err = service.ApiFuncWithMsgpackedParam(s); err != nil {
		return err
	}
	if msgPackResult, err = msgpack.Marshal(ret); err != nil {
		return
	}
	ctx := context.Background()
	pipline := config.Rds.Pipeline()
	pipline.RPush(ctx, BackToID, msgPackResult)
	pipline.Expire(ctx, BackToID, time.Second*20)
	_, err = pipline.Exec(ctx)
	return err
}
