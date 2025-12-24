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
			result     interface{}
			bs         []byte
			ok         bool
			err        error
			httpStatus int = http.StatusOK
			svcCtx     *DoptimeReqCtx
			rds        *redis.Client
			hkey       *redisdb.HashKey[string, interface{}]
			skey       *redisdb.SetKey[string, interface{}]
			zkey       *redisdb.ZSetKey[string, interface{}]
			lKey       *redisdb.ListKey[interface{}]
			strKey     *redisdb.StringKey[string, interface{}]
			s          string = ""
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
		if err = svcCtx.ReplaceKeyFieldTagWithJwtClaims(); err != nil {
			httpStatus = http.StatusInternalServerError
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
		case HLEN:
			if !redisdb.IsAllowedHashOp(svcCtx.Key, redisdb.HLen) {
				goto disallowedPermission
			}
			result, err = rds.HLen(ctx, svcCtx.Key).Result()
		case LLEN:
			if !redisdb.IsAllowedListOp(svcCtx.Key, redisdb.LLen) {
				goto disallowedPermission
			}
			result, err = rds.LLen(ctx, svcCtx.Key).Result()
		case XLEN:
			if !redisdb.IsAllowedStreamOp(svcCtx.Key, redisdb.XLen) {
				goto disallowedPermission
			}
			result, err = rds.XLen(ctx, svcCtx.Key).Result()
		case ZCARD:
			if !redisdb.IsAllowedZSetOp(svcCtx.Key, redisdb.ZCard) {
				goto disallowedPermission
			}
			result, err = rds.ZCard(ctx, svcCtx.Key).Result()
		case SCARD:
			if !redisdb.IsAllowedSetOp(svcCtx.Key, redisdb.SCard) {
				goto disallowedPermission
			}
			result, err = rds.SCard(ctx, svcCtx.Key).Result()

		case SSCAN:
			if !redisdb.IsAllowedSetOp(svcCtx.Key, redisdb.SScan) {
				goto disallowedPermission
			}
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
			if !redisdb.IsAllowedHashOp(svcCtx.Key, redisdb.HScan) {
				goto disallowedPermission
			}
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
			if !redisdb.IsAllowedZSetOp(svcCtx.Key, redisdb.ZScan) {
				goto disallowedPermission
			}
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
			if !redisdb.IsAllowedListOp(svcCtx.Key, redisdb.LRange) {
				goto disallowedPermission
			}
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
		case XRANGE, XRANGEN:
			if !redisdb.IsAllowedStreamOp(svcCtx.Key, redisdb.XRange) {
				goto disallowedPermission
			}
			var (
				start, stop string
				count       int64
			)
			if start = r.FormValue("Start"); start == "" {
				result, err = "false", errors.New("no Start")
			} else if stop = r.FormValue("Stop"); stop == "" {
				result, err = "false", errors.New("no Stop")
			} else if svcCtx.Cmd == XRANGEN {
				if count, err = strconv.ParseInt(r.FormValue("Count"), 10, 64); err != nil {
					result, err = "false", errors.New("parse N error:"+err.Error())
				} else {
					result, err = rds.XRangeN(svcCtx.Ctx, svcCtx.Key, start, stop, count).Result()
				}
			} else {
				result, err = rds.XRange(svcCtx.Ctx, svcCtx.Key, start, stop).Result()
			}
		case XREVRANGE, XREVRANGEN:
			if !redisdb.IsAllowedStreamOp(svcCtx.Key, redisdb.XRange) {
				goto disallowedPermission
			}
			var (
				start, stop string
				count       int64
			)
			if start = r.FormValue("Start"); start == "" {
				result, err = "false", errors.New("no Start")
			} else if stop = r.FormValue("Stop"); stop == "" {
				result, err = "false", errors.New("no Stop")
			} else if svcCtx.Cmd == XREVRANGEN {
				if count, err = strconv.ParseInt(r.FormValue("Count"), 10, 64); err != nil {
					result, err = "false", errors.New("parse N error:"+err.Error())
				} else {
					result, err = rds.XRevRangeN(svcCtx.Ctx, svcCtx.Key, start, stop, count).Result()
				}
			} else {
				result, err = rds.XRevRange(svcCtx.Ctx, svcCtx.Key, start, stop).Result()
			}
		case XREAD:
			if !redisdb.IsAllowedStreamOp(svcCtx.Key, redisdb.XRead) {
				goto disallowedPermission
			}
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
			if !redisdb.IsAllowedStringOp(svcCtx.Key, redisdb.Get) {
				goto disallowedPermission
			}
			if strKey, _, err = StringCtxWitchValueSchemaChecked(svcCtx.Key, RedisDataSource, nil); err != nil {
			} else {
				result, err = strKey.Get(svcCtx.Field)
			}
		case HGET:
			if !redisdb.IsAllowedHashOp(svcCtx.Key, redisdb.HGet) {
				goto disallowedPermission
			}
			if hkey, _, err = HashCtxWitchValueSchemaChecked(svcCtx.Key, RedisDataSource, nil); err != nil {
			} else {
				result, err = hkey.HGet(svcCtx.Field)
			}
		case HGETALL:
			if !redisdb.IsAllowedHashOp(svcCtx.Key, redisdb.HGetAll) {
				goto disallowedPermission
			}
			if hkey, _, err = HashCtxWitchValueSchemaChecked(svcCtx.Key, RedisDataSource, nil); err != nil {
			} else {
				result, err = hkey.HGetAll()
			}
		case HMGET:
			if !redisdb.IsAllowedHashOp(svcCtx.Key, redisdb.HMGET) {
				goto disallowedPermission
			}
			if hkey, result, err = HashCtxWitchValueSchemaChecked(svcCtx.Key, RedisDataSource, nil); err != nil {
			} else {
				var fields []interface{} = sliceToInterface(strings.Split(svcCtx.Field, ","))
				result, err = hkey.HMGET(fields...)
			}

		case HSET:
			if !redisdb.IsAllowedHashOp(svcCtx.Key, redisdb.HSet) {
				goto disallowedPermission
			}
			result = 0
			if hkey, result, err = HashCtxWitchValueSchemaChecked(svcCtx.Key, RedisDataSource, svcCtx.MsgpackBody(r, true)); err != nil {
			} else {
				result, err = hkey.HSet(svcCtx.Field, result)
			}

		case HDEL:
			if !redisdb.IsAllowedHashOp(svcCtx.Key, redisdb.HDel) {
				goto disallowedPermission
			}
			result, err = rds.HDel(svcCtx.Ctx, svcCtx.Key, svcCtx.Field).Result()
		case HKEYS:
			if !redisdb.IsAllowedHashOp(svcCtx.Key, redisdb.HKeys) {
				goto disallowedPermission
			}
			if hkey, _, err = HashCtxWitchValueSchemaChecked(svcCtx.Key, RedisDataSource, nil); err != nil {
			} else {
				result, err = hkey.HKeys()
			}
		case HEXISTS:
			if !redisdb.IsAllowedHashOp(svcCtx.Key, redisdb.HExists) {
				goto disallowedPermission
			}
			result = false
			if hkey, _, err = HashCtxWitchValueSchemaChecked(svcCtx.Key, RedisDataSource, nil); err != nil {
			} else {
				result, err = hkey.HExists(svcCtx.Field)
			}
		case HRANDFIELD:
			if !redisdb.IsAllowedHashOp(svcCtx.Key, redisdb.HRandField) {
				goto disallowedPermission
			}
			var count int
			if hkey, _, err = HashCtxWitchValueSchemaChecked(svcCtx.Key, RedisDataSource, nil); err != nil {
			} else if count, err = strconv.Atoi(r.FormValue("Count")); err != nil {
				result, err = "", errors.New("parse count error:"+err.Error())
			} else {
				result, err = hkey.HRandField(count)
			}
		case HVALS:
			if !redisdb.IsAllowedHashOp(svcCtx.Key, redisdb.HVals) {
				goto disallowedPermission
			}
			if hkey, _, err = HashCtxWitchValueSchemaChecked(svcCtx.Key, RedisDataSource, nil); err != nil {
			} else {
				result, err = hkey.HVals()
			}
		case SISMEMBER:
			if !redisdb.IsAllowedSetOp(svcCtx.Key, redisdb.SIsMember) {
				goto disallowedPermission
			}
			if skey, _, err = SetCtxWitchValueSchemaChecked(svcCtx.Key, RedisDataSource, nil); err != nil {
			} else {
				result, err = skey.SIsMember(r.FormValue("Member"))
			}

		case ZRANGE:
			if !redisdb.IsAllowedZSetOp(svcCtx.Key, redisdb.ZRange) {
				goto disallowedPermission
			}
			var (
				start, stop int64 = 0, -1
			)
			result = ""
			if zkey, _, err = ZSetCtxWitchValueSchemaChecked(svcCtx.Key, RedisDataSource, nil); err != nil {
			} else if start, err = strconv.ParseInt(r.FormValue("Start"), 10, 64); err != nil {
			} else if stop, err = strconv.ParseInt(r.FormValue("Stop"), 10, 64); err != nil {
			} else if r.FormValue("WITHSCORES") == "true" {
				var scores []float64
				if result, scores, err = zkey.ZRangeWithScores(start, stop); err == nil {
					result = map[string]interface{}{"members": result, "scores": scores}
				}
			} else {
				result, err = zkey.ZRange(start, stop)
			}
		case ZRANGEBYSCORE:
			if !redisdb.IsAllowedZSetOp(svcCtx.Key, redisdb.ZRangeByScore) {
				goto disallowedPermission
			}
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
				if result, scores, err = zkey.ZRangeByScoreWithScores(&redis.ZRangeBy{Min: Min, Max: Max, Offset: offset, Count: count}); err == nil {
					result = map[string]interface{}{"members": result, "scores": scores}
				}
			} else {
				result, err = zkey.ZRangeByScore(&redis.ZRangeBy{Min: Min, Max: Max, Offset: offset, Count: count})
			}
		case ZREVRANGE:
			if !redisdb.IsAllowedZSetOp(svcCtx.Key, redisdb.ZRevRange) {
				goto disallowedPermission
			}
			var (
				start, stop int64 = 0, -1
				rlts        []redis.Z
			)
			result = ""
			if zkey, _, err = ZSetCtxWitchValueSchemaChecked(svcCtx.Key, RedisDataSource, nil); err != nil {
			} else if start, err = strconv.ParseInt(r.FormValue("Start"), 10, 64); err != nil {
			} else if stop, err = strconv.ParseInt(r.FormValue("Stop"), 10, 64); err != nil {
			} else if r.FormValue("WITHSCORES") == "true" {
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
				result, err = zkey.ZRevRange(start, stop)
			}
		case ZREVRANGEBYSCORE:
			if !redisdb.IsAllowedZSetOp(svcCtx.Key, redisdb.ZRevRangeByScore) {
				goto disallowedPermission
			}
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
				result, err = zkey.ZRevRangeByScore(&redis.ZRangeBy{Min: Min, Max: Max, Offset: offset, Count: count})
			}
		case ZRANK:
			if !redisdb.IsAllowedZSetOp(svcCtx.Key, redisdb.ZRank) {
				goto disallowedPermission
			}
			if zkey, _, err = ZSetCtxWitchValueSchemaChecked(svcCtx.Key, RedisDataSource, nil); err != nil {
			} else {
				result, err = zkey.ZRank(r.FormValue("Member"))
			}
		case ZCOUNT:
			if !redisdb.IsAllowedZSetOp(svcCtx.Key, redisdb.ZCount) {
				goto disallowedPermission
			}
			if zkey, _, err = ZSetCtxWitchValueSchemaChecked(svcCtx.Key, RedisDataSource, nil); err != nil {
			} else {
				result, err = zkey.ZCount(r.FormValue("Min"), r.FormValue("Max"))
			}
		case ZSCORE:
			if !redisdb.IsAllowedZSetOp(svcCtx.Key, redisdb.ZScore) {
				goto disallowedPermission
			}
			if zkey, _, err = ZSetCtxWitchValueSchemaChecked(svcCtx.Key, RedisDataSource, nil); err != nil {
			} else {
				result, err = zkey.ZScore(r.FormValue("Member"))
			}

		case SCAN:
			if !redisdb.IsAllowedZSetOp(svcCtx.Key, redisdb.ZSetOp(redisdb.SScan)) {
				goto disallowedPermission
			}
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
			if !redisdb.IsAllowedListOp(svcCtx.Key, redisdb.LIndex) {
				goto disallowedPermission
			}
			var index int64
			if lKey, _, err = ListCtxWitchValueSchemaChecked(svcCtx.Key, RedisDataSource, nil); err != nil {
			} else if index, err = strconv.ParseInt(r.FormValue("Index"), 10, 64); err != nil {
				result, err = "", errors.New("parse index error:"+err.Error())
			} else {
				result, err = lKey.LIndex(index)
			}
		case LPOP:
			if !redisdb.IsAllowedListOp(svcCtx.Key, redisdb.LPop) {
				goto disallowedPermission
			}
			if lKey, _, err = ListCtxWitchValueSchemaChecked(svcCtx.Key, RedisDataSource, nil); err != nil {
			} else {
				result, err = lKey.LPop()
			}
		case LPUSH:
			if !redisdb.IsAllowedListOp(svcCtx.Key, redisdb.LPush) {
				goto disallowedPermission
			}
			result = "false"
			if lKey, result, err = ListCtxWitchValueSchemaChecked(svcCtx.Key, RedisDataSource, svcCtx.MsgpackBody(r, true)); err != nil {
			} else {
				err = lKey.LPush(result)
			}
		case LREM:
			if !redisdb.IsAllowedListOp(svcCtx.Key, redisdb.LRem) {
				goto disallowedPermission
			}
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
			if !redisdb.IsAllowedListOp(svcCtx.Key, redisdb.LSet) {
				goto disallowedPermission
			}
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
			if !redisdb.IsAllowedListOp(svcCtx.Key, redisdb.LTrim) {
				goto disallowedPermission
			}
			var start, stop int64
			if start, err = strconv.ParseInt(r.FormValue("Start"), 10, 64); err != nil {
				result, err = "false", errors.New("parse start error:"+err.Error())
			} else if stop, err = strconv.ParseInt(r.FormValue("Stop"), 10, 64); err != nil {
				result, err = "false", errors.New("parse stop error:"+err.Error())
			} else if err = rds.LTrim(svcCtx.Ctx, svcCtx.Key, start, stop).Err(); err == nil {
				result = "true"
			}
		case RPOP:
			if !redisdb.IsAllowedListOp(svcCtx.Key, redisdb.RPop) {
				goto disallowedPermission
			}
			if lKey, _, err = ListCtxWitchValueSchemaChecked(svcCtx.Key, RedisDataSource, nil); err != nil {
			} else {
				result, err = lKey.RPop()
			}
		case RPUSH:
			if !redisdb.IsAllowedListOp(svcCtx.Key, redisdb.RPush) {
				goto disallowedPermission
			}
			result = "false"
			lKey, result, err = ListCtxWitchValueSchemaChecked(svcCtx.Key, RedisDataSource, svcCtx.MsgpackBody(r, true))
			if err != nil {
				goto responseHttp
			}
			if err = lKey.RPush(svcCtx.Ctx, svcCtx.Key, result); err == nil {
				result = "true"
			}
		case RPUSHX:
			if !redisdb.IsAllowedListOp(svcCtx.Key, redisdb.RPushX) {
				goto disallowedPermission
			}
			result = "false"
			lKey, result, err = ListCtxWitchValueSchemaChecked(svcCtx.Key, RedisDataSource, svcCtx.MsgpackBody(r, true))
			if err != nil {
			} else if err = lKey.RPushX(svcCtx.Ctx, svcCtx.Key, result); err == nil {
				result = "true"
			}
		case ZADD:
			if !redisdb.IsAllowedZSetOp(svcCtx.Key, redisdb.ZAdd) {
				goto disallowedPermission
			}
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
			if !redisdb.IsAllowedStringOp(svcCtx.Key, redisdb.Set) {
				goto disallowedPermission
			}
			result = "false"
			if strKey, result, err = StringCtxWitchValueSchemaChecked(svcCtx.Key, RedisDataSource, bs); err != nil {
			} else {
				err = strKey.Set(svcCtx.Key+":"+svcCtx.Field, result, 0)
			}
		case DEL:
			if !redisdb.IsAllowedStringOp(svcCtx.Key, redisdb.Del) {
				goto disallowedPermission
			}
			result = "false"
			if err = rds.HDel(svcCtx.Ctx, svcCtx.Key, "del").Err(); err == nil {
				result = "true"
			}
		case ZREM:
			if !redisdb.IsAllowedZSetOp(svcCtx.Key, redisdb.ZRem) {
				goto disallowedPermission
			}
			result = "false"
			var MemberStr = strings.Split(r.FormValue("Member"), ",")
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
			if !redisdb.IsAllowedZSetOp(svcCtx.Key, redisdb.ZRemRangeByScore) {
				goto disallowedPermission
			}
			result = "false"
			var Min, Max = r.FormValue("Min"), r.FormValue("Max")
			if Min == "" || Max == "" {
				err = errors.New("no Min or Max")
			} else if err = rds.ZRemRangeByScore(svcCtx.Ctx, svcCtx.Key, Min, Max).Err(); err == nil {
				result = "true"
			}
		case ZINCRBY:
			if !redisdb.IsAllowedZSetOp(svcCtx.Key, redisdb.ZIncrBy) {
				goto disallowedPermission
			}
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
			if !redisdb.IsAllowedHashOp(svcCtx.Key, redisdb.HIncrBy) {
				goto disallowedPermission
			}
			var (
				incr int64
			)
			if incr, err = strconv.ParseInt(r.FormValue("Incr"), 10, 64); err != nil {
				result, err = "false", errors.New("parse Incr error:"+err.Error())
			} else if err = rds.HIncrBy(svcCtx.Ctx, svcCtx.Key, svcCtx.Field, incr).Err(); err == nil {
				result = "true"
			}
		case HINCRBYFLOAT:
			if !redisdb.IsAllowedHashOp(svcCtx.Key, redisdb.HIncrByFloat) {
				goto disallowedPermission
			}
			var (
				incr float64
			)
			if incr, err = strconv.ParseFloat(r.FormValue("Incr"), 64); err != nil {
				result, err = "false", errors.New("parse Incr error:"+err.Error())
			} else if err = rds.HIncrByFloat(svcCtx.Ctx, svcCtx.Key, svcCtx.Field, incr).Err(); err == nil {
				result = "true"
			}
		case XADD:
			if !redisdb.IsAllowedStreamOp(svcCtx.Key, redisdb.XAdd) {
				goto disallowedPermission
			}
			if id := r.FormValue("ID"); id == "" {
				result, err = "false", errors.New("no ID")
			} else if err = rds.XAdd(svcCtx.Ctx, &redis.XAddArgs{Stream: svcCtx.Key, ID: id, Values: svcCtx.MsgpackBody(r, true)}).Err(); err != nil {
				result = "false"
			}
		case XDEL:
			if !redisdb.IsAllowedStreamOp(svcCtx.Key, redisdb.XDel) {
				goto disallowedPermission
			}
			if id := r.FormValue("ID"); id == "" {
				result, err = "false", errors.New("no ID")
			} else if err = rds.XDel(svcCtx.Ctx, svcCtx.Key, id).Err(); err != nil {
				result = "false"
			} else {
				result = "true"
			}

		case TYPE:
			if !redisdb.IsAllowedCommon(svcCtx.Key, redisdb.Type) {
				goto disallowedPermission
			}
			result, err = rds.Type(svcCtx.Ctx, svcCtx.Key).Result()
		case EXPIRE:
			if !redisdb.IsAllowedCommon(svcCtx.Key, redisdb.Expire) {
				goto disallowedPermission
			}
			var seconds int64
			if seconds, err = strconv.ParseInt(r.FormValue("Seconds"), 10, 64); err != nil {
				result, err = "false", errors.New("parse seconds error:"+err.Error())
			} else if err = rds.Expire(svcCtx.Ctx, svcCtx.Key, time.Duration(seconds)*time.Second).Err(); err == nil {
				result = "true"
			}
		case EXPIREAT:
			if !redisdb.IsAllowedCommon(svcCtx.Key, redisdb.Expire) {
				goto disallowedPermission
			}
			var timestamp int64
			if timestamp, err = strconv.ParseInt(r.FormValue("Timestamp"), 10, 64); err != nil {
				result, err = "false", errors.New("parse seconds error:"+err.Error())
			} else if err = rds.ExpireAt(svcCtx.Ctx, svcCtx.Key, time.Unix(timestamp, 0)).Err(); err == nil {
				result = "true"
			}
		case PERSIST:
			if !redisdb.IsAllowedCommon(svcCtx.Key, redisdb.Persist) {
				goto disallowedPermission
			}
			if err = rds.Persist(svcCtx.Ctx, svcCtx.Key).Err(); err == nil {
				result = "true"
			}
		case TTL:
			if !redisdb.IsAllowedCommon(svcCtx.Key, redisdb.TTL) {
				goto disallowedPermission
			}
			result, err = rds.TTL(svcCtx.Ctx, svcCtx.Key).Result()
		case PTTL:
			if !redisdb.IsAllowedCommon(svcCtx.Key, redisdb.TTL) {
				goto disallowedPermission
			}
			result, err = rds.PTTL(svcCtx.Ctx, svcCtx.Key).Result()
		case RENAME:
			if !redisdb.IsAllowedCommon(svcCtx.Key, redisdb.Rename) {
				goto disallowedPermission
			}
			if newKey := r.FormValue("NewKey"); newKey == "" {
				result, err = "false", errors.New("no NewKey")
			} else if err = rds.Rename(svcCtx.Ctx, svcCtx.Key, newKey).Err(); err == nil {
				result = "true"
			}
		case RENAMEX:
			if !redisdb.IsAllowedCommon(svcCtx.Key, redisdb.Rename) {
				goto disallowedPermission
			}
			if newKey := r.FormValue("NewKey"); newKey == "" {
				result, err = "false", errors.New("no NewKey")
			} else if err = rds.RenameNX(svcCtx.Ctx, svcCtx.Key, newKey).Err(); err == nil {
				result = "true"
			}
		case EXISTS:
			if !redisdb.IsAllowedCommon(svcCtx.Key, redisdb.Exists) {
				goto disallowedPermission
			}
			result, err = rds.Exists(svcCtx.Ctx, svcCtx.Key).Result()
		case TIME:
			if !redisdb.IsAllowedDBOp(redisdb.DBTime) {
				goto disallowedPermission
			}
			result = ""
			var tm time.Time
			var nonKey = redisdb.NewRedisKey[string, interface{}](redisdb.Opt.Key("nonkey"), redisdb.Opt.Rds(RedisDataSource))
			if tm, err = nonKey.Time(); err == nil {
				result = tm.UnixMilli()
			}
		case KEYS:
			if !redisdb.IsAllowedDBOp(redisdb.DBKeys) {
				goto disallowedPermission
			}
			result, err = rds.Keys(svcCtx.Ctx, svcCtx.Key).Result()
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
		return

	disallowedPermission:
		httpStatus = http.StatusForbidden
		err = ErrOperationNotPermited
		goto responseHttp

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
	//wait, till all the apis are loaded
	logger.Info().Any("port", cfghttp.Port).Any("path", cfghttp.Path).Msg("doptime http server is starting")
	go httpStart(cfghttp.Path, cfghttp.Port)
}
