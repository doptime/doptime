package libapi

import (
	"context"
	"strconv"
	"time"

	"github.com/doptime/doptime/api"
	"github.com/doptime/doptime/config"
	"github.com/doptime/doptime/dlog"
	"github.com/redis/go-redis/v9"
)

type InLockKey struct {
	Key        string
	DurationMs int64
}

var removeCounter int64 = 90

var ApiLockKey = api.Api(func(req *InLockKey) (ok bool, err error) {
	var (
		now    int64 = time.Now().UnixMilli()
		timeAt int64 = now + req.DurationMs
		score  float64
		rds    *redis.Client
	)
	if rds, ok = config.Rds.Get("default"); !ok {
		dlog.Error().Err(err).Str("DataSource name not defined in enviroment while calling ApiLockKey", "default").Send()
		return false, err
	}
	if score, err = rds.ZScore(context.Background(), "KeyLocker", req.Key).Result(); err != nil {
		if err != redis.Nil {
			return false, err
		} else {
			ok = true
		}
	} else if score < float64(now) {
		ok = true
	}
	//update only when key not exists, or expired
	if ok {
		rds.ZAdd(context.Background(), "KeyLocker", redis.Z{Score: float64(timeAt), Member: req.Key})
	}

	//auto remove expired keys
	if removeCounter++; removeCounter > 100 {
		removeCounter = 0
		go rds.ZRemRangeByScore(context.Background(), "KeyLocker", "0", strconv.FormatInt(now, 10))
	}
	return ok, nil
})
