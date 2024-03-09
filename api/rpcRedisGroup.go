package api

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"github.com/yangkequn/goflow/config"
)

func defaultXReadGroupArgs(serviceNames []string) *redis.XReadGroupArgs {
	var (
		streams []string
	)
	streams = append(streams, serviceNames...)
	//from services to ServiceInfos
	for i := 0; i < len(serviceNames); i++ {
		//append default stream id
		streams = append(streams, ">")
	}

	//ServiceBatchSize is the number of tasks that a service can read from redis at the same time
	args := &redis.XReadGroupArgs{Streams: streams, Block: time.Second * 20, Count: config.Cfg.Api.ServiceBatchSize, NoAck: true, Group: "group0", Consumer: "goflow"}
	return args
}
func XGroupEnsureCreated(c context.Context, ServiceNames []string, rds *redis.Client) (err error) {
	var (
		waitGroup sync.WaitGroup
	)
	waitGroup.Add(len(ServiceNames))

	XGroupCreateOne := func(c context.Context, serviceName string) (err error) {
		var (
			cmdStream      *redis.XInfoStreamCmd
			cmdXInfoGroups *redis.XInfoGroupsCmd
		)
		defer waitGroup.Done()
		//if stream key does not exist, create a placeholder stream
		//other wise, NOGROUP No such key will be returned
		if cmdStream = rds.XInfoStream(c, serviceName); cmdStream.Err() != nil {
			log.Info().AnErr("XInfoStream not exist", cmdStream.Err()).Str("try recreating streadming", serviceName).Send()
			if cmdStream.Err() == redis.Nil {
				//create a placeholder stream
				if cmd := rds.XAdd(c, &redis.XAddArgs{Stream: serviceName, MaxLen: 4096, Values: []string{"data", ""}}); cmd.Err() != nil {
					log.Info().AnErr("XAdd", cmd.Err()).Send()
					return cmd.Err()
				}
			} else {
				log.Info().AnErr("XInfoStream", cmdStream.Err()).Send()
				return cmdStream.Err()
			}
		}
		//continue if the group already exists
		if cmdXInfoGroups = rds.XInfoGroups(c, serviceName); cmdXInfoGroups.Err() == nil && len(cmdXInfoGroups.Val()) > 0 {
			var groups []string
			var group0Exists bool = false
			for _, group := range cmdXInfoGroups.Val() {
				if group.Name == "group0" {
					group0Exists = true
				}

				groups = append(groups, group.Name)
			}
			log.Info().Str("existing groups :", strings.Join(groups, ",")).Any("group0 exists", group0Exists).Send()
			if group0Exists {
				return nil
			}
		}
		//create a group if none exists
		if cmd := rds.XGroupCreateMkStream(c, serviceName, "group0", "$"); cmd.Err() != nil {
			log.Info().AnErr("XGroupCreateOne", cmd.Err()).Send()
			return cmd.Err()
		}
		log.Info().Str("XGroupCreateOne success", serviceName).Send()

		return nil
	}
	//return until all groups are created
	for _, serviceName := range ServiceNames {
		go XGroupCreateOne(c, serviceName)
	}
	waitGroup.Wait()
	return nil

}
