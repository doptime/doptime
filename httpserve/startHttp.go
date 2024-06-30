package httpserve

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/doptime/doptime/config"
	"github.com/doptime/doptime/dlog"
	"github.com/doptime/doptime/permission"
	"github.com/doptime/doptime/rdsdb"
	"github.com/redis/go-redis/v9"
	"github.com/vmihailenco/msgpack/v5"
)

func Debug() {

}

// listten to a port and start http server
func httpStart(path string, port int64) {
	//get item
	router := http.NewServeMux()
	router.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		var (
			result     interface{}
			bs         []byte
			s          string
			ok         bool
			err        error
			httpStatus int = http.StatusOK
			svcCtx     *HttpContext
			rds        *redis.Client
			operation  string
		)
		if CorsChecked(r, w) {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*12000)
		defer cancel()
		if svcCtx, err = NewHttpContext(ctx, r, w); err != nil || svcCtx == nil {
			httpStatus = http.StatusBadRequest
		} else if rds, ok = config.Rds[svcCtx.RedisDataSource]; !ok {
			httpStatus = http.StatusInternalServerError
		} else if operation, err = svcCtx.KeyFieldAtJwt(); err != nil {
			httpStatus = http.StatusInternalServerError
		} else if svcCtx.SUToken != "" && config.Cfg.Settings.SUToken != svcCtx.SUToken {
			httpStatus = http.StatusForbidden
			err = ErrSUNotMatch
		} else if svcCtx.SUToken == "" && !permission.IsPermitted(svcCtx.Key, operation) {
			httpStatus = http.StatusForbidden
			err = ErrOperationNotPermited
		} else if svcCtx.Cmd == "API" {
			result, err = svcCtx.APiHandler()
		} else {

			switch svcCtx.Cmd {
			// all data that appears in the form or body is json format, will be stored in paramIn["JsonPack"]
			// this is used to support 3rd party api
			case "HLEN":
				result, err = rds.HLen(ctx, svcCtx.Key).Result()
			case "LLEN":
				result, err = rds.LLen(ctx, svcCtx.Key).Result()
			case "XLEN":
				result, err = rds.XLen(ctx, svcCtx.Key).Result()
			case "ZCARD":
				result, err = rds.ZCard(ctx, svcCtx.Key).Result()
			case "SCARD":
				result, err = rds.SCard(ctx, svcCtx.Key).Result()

			case "SSCAN":
				var (
					cursor uint64
					count  int64
					keys   []string
					match  string
				)
				result = ""
				if cursor, err = strconv.ParseUint(svcCtx.Req.FormValue("Cursor"), 10, 64); err != nil {
				} else if match = svcCtx.Req.FormValue("Match"); match == "" {
				} else if count, err = strconv.ParseInt(svcCtx.Req.FormValue("Count"), 10, 64); err != nil {
				} else if keys, cursor, err = rds.SScan(context.Background(), svcCtx.Key, cursor, match, count).Result(); err == nil {
					result = map[string]interface{}{"keys": convertKeysToStringBytes(keys), "cursor": cursor}
				}
			case "HSCAN":
				var (
					cursor  uint64
					count   int64
					keys    []string
					match   string
					novalue bool
				)
				hKey := rdsdb.HashKey[string, interface{}](rdsdb.Option.WithKey(svcCtx.Key), rdsdb.Option.WithRds(svcCtx.RedisDataSource))
				result = ""
				if cursor, err = strconv.ParseUint(svcCtx.Req.FormValue("Cursor"), 10, 64); err != nil {
				} else if match = svcCtx.Req.FormValue("Match"); match == "" {
				} else if count, err = strconv.ParseInt(svcCtx.Req.FormValue("Count"), 10, 64); err != nil {
				} else if novalue, err = strconv.ParseBool(svcCtx.Req.FormValue("NOVALUE")); novalue {
					if keys, cursor, err = hKey.HScanNoValues(cursor, match, count); err == nil {
						result = map[string]interface{}{"data": keys, "cursor": cursor}
					}
				} else {
					kvmap, cursorRet, err := hKey.HScan(cursor, match, count)
					if err == nil {
						result = map[string]interface{}{"data": kvmap, "cursor": cursorRet}
					}
				}
			case "ZSCAN":
				var (
					cursor uint64
					count  int64
					keys   []string
					match  string
				)
				result = ""
				if cursor, err = strconv.ParseUint(svcCtx.Req.FormValue("Cursor"), 10, 64); err != nil {
				} else if match = svcCtx.Req.FormValue("Match"); match == "" {
				} else if count, err = strconv.ParseInt(svcCtx.Req.FormValue("Count"), 10, 64); err != nil {
				} else if keys, cursor, err = rds.ZScan(context.Background(), svcCtx.Key, cursor, match, count).Result(); err == nil {
					result = map[string]interface{}{"keys": convertKeysToStringBytes(keys), "cursor": cursor}
				}
			case "LRANGE":
				var (
					start, stop int64 = 0, -1
				)
				if start, err = strconv.ParseInt(svcCtx.Req.FormValue("Start"), 10, 64); err != nil {
					result, err = "", errors.New("parse start error:"+err.Error())
				} else if stop, err = strconv.ParseInt(svcCtx.Req.FormValue("Stop"), 10, 64); err != nil {
					result, err = "", errors.New("parse stop error:"+err.Error())
				} else if result, err = rds.LRange(context.Background(), svcCtx.Key, start, stop).Result(); err == nil {
					result = convertKeysToBytes(result.([]string))
				}
			case "XRANGE":
				var (
					start, stop string
				)
				if start = svcCtx.Req.FormValue("Start"); start == "" {
					result, err = "false", errors.New("no Start")
				} else if stop = svcCtx.Req.FormValue("Stop"); stop == "" {
					result, err = "false", errors.New("no Stop")
				} else {
					result, err = rds.XRange(svcCtx.Ctx, svcCtx.Key, start, stop).Result()
				}
			case "XRANGEN":
				var (
					start, stop string
					count       int64
				)
				if start = svcCtx.Req.FormValue("Start"); start == "" {
					result, err = "false", errors.New("no Start")
				} else if stop = svcCtx.Req.FormValue("Stop"); stop == "" {
					result, err = "false", errors.New("no Stop")
				} else if count, err = strconv.ParseInt(svcCtx.Req.FormValue("Count"), 10, 64); err != nil {
					result, err = "false", errors.New("parse N error:"+err.Error())
				} else {
					result, err = rds.XRangeN(svcCtx.Ctx, svcCtx.Key, start, stop, count).Result()
				}
			case "XREVRANGE":
				var (
					start, stop string
				)
				if start = svcCtx.Req.FormValue("Start"); start == "" {
					result, err = "false", errors.New("no Start")
				} else if stop = svcCtx.Req.FormValue("Stop"); stop == "" {
					result, err = "false", errors.New("no Stop")
				} else {
					result, err = rds.XRevRange(svcCtx.Ctx, svcCtx.Key, start, stop).Result()
				}
			case "XREVRANGEN":
				var (
					start, stop string
					count       int64
				)
				if start = svcCtx.Req.FormValue("Start"); start == "" {
					result, err = "false", errors.New("no Start")
				} else if stop = svcCtx.Req.FormValue("Stop"); stop == "" {
					result, err = "false", errors.New("no Stop")
				} else if count, err = strconv.ParseInt(svcCtx.Req.FormValue("Count"), 10, 64); err != nil {
					result, err = "false", errors.New("parse N error:"+err.Error())
				} else {
					result, err = rds.XRevRangeN(svcCtx.Ctx, svcCtx.Key, start, stop, count).Result()
				}
			case "XREAD":
				var (
					count int64
					block time.Duration
				)
				if count, err = strconv.ParseInt(svcCtx.Req.FormValue("Count"), 10, 64); err != nil {
					result, err = "false", errors.New("parse count error:"+err.Error())
				} else if block, err = time.ParseDuration(svcCtx.Req.FormValue("Block")); err != nil {
					result, err = "false", errors.New("parse block error:"+err.Error())
				} else {
					result, err = rds.XRead(svcCtx.Ctx, &redis.XReadArgs{Streams: []string{svcCtx.Key, svcCtx.Req.FormValue("ID")}, Count: count, Block: block}).Result()
				}

			case "GET":
				db := rdsdb.StringKey[string, interface{}](rdsdb.Option.WithKey(svcCtx.Key), rdsdb.Option.WithRds(svcCtx.RedisDataSource))
				result, err = db.Get(svcCtx.Field)
			case "HGET":
				db := rdsdb.HashKey[string, interface{}](rdsdb.Option.WithKey(svcCtx.Key), rdsdb.Option.WithRds(svcCtx.RedisDataSource))
				result, err = db.HGet(svcCtx.Field)
			case "HGETALL":
				db := rdsdb.HashKey[string, interface{}](rdsdb.Option.WithKey(svcCtx.Key), rdsdb.Option.WithRds(svcCtx.RedisDataSource))
				result, err = db.HGetAll()
			case "HMGET":
				db := rdsdb.HashKey[string, interface{}](rdsdb.Option.WithKey(svcCtx.Key), rdsdb.Option.WithRds(svcCtx.RedisDataSource))
				result, err = db.HMGET(strings.Split(svcCtx.Field, ",")...)
			case "HKEYS":
				db := rdsdb.HashKey[string, interface{}](rdsdb.Option.WithKey(svcCtx.Key), rdsdb.Option.WithRds(svcCtx.RedisDataSource))
				result, err = db.HKeys()
			case "HEXISTS":
				db := rdsdb.HashKey[string, interface{}](rdsdb.Option.WithKey(svcCtx.Key), rdsdb.Option.WithRds(svcCtx.RedisDataSource))
				result, err = db.HExists(svcCtx.Field)
			case "HRANDFIELD":
				db := rdsdb.HashKey[string, interface{}](rdsdb.Option.WithKey(svcCtx.Key), rdsdb.Option.WithRds(svcCtx.RedisDataSource))
				var count int
				if count, err = strconv.Atoi(svcCtx.Req.FormValue("Count")); err != nil {
					result, err = "", errors.New("parse count error:"+err.Error())
				} else {
					result, err = db.HRandField(count)
				}
			case "HVALS":
				db := rdsdb.HashKey[string, interface{}](rdsdb.Option.WithKey(svcCtx.Key), rdsdb.Option.WithRds(svcCtx.RedisDataSource))
				result, err = db.HVals()
			case "SISMEMBER":
				db := rdsdb.SetKey[string, interface{}](rdsdb.Option.WithKey(svcCtx.Key), rdsdb.Option.WithRds(svcCtx.RedisDataSource))
				result, err = db.SIsMember(svcCtx.Req.FormValue("Member"))
			case "TIME":
				db := rdsdb.NonKey[string, interface{}](rdsdb.Option.WithKey(svcCtx.Key), rdsdb.Option.WithRds(svcCtx.RedisDataSource))
				result = ""
				if tm, err := db.Time(); err == nil {
					result = tm.UnixMilli()
				}
			case "ZRANGE":
				db := rdsdb.ZSetKey[string, interface{}](rdsdb.Option.WithKey(svcCtx.Key), rdsdb.Option.WithRds(svcCtx.RedisDataSource))
				var (
					start, stop int64 = 0, -1
				)
				result = ""
				if start, err = strconv.ParseInt(svcCtx.Req.FormValue("Start"), 10, 64); err != nil {
				} else if stop, err = strconv.ParseInt(svcCtx.Req.FormValue("Stop"), 10, 64); err != nil {
				} else if svcCtx.Req.FormValue("WITHSCORES") == "true" {
					// ZRANGE key start stop [WITHSCORES==true]
					var scores []float64
					if result, scores, err = db.ZRangeWithScores(start, stop); err == nil {
						result = map[string]interface{}{"members": result, "scores": scores}
					}
				} else {
					// ZRANGE key start stop [WITHSCORES==false]
					result, err = db.ZRange(start, stop)
				}
			case "ZRANGEBYSCORE":
				db := rdsdb.ZSetKey[string, interface{}](rdsdb.Option.WithKey(svcCtx.Key), rdsdb.Option.WithRds(svcCtx.RedisDataSource))
				var (
					offset, count int64 = 0, -1
					scores        []float64
				)
				result = ""
				if Min := svcCtx.Req.FormValue("Min"); Min == "" {
					result, err = "false", errors.New("no Min")
				} else if Max := svcCtx.Req.FormValue("Max"); Max == "" {
					result, err = "false", errors.New("no Max")
				} else if offset, err = strconv.ParseInt(svcCtx.Req.FormValue("Offset"), 10, 64); err != nil {
					result, err = "false", errors.New("parse offset error:"+err.Error())
				} else if count, err = strconv.ParseInt(svcCtx.Req.FormValue("Count"), 10, 64); err != nil {
					result, err = "false", errors.New("parse count error:"+err.Error())
				} else if svcCtx.Req.FormValue("WITHSCORES") == "true" {
					//ZRANGEBYSCORE key min max [WITHSCORES==true]
					if result, scores, err = db.ZRangeByScoreWithScores(&redis.ZRangeBy{Min: Min, Max: Max, Offset: offset, Count: count}); err == nil {
						result = map[string]interface{}{"members": result, "scores": scores}
					}
				} else {
					//ZRANGEBYSCORE key min max [WITHSCORES==false]
					result, err = db.ZRangeByScore(&redis.ZRangeBy{Min: Min, Max: Max, Offset: offset, Count: count})
				}
			case "ZREVRANGE":
				db := rdsdb.ZSetKey[string, interface{}](rdsdb.Option.WithKey(svcCtx.Key), rdsdb.Option.WithRds(svcCtx.RedisDataSource))
				var (
					start, stop int64 = 0, -1
				)
				result = ""
				if start, err = strconv.ParseInt(svcCtx.Req.FormValue("Start"), 10, 64); err != nil {
				} else if stop, err = strconv.ParseInt(svcCtx.Req.FormValue("Stop"), 10, 64); err != nil {
				} else if svcCtx.Req.FormValue("WITHSCORES") == "true" {
					// ZREVRANGE key start stop [WITHSCORES==true]
					cmd := db.Rds.ZRevRangeWithScores(context.Background(), db.Key, start, stop)
					if rlts, err := cmd.Result(); err == nil {
						var memberScoreSlice []interface{}
						for _, rlt := range rlts {
							memberScoreSlice = append(memberScoreSlice, rlt.Member)
							memberScoreSlice = append(memberScoreSlice, rlt.Score)
						}
						result = memberScoreSlice
					}
				} else {
					// ZREVRANGE key start stop [WITHSCORES==false]
					result, err = db.ZRevRange(start, stop)
				}
			case "ZREVRANGEBYSCORE":
				db := rdsdb.ZSetKey[string, interface{}](rdsdb.Option.WithKey(svcCtx.Key), rdsdb.Option.WithRds(svcCtx.RedisDataSource))
				var (
					offset, count int64 = 0, -1
				)
				if Min, Max := svcCtx.Req.FormValue("Min"), svcCtx.Req.FormValue("Max"); Min == "" || Max == "" {
					result, err = "", errors.New("no Min or Max")
				} else if offset, err = strconv.ParseInt(svcCtx.Req.FormValue("Offset"), 10, 64); err != nil {
					result, err = "", errors.New("parse offset error:"+err.Error())
				} else if count, err = strconv.ParseInt(svcCtx.Req.FormValue("Count"), 10, 64); err != nil {
					result, err = "", errors.New("parse count error:"+err.Error())
				} else if svcCtx.Req.FormValue("WITHSCORES") == "true" {
					cmd := db.Rds.ZRevRangeByScoreWithScores(context.Background(), db.Key, &redis.ZRangeBy{Min: Min, Max: Max, Offset: offset, Count: count})
					if rlts, err := cmd.Result(); err == nil {
						var memberScoreSlice []interface{}
						for _, rlt := range rlts {
							memberScoreSlice = append(memberScoreSlice, rlt.Member)
							memberScoreSlice = append(memberScoreSlice, rlt.Score)
						}
						result, err = memberScoreSlice, nil
					}
				} else {
					//ZREVRANGEBYSCORE key max min [WITHSCORES==false]
					result, err = db.ZRevRangeByScore(&redis.ZRangeBy{Min: Min, Max: Max, Offset: offset, Count: count})
				}
			case "ZRANK":
				db := rdsdb.ZSetKey[string, interface{}](rdsdb.Option.WithKey(svcCtx.Key), rdsdb.Option.WithRds(svcCtx.RedisDataSource))
				result, err = db.ZRank(svcCtx.Req.FormValue("Member"))
			case "ZCOUNT":
				db := rdsdb.ZSetKey[string, interface{}](rdsdb.Option.WithKey(svcCtx.Key), rdsdb.Option.WithRds(svcCtx.RedisDataSource))
				result, err = db.ZCount(svcCtx.Req.FormValue("Min"), svcCtx.Req.FormValue("Max"))
			case "ZSCORE":
				db := rdsdb.ZSetKey[string, interface{}](rdsdb.Option.WithKey(svcCtx.Key), rdsdb.Option.WithRds(svcCtx.RedisDataSource))
				result, err = db.ZScore(svcCtx.Req.FormValue("Member"))
			case "SCAN":
				db := rdsdb.NonKey[string, interface{}](rdsdb.Option.WithKey(svcCtx.Key), rdsdb.Option.WithRds(svcCtx.RedisDataSource))
				var (
					cursor uint64
					count  int64
					keys   []string
					match  string
				)
				result = ""
				if cursor, err = strconv.ParseUint(svcCtx.Req.FormValue("Cursor"), 10, 64); err != nil {
				} else if match = svcCtx.Req.FormValue("Match"); match == "" {
				} else if count, err = strconv.ParseInt(svcCtx.Req.FormValue("Count"), 10, 64); err != nil {
				} else if keys, cursor, err = db.Scan(cursor, match, count); err != nil {
				} else {
					result, err = json.Marshal(map[string]interface{}{"keys": keys, "cursor": cursor})
				}
			case "LINDEX":
				db := rdsdb.ListKey[string, interface{}](rdsdb.Option.WithKey(svcCtx.Key), rdsdb.Option.WithRds(svcCtx.RedisDataSource))
				var index int64
				if index, err = strconv.ParseInt(svcCtx.Req.FormValue("Index"), 10, 64); err != nil {
					result, err = "", errors.New("parse index error:"+err.Error())
				} else {
					result, err = db.LIndex(index)
				}
			case "LPOP":
				db := rdsdb.ListKey[string, interface{}](rdsdb.Option.WithKey(svcCtx.Key), rdsdb.Option.WithRds(svcCtx.RedisDataSource))
				result, err = db.LPop()
			case "LPUSH":
				result = "false"
				if bs, err = svcCtx.MsgpackBody(); err != nil {
				} else if err = rdsdb.CheckDataSchema(svcCtx.Key, bs); err != nil {
				} else if err = rds.LPush(svcCtx.Ctx, svcCtx.Key, bs).Err(); err == nil {
					result = "true"
				}
			case "LREM":
				var count int64
				if count, err = strconv.ParseInt(svcCtx.Req.FormValue("Count"), 10, 64); err != nil {
					result, err = "false", errors.New("parse count error:"+err.Error())
				} else if bs, err = svcCtx.MsgpackBody(); err != nil {
				} else if err = rds.LRem(svcCtx.Ctx, svcCtx.Key, count, bs).Err(); err == nil {
					result = "true"
				}
			case "LSET":
				result = "false"
				var index int64
				if index, err = strconv.ParseInt(svcCtx.Req.FormValue("Index"), 10, 64); err != nil {
					err = errors.New("parse index error:" + err.Error())
				} else if bs, err = svcCtx.MsgpackBody(); err != nil {
				} else if err = rds.LSet(svcCtx.Ctx, svcCtx.Key, index, bs).Err(); err == nil {
					result = "true"
				}
			case "LTRIM":
				var start, stop int64
				if start, err = strconv.ParseInt(svcCtx.Req.FormValue("Start"), 10, 64); err != nil {
					result, err = "false", errors.New("parse start error:"+err.Error())
				} else if stop, err = strconv.ParseInt(svcCtx.Req.FormValue("Stop"), 10, 64); err != nil {
					result, err = "false", errors.New("parse stop error:"+err.Error())
				} else if err = rds.LTrim(svcCtx.Ctx, svcCtx.Key, start, stop).Err(); err == nil {
					result = "true"
				}
			case "RPOP":
				db := rdsdb.ListKey[string, interface{}](rdsdb.Option.WithKey(svcCtx.Key), rdsdb.Option.WithRds(svcCtx.RedisDataSource))
				result, err = db.RPop()
			case "RPUSH":
				result = "false"
				if bs, err = svcCtx.MsgpackBody(); err != nil {
				} else if err = rdsdb.CheckDataSchema(svcCtx.Key, bs); err != nil {
				} else if err = rds.RPush(svcCtx.Ctx, svcCtx.Key, bs).Err(); err == nil {
					result = "true"
				}

			case "RPUSHX":
				result = "false"
				if bs, err = svcCtx.MsgpackBody(); err != nil {
				} else if err = rdsdb.CheckDataSchema(svcCtx.Key, bs); err != nil {
				} else if err = rds.RPushX(svcCtx.Ctx, svcCtx.Key, bs).Err(); err == nil {
					result = "true"
				}

			case "ZADD":
				db := rdsdb.ZSetKey[string, interface{}](rdsdb.Option.WithKey(svcCtx.Key), rdsdb.Option.WithRds(svcCtx.RedisDataSource))
				var Score float64
				var obj interface{}
				if Score, err = strconv.ParseFloat(svcCtx.Req.FormValue("Score"), 64); err != nil {
					result, err = "false", errors.New("parameter Score shoule be float")
				} else if MsgPack := svcCtx.MsgpackBodyBytes(); len(MsgPack) == 0 {
					result, err = "false", errors.New("missing MsgPack content")
				} else if result, err = "true", msgpack.Unmarshal(MsgPack, &obj); err != nil {
					result = "false"
				} else if err = db.ZAdd(redis.Z{Score: Score, Member: obj}); err != nil {
					result = "false"
				}

			case "SET":
				result = "false"
				if svcCtx.Key == "" || svcCtx.Field == "" {
					err = ErrEmptyKeyOrField
				} else if bytes, err := svcCtx.MsgpackBody(); err != nil {
				} else if err = rdsdb.CheckDataSchema(svcCtx.Key, bs); err != nil {
				} else if rds.Set(svcCtx.Ctx, svcCtx.Key+":"+svcCtx.Field, bytes, 0).Err() == nil {
					result = "true"
				}

			case "HSET":
				result = "false"
				//error if empty Key or Field
				if svcCtx.Key == "" || svcCtx.Field == "" {
					err = ErrEmptyKeyOrField
				} else if bs, err = svcCtx.MsgpackBody(); err != nil {
				} else if err = rdsdb.CheckDataSchema(svcCtx.Key, bs); err != nil {
				} else if err = rds.HSet(svcCtx.Ctx, svcCtx.Key, svcCtx.Field, bs).Err(); err == nil {
					result = "true"
				}
			case "HDEL":
				result = "false"
				//error if empty Key or Field
				if svcCtx.Field == "" {
					err = ErrEmptyKeyOrField
				} else if err = rds.HDel(svcCtx.Ctx, svcCtx.Key, svcCtx.Field).Err(); err == nil {
					result = "true"
				}
			case "DEL":
				result = "false"
				if err = rds.HDel(svcCtx.Ctx, svcCtx.Key, "del").Err(); err == nil {
					result = "true"
				}
			case "ZREM":
				result = "false"
				var MemberStr = strings.Split(svcCtx.Req.FormValue("Member"), ",")
				//convert Member to []interface{}
				var Member = make([]interface{}, len(MemberStr))
				for i, v := range MemberStr {
					Member[i] = v
				}

				if len(Member) == 0 {
					err = errors.New("no Member")
				} else if err = rds.ZRem(svcCtx.Ctx, svcCtx.Key, Member...).Err(); err == nil {
					result = "true"
				}

			case "ZREMRANGEBYSCORE":
				result = "false"
				var Min, Max = svcCtx.Req.FormValue("Min"), svcCtx.Req.FormValue("Max")
				if Min == "" || Max == "" {
					err = errors.New("no Min or Max")
				} else if err = rds.ZRemRangeByScore(svcCtx.Ctx, svcCtx.Key, Min, Max).Err(); err == nil {
					result = "true"
				}
			case "TYPE":
				result, err = rds.Type(svcCtx.Ctx, svcCtx.Key).Result()
			case "EXPIRE":
				var seconds int64
				if seconds, err = strconv.ParseInt(svcCtx.Req.FormValue("Seconds"), 10, 64); err != nil {
					result, err = "false", errors.New("parse seconds error:"+err.Error())
				} else if err = rds.Expire(svcCtx.Ctx, svcCtx.Key, time.Duration(seconds)*time.Second).Err(); err == nil {
					result = "true"
				}
			case "EXPIREAT":
				var timestamp int64
				if timestamp, err = strconv.ParseInt(svcCtx.Req.FormValue("Timestamp"), 10, 64); err != nil {
					result, err = "false", errors.New("parse seconds error:"+err.Error())
				} else if err = rds.ExpireAt(svcCtx.Ctx, svcCtx.Key, time.Unix(timestamp, 0)).Err(); err == nil {
					result = "true"
				}
			case "PERSIST":
				if err = rds.Persist(svcCtx.Ctx, svcCtx.Key).Err(); err == nil {
					result = "true"
				}
			case "TTL":
				result, err = rds.TTL(svcCtx.Ctx, svcCtx.Key).Result()
			case "PTTL":
				result, err = rds.PTTL(svcCtx.Ctx, svcCtx.Key).Result()
			case "RENAME":
				if newKey := svcCtx.Req.FormValue("NewKey"); newKey == "" {
					result, err = "false", errors.New("no NewKey")
				} else if err = rds.Rename(svcCtx.Ctx, svcCtx.Key, newKey).Err(); err == nil {
					result = "true"
				}
			case "RENAMEX":
				if newKey := svcCtx.Req.FormValue("NewKey"); newKey == "" {
					result, err = "false", errors.New("no NewKey")
				} else if err = rds.RenameNX(svcCtx.Ctx, svcCtx.Key, newKey).Err(); err == nil {
					result = "true"
				}
			case "EXISTS":
				result, err = rds.Exists(svcCtx.Ctx, svcCtx.Key).Result()
			case "KEYS":
				result, err = rds.Keys(svcCtx.Ctx, svcCtx.Key).Result()
			case "ZINCRBY":
				var (
					incr   float64
					member string
				)
				if member = svcCtx.Req.FormValue("Member"); member == "" {
					result, err = "false", errors.New("no Member")
				} else if incr, err = strconv.ParseFloat(svcCtx.Req.FormValue("Incr"), 64); err != nil {
					result, err = "false", errors.New("parse Incr error:"+err.Error())
				} else if err = rds.ZIncrBy(svcCtx.Ctx, svcCtx.Key, incr, member).Err(); err == nil {
					result = "true"
				}
			case "HINCRBY":
				var (
					incr int64
				)
				if incr, err = strconv.ParseInt(svcCtx.Req.FormValue("Incr"), 10, 64); err != nil {
					result, err = "false", errors.New("parse Incr error:"+err.Error())
				} else if err = rds.HIncrBy(svcCtx.Ctx, svcCtx.Key, svcCtx.Field, incr).Err(); err == nil {
					result = "true"
				}
			case "HINCRBYFLOAT":
				var (
					incr float64
				)
				if incr, err = strconv.ParseFloat(svcCtx.Req.FormValue("Incr"), 64); err != nil {
					result, err = "false", errors.New("parse Incr error:"+err.Error())
				} else if err = rds.HIncrByFloat(svcCtx.Ctx, svcCtx.Key, svcCtx.Field, incr).Err(); err == nil {
					result = "true"
				}
			case "XADD":
				var (
					id  string
					obj interface{}
				)
				if id = svcCtx.Req.FormValue("ID"); id == "" {
					result, err = "false", errors.New("no ID")
				} else if MsgPack := svcCtx.MsgpackBodyBytes(); len(MsgPack) == 0 {
					result, err = "false", errors.New("missing MsgPack content")
				} else if result, err = "true", msgpack.Unmarshal(MsgPack, &obj); err != nil {
					result = "false"
				} else if err = rds.XAdd(svcCtx.Ctx, &redis.XAddArgs{Stream: svcCtx.Key, Values: map[string]interface{}{id: obj}}).Err(); err != nil {
					result = "false"
				}
			case "XDEL":
				var (
					id string
				)
				if id = svcCtx.Req.FormValue("ID"); id == "" {
					result, err = "false", errors.New("no ID")
				} else if err = rds.XDel(svcCtx.Ctx, svcCtx.Key, id).Err(); err != nil {
					result = "false"
				} else {
					result = "true"
				}

			//case default
			default:
				result, err = nil, ErrBadCommand
			}
		}

		if len(config.Cfg.Http.CORES) > 0 {
			w.Header().Set("Access-Control-Allow-Origin", config.Cfg.Http.CORES)
		}

		if err == nil {
			if svcCtx.ResponseContentType == "application/msgpack" {
				if bs, err = msgpack.Marshal(result); err != nil {
					httpStatus = http.StatusInternalServerError
				}
			} else if bs, ok = result.([]byte); ok {
			} else if s, ok = result.(string); ok {
				bs = []byte(s)
			} else {
				if bs, err = json.Marshal(result); err == nil {
					//json Compact b
					var dst *bytes.Buffer = bytes.NewBuffer([]byte{})
					if err = json.Compact(dst, bs); err == nil {
						bs = dst.Bytes()
					}
				}
			}
		}
		//this err may be from json.marshal, so don't move it to the above else if
		if err != nil {
			if bs = []byte(err.Error()); bytes.Contains(bs, []byte("JWT")) {
				httpStatus = http.StatusUnauthorized
			} else if httpStatus == http.StatusOK {
				// this if is needed, because  httpStatus may have already setted as StatusBadRequest
				httpStatus = http.StatusInternalServerError
			}
		}

		//set Content-Type
		if svcCtx != nil && len(svcCtx.ResponseContentType) > 0 {
			svcCtx.Rsb.Header().Set("Content-Type", svcCtx.ResponseContentType)
		}
		w.WriteHeader(httpStatus)
		w.Write(bs)
	})

	server := &http.Server{
		Addr:              ":" + strconv.FormatInt(port, 10),
		Handler:           router,
		ReadTimeout:       50 * time.Second,
		ReadHeaderTimeout: 50 * time.Second,
		WriteTimeout:      50 * time.Second, //10ms Redundant time
		IdleTimeout:       15 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		dlog.Error().Err(err).Msg("http server ListenAndServe error")
		return
	}
	dlog.Info().Any("port", port).Any("path", path).Msg("doptime http server started!")
}

func init() {
	//wait 1s, till the permission table is loaded
	for i := 0; !permission.ConfigurationLoaded && i < 100; i++ {
		time.Sleep(time.Millisecond * 10)
	}
	//wait, till all the apis are loaded
	dlog.Info().Any("port", config.Cfg.Http.Port).Any("path", config.Cfg.Http.Path).Msg("doptime http server is starting")
	go httpStart(config.Cfg.Http.Path, config.Cfg.Http.Port)
}
