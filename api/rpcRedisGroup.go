package api

import (
	"context"
	"strings"
	"time"

	"github.com/doptime/doptime/dlog"
	"github.com/redis/go-redis/v9"
)

var ServiceBatchSize int64 = 64

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
	args := &redis.XReadGroupArgs{Streams: streams, Block: time.Second * 20, Count: ServiceBatchSize, NoAck: true, Group: "group0", Consumer: "doptime"}
	return args
}
func XGroupEnsureCreatedOneGroup(c context.Context, serviceName string, rds *redis.Client) (err error) {
	var (
		cmdStream      *redis.XInfoStreamCmd
		cmdXInfoGroups *redis.XInfoGroupsCmd
		groups         []string
		group0Exists   bool = false
	)
	//if stream key does not exist, create a placeholder stream
	//other wise, NOGROUP No such key will be returned
	if cmdStream = rds.XInfoStream(c, serviceName); cmdStream.Err() != nil {
		dlog.Info().AnErr("XInfoStream not exist", cmdStream.Err()).Str("try recreating stream", serviceName).Send()
		//create a placeholder stream
		if cmd := rds.XAdd(c, &redis.XAddArgs{Stream: serviceName, MaxLen: 4096, Values: []string{"data", ""}}); cmd.Err() != nil {
			dlog.Info().AnErr("XAdd err in recreating stream while XGroupEnsureCreatedOneGroup", cmd.Err()).Send()
			return cmd.Err()
		}
	} else {
		dlog.Info().Str("XInfoStream success, key already exists", serviceName).Send()
	}
	//continue if the group already exists
	if cmdXInfoGroups = rds.XInfoGroups(c, serviceName); cmdXInfoGroups.Err() == nil && len(cmdXInfoGroups.Val()) > 0 {
		for _, group := range cmdXInfoGroups.Val() {
			if group.Name == "group0" {
				group0Exists = true
			}
			groups = append(groups, group.Name)
		}
		dlog.Info().Str("existing groups :", strings.Join(groups, ",")).Any("group0 exists", group0Exists).Send()
		if group0Exists {
			return nil
		}
	}
	//create a group if none exists
	if cmd := rds.XGroupCreateMkStream(c, serviceName, "group0", "$"); cmd.Err() != nil {
		dlog.Info().AnErr("XGroupCreateOne", cmd.Err()).Send()
		return cmd.Err()
	}
	dlog.Info().Str("XGroupCreateOne success", serviceName).Send()
	return nil

}
