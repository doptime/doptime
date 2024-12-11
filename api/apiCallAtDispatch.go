package api

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/doptime/config/cfgredis"
	"github.com/doptime/doptime/httpserve/httpapi"
	"github.com/doptime/logger"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/redis/go-redis/v9"
)

type TaskAtFuture struct {
	ServiceName  string
	TimeAtUnixNs int64
}

var TasksAtFutureList = []*TaskAtFuture{}
var mut sync.Mutex = sync.Mutex{}

func GetServiceDB(serviceName string) (db *redis.Client, ok bool) {
	var (
		exists bool
	)
	serviceInfo, _ := httpapi.ApiViaHttp.Get(serviceName)
	DataSource := serviceInfo.GetDataSource()
	if db, exists = cfgredis.Servers.Get(DataSource); !exists {
		logger.Error().Str("DataSource not defined in enviroment. Please check the configuration", DataSource).Send()
	}
	return db, exists
}

// the reason why rpc can be removed locally is that the when doing rpc. api will recheck the data. only non empty data will be processed
func rpcCallAtTaskRemoveOne(serviceName string, timeAtStr string) {
	var (
		rds, redisExists = GetServiceDB(serviceName)
		TimeAtUnixNs     int64
		err              error
	)
	if !redisExists {
		logger.Error().Str("rpcCallAtTaskRemoveOne missing redis server", serviceName).Send()
		return
	}

	if TimeAtUnixNs, err = strconv.ParseInt(timeAtStr, 10, 64); err != nil {
		logger.Info().Err(err).Send()
		return
	}

	index := sort.Search(len(TasksAtFutureList), func(i int) bool {
		return TasksAtFutureList[i].TimeAtUnixNs == TimeAtUnixNs && TasksAtFutureList[i].ServiceName == serviceName
	})
	if index >= 0 && index < len(TasksAtFutureList) {
		mut.Lock()
		TasksAtFutureList = append(TasksAtFutureList[:index], TasksAtFutureList[index+1:]...)
		mut.Unlock()
		go rds.HDel(context.Background(), serviceName+":delay", timeAtStr)
	}
}

// put parameter to redis ,make it persistent
func rpcCallAtTaskAddOne(serviceName string, timeAtStr string, bytesValue string) {
	var (
		rds, redisExists = GetServiceDB(serviceName)
		err              error
	)
	if !redisExists {
		logger.Error().Str("rpcCallAtTaskAddOne missing redis server", serviceName).Send()
		return
	}
	task := &TaskAtFuture{ServiceName: serviceName}
	if task.TimeAtUnixNs, err = strconv.ParseInt(timeAtStr, 10, 64); err != nil {
		logger.Info().Err(err).Send()
		return
	}
	if cmd := rds.HSet(context.Background(), serviceName+":delay", timeAtStr, bytesValue); cmd.Err() != nil {
		logger.Info().Err(cmd.Err()).Send()
		return
	}
	index := sort.Search(len(TasksAtFutureList), func(i int) bool { return TasksAtFutureList[i].TimeAtUnixNs < task.TimeAtUnixNs })

	// Insert the new task into the TasksAtFuture at the found index.
	mut.Lock()
	TasksAtFutureList = append(TasksAtFutureList[:index], append([]*TaskAtFuture{task}, TasksAtFutureList[index:]...)...)
	mut.Unlock()
}
func rpcCallAtDispatcher() {
	var (
		data                  string
		TaskAtFutureNs, nowNs int64
		err                   error
		cmd                   []redis.Cmder
		rds                   *redis.Client
		redisExists           bool
	)
	for {
		if len(TasksAtFutureList) == 0 {
			time.Sleep(time.Millisecond * 100)
			continue
		}
		nowNs = time.Now().UnixNano()
		mut.Lock()
		task := TasksAtFutureList[0]
		mut.Unlock()
		TaskAtFutureNs = task.TimeAtUnixNs

		if timeSpan := TaskAtFutureNs - nowNs; timeSpan > 0 {
			if timeSpan > 100*1000*1000 {
				timeSpan = 100 * 1000 * 1000
			}
			time.Sleep(time.Duration(timeSpan))
			continue
		}
		TasksAtFutureList = TasksAtFutureList[1:]
		strTime := strconv.FormatInt(TaskAtFutureNs, 10)
		if rds, redisExists = GetServiceDB(task.ServiceName); !redisExists {
			logger.Error().Str("rpcCallAtDispatcher missing redis server", task.ServiceName).Send()
			return
		}
		pipeline := rds.Pipeline()
		pipeline.HGet(context.Background(), task.ServiceName+":delay", strTime)
		pipeline.HDel(context.Background(), task.ServiceName+":delay", strTime)
		if cmd, err = pipeline.Exec(context.Background()); err != nil {
			logger.Info().Err(err).Send()
			continue
		} else if len(cmd) != 2 || cmd[1].Err() != nil {
			continue
		}
		if data = cmd[0].(*redis.StringCmd).Val(); len(data) == 0 {
			continue
		}
		fmt.Println("rpcCallAtRoutine key", task.ServiceName+":delay field", TaskAtFutureNs, "data length", len(data), "data", string(data))
		CallApiLocallyAndSendBackResult(task.ServiceName, strconv.FormatInt(TaskAtFutureNs, 10), []byte(data))
	}
}

var APIGroupByRdsToReceiveJob = cmap.New[[]string]()

func rpcCallAtTasksLoad() {
	var (
		timeAtStrs []string
		cmd        []redis.Cmder
		err        error
		rds        *redis.Client
		exists     bool
	)
	logger.Info().Msg("rpcCallAtTasksLoading started")
	var _TasksAtFutureList = []*TaskAtFuture{}
	for _, dataSource := range APIGroupByRdsToReceiveJob.Keys() {
		services, ok := APIGroupByRdsToReceiveJob.Get(dataSource)
		if !ok {
			continue
		}
		if rds, exists = cfgredis.Servers.Get(dataSource); !exists {
			logger.Info().AnErr("err LoadDelayApiTask, ", err).Send()
			continue
		}
		pipeline := rds.Pipeline()
		for _, service := range services {
			pipeline.HKeys(context.Background(), service+":delay")
		}
		if cmd, err = pipeline.Exec(context.Background()); err != nil {
			logger.Info().AnErr("err LoadDelayApiTask, ", err).Send()
			continue
		}

		for i, service := range services {
			if err = cmd[i].(*redis.StringSliceCmd).Err(); err != nil {
				continue
			}
			timeAtStrs = cmd[i].(*redis.StringSliceCmd).Val()
			for _, timeAtStr := range timeAtStrs {
				if timeAt, err := strconv.ParseInt(timeAtStr, 10, 64); err == nil {
					_TasksAtFutureList = append(_TasksAtFutureList, &TaskAtFuture{ServiceName: service, TimeAtUnixNs: timeAt})
				}
			}
		}

	}
	sort.Slice(_TasksAtFutureList, func(i, j int) bool {
		return _TasksAtFutureList[i].TimeAtUnixNs < _TasksAtFutureList[j].TimeAtUnixNs
	})
	mut.Lock()
	TasksAtFutureList = _TasksAtFutureList
	mut.Unlock()
	logger.Info().Msg("rpcCallAtTasksLoading completed")
}
func init() {
	go func() {
		rpcCallAtTasksLoad()

		rpcCallAtDispatcher()
	}()

}
