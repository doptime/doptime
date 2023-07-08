package api

import (
	"context"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/vmihailenco/msgpack/v5"
	"github.com/yangkequn/saavuu/config"
	"github.com/yangkequn/saavuu/logger"
)

// put parameter to redis ,make it persistent
func delayTaskAddOne(serviceName string, dueTimeStr string, bytesValue string) {
	if cmd := config.Rds.HSet(context.Background(), serviceName+":delay", dueTimeStr, bytesValue); cmd.Err() != nil {
		logger.Lshortfile.Println(cmd.Err())
		return
	}
	go delayTaskDoOne(serviceName, dueTimeStr)
}
func delayTaskDoOne(serviceName, dueTimeStr string) {
	var (
		ret                                        interface{}
		bytes, msgPackResult                       []byte
		dueTimeUnixMilliSecond, nowUnixMilliSecond int64
		err                                        error
		cmd                                        []redis.Cmder
	)
	nowUnixMilliSecond = time.Now().UnixMilli()
	if dueTimeUnixMilliSecond, err = strconv.ParseInt(dueTimeStr, 10, 64); err != nil {
		logger.Lshortfile.Println(err)
		return
	}
	time.Sleep(time.Duration(dueTimeUnixMilliSecond-nowUnixMilliSecond) * time.Millisecond)
	pipeline := config.Rds.Pipeline()
	pipeline.HGet(context.Background(), serviceName+":delay", dueTimeStr)
	pipeline.HDel(context.Background(), serviceName+":delay", dueTimeStr)
	if cmd, err = pipeline.Exec(context.Background()); err != nil {
		logger.Lshortfile.Println(err)
		return
	}
	if bytes, err = cmd[0].(*redis.StringCmd).Bytes(); err == nil {
		if ret, err = ApiServices[serviceName].ApiFuncWithMsgpackedParam(bytes); err != nil {
			logger.Lshortfile.Println(err)
			return
		}
		if msgPackResult, err = msgpack.Marshal(ret); err != nil {
			return
		}

		//Post Back
		ctx := context.Background()
		pipline := config.Rds.Pipeline()
		var BackToID = serviceName
		pipline.RPush(ctx, BackToID, msgPackResult)
		pipline.Expire(ctx, BackToID, time.Second*20)
		pipline.Exec(ctx)
	}
}

func delayTasksLoad() {
	var (
		services    = apiServiceNames()
		dueTimeStrs []string
		cmd         []redis.Cmder
		err         error
	)
	pipeline := config.Rds.Pipeline()
	for _, service := range services {
		pipeline.HKeys(context.Background(), service+":delay")
	}
	if cmd, err = pipeline.Exec(context.Background()); err != nil {
		logger.Lshortfile.Println("err LoadDelayApiTask, ", err)
		return
	}
	for i, service := range services {
		dueTimeStrs = cmd[i].(*redis.StringSliceCmd).Val()
		for _, dueTimeStr := range dueTimeStrs {
			go delayTaskDoOne(service, dueTimeStr)
		}
	}
}
