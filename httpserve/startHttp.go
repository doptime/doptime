package httpserve

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/doptime/config/cfghttp"
	"github.com/doptime/config/cfgredis"
	"github.com/doptime/doptime/httpserve/httpapi"
	"github.com/doptime/logger"
	"github.com/doptime/redisdb"
	"github.com/redis/go-redis/v9"
	"github.com/vmihailenco/msgpack/v5"
)

func Debug() {

}

var hashKey = redisdb.NewHashKey[string, interface{}](redisdb.Opt.Key("_"))
var zsetKey = redisdb.NewZSetKey[string, interface{}](redisdb.Opt.Key("_"))
var listKey = redisdb.NewListKey[interface{}](redisdb.Opt.Key("_"))
var setKey = redisdb.NewSetKey[string, interface{}](redisdb.Opt.Key("_"))
var stringKey = redisdb.NewStringKey[string, interface{}](redisdb.Opt.Key("_"))

// get item
var httpRoter = http.NewServeMux()
var mu sync.Mutex // 创建一个互斥锁，确保多路复用器的操作是线程安全的
// 动态添加路由的函数
func AddRoute(path string, handlerFunc http.HandlerFunc) {
	mu.Lock() // 锁定，确保修改路由时不会发生并发冲突
	defer mu.Unlock()
	httpRoter.HandleFunc(path, handlerFunc)
}

// listten to a port and start http server
func httpStart(path string, port int64) {
	httpRoter.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		var (
			result       interface{}
			bs           []byte
			ok           bool
			err          error
			httpStatus   int = http.StatusOK
			svcCtx       *DoptimeReqCtx
			rds          *redis.Client
			hkey         *redisdb.HashKey[string, interface{}]
			skey         *redisdb.SetKey[string, interface{}]
			zkey         *redisdb.ZSetKey[string, interface{}]
			lKey         *redisdb.ListKey[interface{}]
			strKey       *redisdb.StringKey[string, interface{}]
			operation, s string = "", ""
			//load redis datasource value from form
			RedisDataSource            = r.FormValue("ds")
			ResponseContentType string = "application/json"
		)
		//default response content type: application/json
		if rt := r.FormValue("rt"); rt != "" {
			ResponseContentType = rt
		}
		//default RedisDataSource is "default"
		if RedisDataSource == "" {
			RedisDataSource = "default"
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*12000)
		defer cancel()

		if CorsChecked(r, w) {
			goto responseHttp
		}

		//should be valid http request. either data operation or api operation
		if svcCtx, err = NewHttpContext(ctx, r, w); err != nil || svcCtx == nil {
			httpStatus = http.StatusBadRequest
			goto responseHttp
		}
		//there should be a valid cmd & key if needed
		if err = svcCtx.EnsureKeyFieldIsValid(); err != nil {
			httpStatus = http.StatusBadRequest
			goto responseHttp
		}

		//RedisDataSource should be valid
		if rds, ok = cfgredis.Servers.Get(RedisDataSource); !ok {
			httpStatus = http.StatusInternalServerError
			goto responseHttp
		}

		//@Tag in key or field should be replaced by value in Jwt
		if operation, err = svcCtx.UpdateKeyFieldWithJwtClaims(); err != nil {
			httpStatus = http.StatusInternalServerError
			goto responseHttp
		}
		//auth check
		if DataOpSuperUserToken := r.FormValue("su"); DataOpSuperUserToken != "" && cfghttp.SUToken != DataOpSuperUserToken {
			httpStatus = http.StatusForbidden
			err = ErrSUNotMatch
			goto responseHttp
		} else if DataOpSuperUserToken == "" && !IsPermitted(operation) {
			httpStatus = http.StatusForbidden
			err = ErrOperationNotPermited
			goto responseHttp
		}

		// case calling "API"
		if _, isDataKey := DataCmdRequireKey[svcCtx.Cmd]; !isDataKey {
			var (
				paramIn     map[string]interface{} = map[string]interface{}{}
				ServiceName string                 = svcCtx.Key
				_api        httpapi.ApiInterface
				ok          bool
			)

			svcCtx.ParamIn, err = io.ReadAll(r.Body)
			//marshal body to map[string]interface{}
			if contentType := r.Header.Get("Content-Type"); len(svcCtx.ParamIn) > 0 && len(contentType) > 0 && err == nil {
				switch contentType {
				case "application/octet-stream":
					err = msgpack.Unmarshal(svcCtx.ParamIn, &paramIn)
					if err != nil {
						var interfaceIn interface{}
						if err = msgpack.Unmarshal(svcCtx.ParamIn, &interfaceIn); err == nil {
							paramIn["_msgpack-nonstruct"] = svcCtx.ParamIn
						}
					}
				case "application/json":
					err = json.Unmarshal(svcCtx.ParamIn, &paramIn)
					if err != nil {
						var interfaceIn interface{}
						if err = json.Unmarshal(svcCtx.ParamIn, &interfaceIn); err == nil {
							paramIn["_jsonpack-nonstruct"] = svcCtx.ParamIn
						}
					}
				}
				if err != nil {
					goto responseHttp
				}
			}

			if _api, ok = httpapi.GetApiByName(ServiceName); !ok {
				result, err = nil, fmt.Errorf("err no such api")
				goto responseHttp
			}
			//convert query fields to JsonPack. but ignore K field(api name )
			r.ParseForm()
			svcCtx.MergeFormParam(r.Form, paramIn)

			_api.MergeHeaderParam(r, paramIn)

			//prevent forged jwt field: remove nay field that starts with "Jwt" in paramIn
			svcCtx.MergeJwtParam(paramIn)

			result, err = _api.CallByMap(paramIn)
			goto responseHttp
		}

		// data operation
		switch svcCtx.Cmd {
		// all data that appears in the form or body is json format, will be stored in paramIn["JsonPack"]
		// this is used to support 3rd party api
		case HLEN:
			result, err = rds.HLen(ctx, svcCtx.Key).Result()
		case LLEN:
			result, err = rds.LLen(ctx, svcCtx.Key).Result()
		case XLEN:
			result, err = rds.XLen(ctx, svcCtx.Key).Result()
		case ZCARD:
			result, err = rds.ZCard(ctx, svcCtx.Key).Result()
		case SCARD:
			result, err = rds.SCard(ctx, svcCtx.Key).Result()

		case SSCAN:
			var (
				cursor uint64
				count  int64
				values []interface{}
				match  string
			)
			result = ""
			skey, _, err = SetCtxWitchValueSchemaChecked(svcCtx.Key, RedisDataSource, nil)
			if err != nil {
			} else if cursor, err = strconv.ParseUint(r.FormValue("Cursor"), 10, 64); err != nil {
			} else if match = r.FormValue("Match"); match == "" {
			} else if count, err = strconv.ParseInt(r.FormValue("Count"), 10, 64); err != nil {
			} else if values, cursor, err = skey.SScan(cursor, match, count); err == nil {
				result = map[string]interface{}{"values": values, "cursor": cursor}
			}
		case HSCAN:
			var (
				cursor    uint64
				count     int64
				keys      []string
				match     string
				values    []interface{}
				cursorRet uint64
			)

			result = ""
			if hkey, _, err = HashCtxWitchValueSchemaChecked(svcCtx.Key, RedisDataSource, nil); err != nil {
			} else if cursor, err = strconv.ParseUint(r.FormValue("Cursor"), 10, 64); err != nil {
			} else if match = r.FormValue("Match"); match == "" {
			} else if count, err = strconv.ParseInt(r.FormValue("Count"), 10, 64); err != nil {
			} else if novalue := r.FormValue("NOVALUE"); novalue == "true" {
				if keys, cursor, err = hkey.HScanNoValues(cursor, match, count); err == nil {
					result = map[string]interface{}{"keys": keys, "cursor": cursor}
				}
			} else {
				keys, values, cursorRet, err = hkey.HScan(cursor, match, count)
				if err == nil {
					result = map[string]interface{}{"keys": keys, "values": values, "cursor": cursorRet}
				}
			}
		case ZSCAN:
			var (
				cursor uint64
				count  int64
				values []interface{}
				match  string
			)
			if zkey, _, err = ZSetCtxWitchValueSchemaChecked(svcCtx.Key, RedisDataSource, nil); err != nil {
			} else if cursor, err = strconv.ParseUint(r.FormValue("Cursor"), 10, 64); err != nil {
			} else if match = r.FormValue("Match"); match == "" {
			} else if count, err = strconv.ParseInt(r.FormValue("Count"), 10, 64); err != nil {
			} else if values, cursor, err = zkey.ZScan(cursor, match, count); err == nil {
				result = map[string]interface{}{"values": values, "cursor": cursor}
			}
		case LRANGE:
			var (
				start, stop int64 = 0, -1
				hkey        *redisdb.ListKey[interface{}]
			)
			hkey, _, err = ListCtxWitchValueSchemaChecked(svcCtx.Key, RedisDataSource, nil)
			if err != nil {
				return
			}
			if start, err = strconv.ParseInt(r.FormValue("Start"), 10, 64); err != nil {
				result, err = "", errors.New("parse start error:"+err.Error())
			} else if stop, err = strconv.ParseInt(r.FormValue("Stop"), 10, 64); err != nil {
				result, err = "", errors.New("parse stop error:"+err.Error())
			} else {
				result, err = hkey.LRange(start, stop)
			}
		case XRANGE:
			var (
				start, stop string
			)
			if start = r.FormValue("Start"); start == "" {
				result, err = "false", errors.New("no Start")
			} else if stop = r.FormValue("Stop"); stop == "" {
				result, err = "false", errors.New("no Stop")
			} else {
				result, err = rds.XRange(svcCtx.Ctx, svcCtx.Key, start, stop).Result()
			}
		case XRANGEN:
			var (
				start, stop string
				count       int64
			)
			if start = r.FormValue("Start"); start == "" {
				result, err = "false", errors.New("no Start")
			} else if stop = r.FormValue("Stop"); stop == "" {
				result, err = "false", errors.New("no Stop")
			} else if count, err = strconv.ParseInt(r.FormValue("Count"), 10, 64); err != nil {
				result, err = "false", errors.New("parse N error:"+err.Error())
			} else {
				result, err = rds.XRangeN(svcCtx.Ctx, svcCtx.Key, start, stop, count).Result()
			}
		case XREVRANGE:
			var (
				start, stop string
			)
			if start = r.FormValue("Start"); start == "" {
				result, err = "false", errors.New("no Start")
			} else if stop = r.FormValue("Stop"); stop == "" {
				result, err = "false", errors.New("no Stop")
			} else {
				result, err = rds.XRevRange(svcCtx.Ctx, svcCtx.Key, start, stop).Result()
			}
		case XREVRANGEN:
			var (
				start, stop string
				count       int64
			)
			if start = r.FormValue("Start"); start == "" {
				result, err = "false", errors.New("no Start")
			} else if stop = r.FormValue("Stop"); stop == "" {
				result, err = "false", errors.New("no Stop")
			} else if count, err = strconv.ParseInt(r.FormValue("Count"), 10, 64); err != nil {
				result, err = "false", errors.New("parse N error:"+err.Error())
			} else {
				result, err = rds.XRevRangeN(svcCtx.Ctx, svcCtx.Key, start, stop, count).Result()
			}
		case XREAD:
			var (
				count int64
				block time.Duration
			)
			if count, err = strconv.ParseInt(r.FormValue("Count"), 10, 64); err != nil {
				result, err = "false", errors.New("parse count error:"+err.Error())
			} else if block, err = time.ParseDuration(r.FormValue("Block")); err != nil {
				result, err = "false", errors.New("parse block error:"+err.Error())
			} else {
				result, err = rds.XRead(svcCtx.Ctx, &redis.XReadArgs{Streams: []string{svcCtx.Key, r.FormValue("ID")}, Count: count, Block: block}).Result()
			}
		case GET:
			if strKey, _, err = StringCtxWitchValueSchemaChecked(svcCtx.Key, RedisDataSource, nil); err != nil {
			} else {
				result, err = strKey.Get(svcCtx.Field)
			}
		case HGET:
			if hkey, _, err = HashCtxWitchValueSchemaChecked(svcCtx.Key, RedisDataSource, nil); err != nil {
			} else {
				result, err = hkey.HGet(svcCtx.Field)
			}
		case HGETALL:
			if hkey, _, err = HashCtxWitchValueSchemaChecked(svcCtx.Key, RedisDataSource, nil); err != nil {
			} else {
				result, err = hkey.HGetAll()
			}
		case HMGET:
			if hkey, result, err = HashCtxWitchValueSchemaChecked(svcCtx.Key, RedisDataSource, nil); err != nil {
			} else {
				//convert strings.Split(svcCtx.Field, ",") to types of []interface{}
				var fields []interface{} = sliceToInterface(strings.Split(svcCtx.Field, ","))
				result, err = hkey.HMGET(fields...)
			}

		case HSET:
			result = 0
			if hkey, result, err = HashCtxWitchValueSchemaChecked(svcCtx.Key, RedisDataSource, svcCtx.MsgpackBody(r, true)); err != nil {
			} else {
				result, err = hkey.HSet(svcCtx.Field, result)
			}

		case HDEL:
			result, err = rds.HDel(svcCtx.Ctx, svcCtx.Key, svcCtx.Field).Result()
		case HKEYS:
			if hkey, _, err = HashCtxWitchValueSchemaChecked(svcCtx.Key, RedisDataSource, nil); err != nil {
			} else {
				result, err = hkey.HKeys()
			}
		case HEXISTS:
			result = false
			if hkey, _, err = HashCtxWitchValueSchemaChecked(svcCtx.Key, RedisDataSource, nil); err != nil {
			} else {
				result, err = hkey.HExists(svcCtx.Field)
			}
		case HRANDFIELD:
			var count int
			if hkey, _, err = HashCtxWitchValueSchemaChecked(svcCtx.Key, RedisDataSource, nil); err != nil {
			} else if count, err = strconv.Atoi(r.FormValue("Count")); err != nil {
				result, err = "", errors.New("parse count error:"+err.Error())
			} else {
				result, err = hkey.HRandField(count)
			}
		case HVALS:
			if hkey, _, err = HashCtxWitchValueSchemaChecked(svcCtx.Key, RedisDataSource, nil); err != nil {
			} else {
				result, err = hkey.HVals()
			}
		case SISMEMBER:
			if skey, _, err = SetCtxWitchValueSchemaChecked(svcCtx.Key, RedisDataSource, nil); err != nil {
			} else {
				result, err = skey.SIsMember(r.FormValue("Member"))
			}
		case TIME:
			result = ""
			var tm time.Time
			if nonKey, _, err = CtxWithValueSchemaChecked(svcCtx.Key, "nonkey", RedisDataSource, nil); err != nil {
			} else if tm, err = nonKey.Time(); err == nil {
				result = tm.UnixMilli()
			}
		case ZRANGE:
			var (
				start, stop int64 = 0, -1
			)
			result = ""
			if zkey, _, err = ZSetCtxWitchValueSchemaChecked(svcCtx.Key, RedisDataSource, nil); err != nil {
			} else if start, err = strconv.ParseInt(r.FormValue("Start"), 10, 64); err != nil {
			} else if stop, err = strconv.ParseInt(r.FormValue("Stop"), 10, 64); err != nil {
			} else if r.FormValue("WITHSCORES") == "true" {
				// ZRANGE key start stop [WITHSCORES==true]
				var scores []float64
				if result, scores, err = zkey.ZRangeWithScores(start, stop); err == nil {
					result = map[string]interface{}{"members": result, "scores": scores}
				}
			} else {
				// ZRANGE key start stop [WITHSCORES==false]
				result, err = zkey.ZRange(start, stop)
			}
		case ZRANGEBYSCORE:
			var (
				offset, count int64 = 0, -1
				scores        []float64
			)
			result = ""
			if zkey, _, err = ZSetCtxWitchValueSchemaChecked(svcCtx.Key, RedisDataSource, nil); err != nil {
			} else if Min := r.FormValue("Min"); Min == "" {
				result, err = "false", errors.New("no Min")
			} else if Max := r.FormValue("Max"); Max == "" {
				result, err = "false", errors.New("no Max")
			} else if offset, err = strconv.ParseInt(r.FormValue("Offset"), 10, 64); err != nil {
				result, err = "false", errors.New("parse offset error:"+err.Error())
			} else if count, err = strconv.ParseInt(r.FormValue("Count"), 10, 64); err != nil {
				result, err = "false", errors.New("parse count error:"+err.Error())
			} else if r.FormValue("WITHSCORES") == "true" {
				//ZRANGEBYSCORE key min max [WITHSCORES==true]
				if result, scores, err = zkey.ZRangeByScoreWithScores(&redis.ZRangeBy{Min: Min, Max: Max, Offset: offset, Count: count}); err == nil {
					result = map[string]interface{}{"members": result, "scores": scores}
				}
			} else {
				//ZRANGEBYSCORE key min max [WITHSCORES==false]
				result, err = zkey.ZRangeByScore(&redis.ZRangeBy{Min: Min, Max: Max, Offset: offset, Count: count})
			}
		case ZREVRANGE:
			var (
				start, stop int64 = 0, -1
				rlts        []redis.Z
			)
			result = ""
			if zkey, _, err = ZSetCtxWitchValueSchemaChecked(svcCtx.Key, RedisDataSource, nil); err != nil {
			} else if start, err = strconv.ParseInt(r.FormValue("Start"), 10, 64); err != nil {
			} else if stop, err = strconv.ParseInt(r.FormValue("Stop"), 10, 64); err != nil {
			} else if r.FormValue("WITHSCORES") == "true" {
				// ZREVRANGE key start stop [WITHSCORES==true]
				cmd := zkey.Rds.ZRevRangeWithScores(context.Background(), zkey.Key, start, stop)
				if rlts, err = cmd.Result(); err == nil {
					var memberScoreSlice []interface{}
					for _, rlt := range rlts {
						memberScoreSlice = append(memberScoreSlice, rlt.Member)
						memberScoreSlice = append(memberScoreSlice, rlt.Score)
					}
					result = memberScoreSlice
				}
			} else {
				// ZREVRANGE key start stop [WITHSCORES==false]
				result, err = zkey.ZRevRange(start, stop)
			}
		case ZREVRANGEBYSCORE:
			var (
				offset, count int64 = 0, -1
				rlts          []redis.Z
			)

			if zkey, _, err = ZSetCtxWitchValueSchemaChecked(svcCtx.Key, RedisDataSource, nil); err != nil {
			} else if Min, Max := r.FormValue("Min"), r.FormValue("Max"); Min == "" || Max == "" {
				result, err = "", errors.New("no Min or Max")
			} else if offset, err = strconv.ParseInt(r.FormValue("Offset"), 10, 64); err != nil {
				result, err = "", errors.New("parse offset error:"+err.Error())
			} else if count, err = strconv.ParseInt(r.FormValue("Count"), 10, 64); err != nil {
				result, err = "", errors.New("parse count error:"+err.Error())
			} else if r.FormValue("WITHSCORES") == "true" {
				cmd := zkey.Rds.ZRevRangeByScoreWithScores(context.Background(), svcCtx.Key, &redis.ZRangeBy{Min: Min, Max: Max, Offset: offset, Count: count})
				if rlts, err = cmd.Result(); err == nil {
					var memberScoreSlice []interface{}
					for _, rlt := range rlts {
						memberScoreSlice = append(memberScoreSlice, rlt.Member)
						memberScoreSlice = append(memberScoreSlice, rlt.Score)
					}
					result, err = memberScoreSlice, nil
				}
			} else {
				//ZREVRANGEBYSCORE key max min [WITHSCORES==false]
				result, err = zkey.ZRevRangeByScore(&redis.ZRangeBy{Min: Min, Max: Max, Offset: offset, Count: count})
			}
		case ZRANK:
			if zkey, _, err = ZSetCtxWitchValueSchemaChecked(svcCtx.Key, RedisDataSource, nil); err != nil {
			} else {
				result, err = zkey.ZRank(r.FormValue("Member"))
			}
		case ZCOUNT:
			if zkey, _, err = ZSetCtxWitchValueSchemaChecked(svcCtx.Key, RedisDataSource, nil); err != nil {
			} else {
				result, err = zkey.ZCount(r.FormValue("Min"), r.FormValue("Max"))
			}
		case ZSCORE:
			if zkey, _, err = ZSetCtxWitchValueSchemaChecked(svcCtx.Key, RedisDataSource, nil); err != nil {
			} else {
				result, err = zkey.ZScore(r.FormValue("Member"))
			}
		case SCAN:
			result = ""
			if skey, _, err = SetCtxWitchValueSchemaChecked(svcCtx.Key, RedisDataSource, nil); err != nil {
			} else {
				var (
					cursor uint64
					count  int64
					keys   []string
					match  string
				)
				if cursor, err = strconv.ParseUint(r.FormValue("Cursor"), 10, 64); err != nil {
				} else if match = r.FormValue("Match"); match == "" {
				} else if count, err = strconv.ParseInt(r.FormValue("Count"), 10, 64); err != nil {
				} else if keys, cursor, err = skey.Scan(cursor, match, count); err != nil {
				} else {
					result, err = json.Marshal(map[string]interface{}{"keys": keys, "cursor": cursor})
				}
			}
		case LINDEX:
			var index int64
			if lKey, _, err = ListCtxWitchValueSchemaChecked(svcCtx.Key, RedisDataSource, nil); err != nil {
			} else if index, err = strconv.ParseInt(r.FormValue("Index"), 10, 64); err != nil {
				result, err = "", errors.New("parse index error:"+err.Error())
			} else {
				result, err = lKey.LIndex(index)
			}

		case LPOP:
			if lKey, _, err = ListCtxWitchValueSchemaChecked(svcCtx.Key, RedisDataSource, nil); err != nil {
			} else {
				result, err = lKey.LPop()
			}
		case LPUSH:
			result = "false"
			if lKey, result, err = ListCtxWitchValueSchemaChecked(svcCtx.Key, RedisDataSource, svcCtx.MsgpackBody(r, true)); err != nil {
			} else {
				err = lKey.LPush(result)
			}
		case LREM:
			var count int64
			if count, err = strconv.ParseInt(r.FormValue("Count"), 10, 64); err != nil {
				result, err = "false", errors.New("parse count error:"+err.Error())
				goto responseHttp
			}
			lKey, result, err = ListCtxWitchValueSchemaChecked(svcCtx.Key, RedisDataSource, svcCtx.MsgpackBody(r, true))
			if err != nil {
				goto responseHttp
			}
			if err = lKey.LRem(count, result); err == nil {
				result = "true"
			}

		case LSET:
			result = "false"
			var index int64
			if index, err = strconv.ParseInt(r.FormValue("Index"), 10, 64); err != nil {
				err = errors.New("parse index error:" + err.Error())
				goto responseHttp
			}
			lKey, result, err = ListCtxWitchValueSchemaChecked(svcCtx.Key, RedisDataSource, svcCtx.MsgpackBody(r, true))
			if err != nil {
				goto responseHttp
			}
			if err = lKey.LSet(index, result); err == nil {
				result = "true"
			}

		case LTRIM:
			var start, stop int64
			if start, err = strconv.ParseInt(r.FormValue("Start"), 10, 64); err != nil {
				result, err = "false", errors.New("parse start error:"+err.Error())
			} else if stop, err = strconv.ParseInt(r.FormValue("Stop"), 10, 64); err != nil {
				result, err = "false", errors.New("parse stop error:"+err.Error())
			} else if err = rds.LTrim(svcCtx.Ctx, svcCtx.Key, start, stop).Err(); err == nil {
				result = "true"
			}
		case RPOP:
			if lKey, _, err = ListCtxWitchValueSchemaChecked(svcCtx.Key, RedisDataSource, nil); err != nil {
			} else {
				result, err = lKey.RPop()
			}
		case RPUSH:
			result = "false"
			lKey, result, err = ListCtxWitchValueSchemaChecked(svcCtx.Key, RedisDataSource, svcCtx.MsgpackBody(r, true))
			if err != nil {
				goto responseHttp
			}
			if err = lKey.RPush(svcCtx.Ctx, svcCtx.Key, result); err == nil {
				result = "true"
			}
		case RPUSHX:
			result = "false"
			lKey, result, err = ListCtxWitchValueSchemaChecked(svcCtx.Key, RedisDataSource, svcCtx.MsgpackBody(r, true))
			if err != nil {
			} else if err = lKey.RPushX(svcCtx.Ctx, svcCtx.Key, result); err == nil {
				result = "true"
			}

		case ZADD:
			var Score float64
			zkey, result, err = ZSetCtxWitchValueSchemaChecked(svcCtx.Key, RedisDataSource, svcCtx.MsgpackBody(r, true))
			if err != nil {
			} else if Score, err = strconv.ParseFloat(r.FormValue("Score"), 64); err != nil {
				result, err = "false", errors.New("parameter Score shoule be float")
			} else if err = zkey.ZAdd(redis.Z{Score: Score, Member: result}); err != nil {
				result = "false"
			} else {
				result = "true"
			}

		case SET:
			result = "false"
			if strKey, result, err = StringCtxWitchValueSchemaChecked(svcCtx.Key, RedisDataSource, bs); err != nil {
			} else {
				err = strKey.Set(svcCtx.Key+":"+svcCtx.Field, result, 0)
			}
		case DEL:
			result = "false"
			if err = rds.HDel(svcCtx.Ctx, svcCtx.Key, "del").Err(); err == nil {
				result = "true"
			}
		case ZREM:
			result = "false"
			var MemberStr = strings.Split(r.FormValue("Member"), ",")
			//convert Member to []interface{}
			var Member = make([]interface{}, len(MemberStr))
			for i, v := range MemberStr {
				Member[i] = v
			}
			if len(Member) == 0 {
				err = errors.New("no Member")
				goto responseHttp
			}
			zkey, _, err = ZSetCtxWitchValueSchemaChecked(svcCtx.Key, RedisDataSource, nil)
			if err != nil {
			} else if err = zkey.ZRem(Member...); err == nil {
				result = "true"
			}

		case ZREMRANGEBYSCORE:
			result = "false"
			var Min, Max = r.FormValue("Min"), r.FormValue("Max")
			if Min == "" || Max == "" {
				err = errors.New("no Min or Max")
			} else if err = rds.ZRemRangeByScore(svcCtx.Ctx, svcCtx.Key, Min, Max).Err(); err == nil {
				result = "true"
			}
		case TYPE:
			result, err = rds.Type(svcCtx.Ctx, svcCtx.Key).Result()
		case EXPIRE:
			var seconds int64
			if seconds, err = strconv.ParseInt(r.FormValue("Seconds"), 10, 64); err != nil {
				result, err = "false", errors.New("parse seconds error:"+err.Error())
			} else if err = rds.Expire(svcCtx.Ctx, svcCtx.Key, time.Duration(seconds)*time.Second).Err(); err == nil {
				result = "true"
			}
		case EXPIREAT:
			var timestamp int64
			if timestamp, err = strconv.ParseInt(r.FormValue("Timestamp"), 10, 64); err != nil {
				result, err = "false", errors.New("parse seconds error:"+err.Error())
			} else if err = rds.ExpireAt(svcCtx.Ctx, svcCtx.Key, time.Unix(timestamp, 0)).Err(); err == nil {
				result = "true"
			}
		case PERSIST:
			if err = rds.Persist(svcCtx.Ctx, svcCtx.Key).Err(); err == nil {
				result = "true"
			}
		case TTL:
			result, err = rds.TTL(svcCtx.Ctx, svcCtx.Key).Result()
		case PTTL:
			result, err = rds.PTTL(svcCtx.Ctx, svcCtx.Key).Result()
		case RENAME:
			if newKey := r.FormValue("NewKey"); newKey == "" {
				result, err = "false", errors.New("no NewKey")
			} else if err = rds.Rename(svcCtx.Ctx, svcCtx.Key, newKey).Err(); err == nil {
				result = "true"
			}
		case RENAMEX:
			if newKey := r.FormValue("NewKey"); newKey == "" {
				result, err = "false", errors.New("no NewKey")
			} else if err = rds.RenameNX(svcCtx.Ctx, svcCtx.Key, newKey).Err(); err == nil {
				result = "true"
			}
		case EXISTS:
			result, err = rds.Exists(svcCtx.Ctx, svcCtx.Key).Result()
		case KEYS:
			result, err = rds.Keys(svcCtx.Ctx, svcCtx.Key).Result()
		case ZINCRBY:
			var (
				incr   float64
				member string
			)
			if member = r.FormValue("Member"); member == "" {
				result, err = "false", errors.New("no Member")
			} else if incr, err = strconv.ParseFloat(r.FormValue("Incr"), 64); err != nil {
				result, err = "false", errors.New("parse Incr error:"+err.Error())
			} else if err = rds.ZIncrBy(svcCtx.Ctx, svcCtx.Key, incr, member).Err(); err == nil {
				result = "true"
			}
		case HINCRBY:
			var (
				incr int64
			)
			if incr, err = strconv.ParseInt(r.FormValue("Incr"), 10, 64); err != nil {
				result, err = "false", errors.New("parse Incr error:"+err.Error())
			} else if err = rds.HIncrBy(svcCtx.Ctx, svcCtx.Key, svcCtx.Field, incr).Err(); err == nil {
				result = "true"
			}
		case HINCRBYFLOAT:
			var (
				incr float64
			)
			if incr, err = strconv.ParseFloat(r.FormValue("Incr"), 64); err != nil {
				result, err = "false", errors.New("parse Incr error:"+err.Error())
			} else if err = rds.HIncrByFloat(svcCtx.Ctx, svcCtx.Key, svcCtx.Field, incr).Err(); err == nil {
				result = "true"
			}
		case XADD:
			if id := r.FormValue("ID"); id == "" {
				result, err = "false", errors.New("no ID")
			} else if err = rds.XAdd(svcCtx.Ctx, &redis.XAddArgs{Stream: svcCtx.Key, ID: id, Values: svcCtx.MsgpackBody(r, true)}).Err(); err != nil {
				result = "false"
			}
		case XDEL:
			if id := r.FormValue("ID"); id == "" {
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

	responseHttp:
		if len(cfghttp.CORES) > 0 {
			w.Header().Set("Access-Control-Allow-Origin", cfghttp.CORES)
		}

		if err == nil {
			if ResponseContentType == "application/msgpack" {
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
		if svcCtx != nil && len(ResponseContentType) > 0 {
			w.Header().Set("Content-Type", ResponseContentType)
		}

		w.WriteHeader(httpStatus)
		w.Write(bs)

	})
	server := &http.Server{
		Addr:              ":" + strconv.FormatInt(port, 10),
		Handler:           httpRoter,
		ReadTimeout:       50 * time.Second,
		ReadHeaderTimeout: 50 * time.Second,
		WriteTimeout:      50 * time.Second, //10ms Redundant time
		IdleTimeout:       15 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		logger.Error().Err(err).Msg("http server ListenAndServe error")
		return
	}
	logger.Info().Any("port", port).Any("path", path).Msg("doptime http server started!")
}

func init() {
	//wait 1s, till the permission table is loaded
	for i := 0; !redisdbPermissionTableLoaded && i < 100; i++ {
		time.Sleep(time.Millisecond * 10)
	}
	//wait, till all the apis are loaded
	logger.Info().Any("port", cfghttp.Port).Any("path", cfghttp.Path).Msg("doptime http server is starting")
	go httpStart(cfghttp.Path, cfghttp.Port)
}
