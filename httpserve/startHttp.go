package httpserve

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/doptime/config/cfghttp"
	"github.com/doptime/doptime/httpserve/httpapi"
	"github.com/doptime/doptime/lib"
	"github.com/doptime/logger"
	"github.com/doptime/redisdb"
	"github.com/redis/go-redis/v9"
	"github.com/vmihailenco/msgpack/v5"
)

func Debug() {}

var httpRoter = http.NewServeMux()
var mu sync.Mutex

func AddRoute(path string, handlerFunc http.HandlerFunc) {
	mu.Lock()
	defer mu.Unlock()
	httpRoter.HandleFunc(path, handlerFunc)
}

func httpStart(path string, port int64) {
	httpRoter.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		var (
			result       interface{}
			values       []interface{}
			valInterface interface{}
			bs           []byte
			ok           bool
			err          error
			httpStatus   int = http.StatusOK
			svcCtx       *DoptimeReqCtx

			// 定义所有类型的接口变量
			hkey      redisdb.IHttpHashKey
			skey      redisdb.IHttpSetKey
			zkey      redisdb.IHttpZSetKey
			lKey      redisdb.IHttpListKey
			strKey    redisdb.IHttpStringKey
			streamKey redisdb.IHttpStreamKey
			vectorKey redisdb.IHttpVectorSetKey

			s                   string = ""
			ResponseContentType string = lib.Ternary(r.FormValue("rt") == "", "application/json", r.FormValue("rt"))
		)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*120)
		defer cancel()

		if CorsChecked(r, w) {
			goto responseHttp
		}

		if svcCtx, err, httpStatus = NewHttpContext(ctx, r, w); httpStatus != http.StatusOK {
			goto responseHttp
		}

		// API Logic
		if _, isDataKey := DataCmdRequireKey[svcCtx.Cmd]; !isDataKey {
			ServiceName := svcCtx.Key
			_api, ok := httpapi.GetApiByName(ServiceName)
			if !ok {
				result, err = nil, fmt.Errorf("err no such api")
				goto responseHttp
			}
			msgpackNonstruct, jsonpackNostruct := svcCtx.BuildParamFromBody(r)
			result, err = _api.CallByMap(svcCtx.Params, msgpackNonstruct, jsonpackNostruct)
			goto responseHttp
		}

		// Data Operation Logic
		switch svcCtx.Cmd {

		// --- LEN / CARD 类 ---
		case HLEN:
			if !redisdb.IsAllowedHashOp(svcCtx.Key, redisdb.HLen) {
				goto disallowedPermission
			}
			// HLen 暂未在 IHttpHashKey 定义，兜底使用 RdsClient
			result, err = svcCtx.RdsClient.HLen(ctx, svcCtx.Key).Result()
		case LLEN:
			if !redisdb.IsAllowedListOp(svcCtx.Key, redisdb.LLen) {
				goto disallowedPermission
			}
			if lKey, err = redisdb.GetHttpListKey(svcCtx.Key, svcCtx.RedisDataSource); err == nil {
				result, err = lKey.LLen()
			}
		case XLEN:
			if !redisdb.IsAllowedStreamOp(svcCtx.Key, redisdb.XLen) {
				goto disallowedPermission
			}
			if streamKey, err = redisdb.GetHttpStreamKey(svcCtx.Key, svcCtx.RedisDataSource); err == nil {
				result, err = streamKey.XLen()
			}
		case ZCARD:
			if !redisdb.IsAllowedZSetOp(svcCtx.Key, redisdb.ZCard) {
				goto disallowedPermission
			}
			if zkey, err = redisdb.GetHttpZSetKey(svcCtx.Key, svcCtx.RedisDataSource); err == nil {
				result, err = zkey.ZCard()
			}
		case SCARD:
			if !redisdb.IsAllowedSetOp(svcCtx.Key, redisdb.SCard) {
				goto disallowedPermission
			}
			if skey, err = redisdb.GetHttpSetKey(svcCtx.Key, svcCtx.RedisDataSource); err == nil {
				result, err = skey.SCard()
			}

		// --- SCAN 类 ---
		case SSCAN:
			if !redisdb.IsAllowedSetOp(svcCtx.Key, redisdb.SScan) {
				goto disallowedPermission
			}
			var cursor uint64
			var count int64
			var match string
			if skey, err = redisdb.GetHttpSetKey(svcCtx.Key, svcCtx.RedisDataSource); err != nil {
			} else if cursor, err = strconv.ParseUint(r.FormValue("Cursor"), 10, 64); err != nil {
			} else if match = r.FormValue("Match"); match == "" {
			} else if count, err = strconv.ParseInt(r.FormValue("Count"), 10, 64); err != nil {
			} else {
				values, cursor, err = skey.SScan(cursor, match, count)
				if err == nil {
					result = map[string]interface{}{"values": values, "cursor": cursor}
				}
			}

		case HSCAN:
			if !redisdb.IsAllowedHashOp(svcCtx.Key, redisdb.HScan) {
				goto disallowedPermission
			}
			var cursor uint64
			var count int64
			var match string
			var keys []string
			var cursorRet uint64

			if hkey, err = redisdb.GetHttpHashKey(svcCtx.Key, svcCtx.RedisDataSource); err != nil {
			} else if cursor, err = strconv.ParseUint(r.FormValue("Cursor"), 10, 64); err != nil {
			} else if match = r.FormValue("Match"); match == "" {
			} else if count, err = strconv.ParseInt(r.FormValue("Count"), 10, 64); err != nil {
			} else if r.FormValue("NOVALUE") == "true" {
				if keys, cursorRet, err = hkey.HScanNoValues(cursor, match, count); err == nil {
					result = map[string]interface{}{"keys": keys, "cursor": cursorRet}
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
			var cursor uint64
			var count int64
			var match string
			if zkey, err = redisdb.GetHttpZSetKey(svcCtx.Key, svcCtx.RedisDataSource); err != nil {
			} else if cursor, err = strconv.ParseUint(r.FormValue("Cursor"), 10, 64); err != nil {
			} else if match = r.FormValue("Match"); match == "" {
			} else if count, err = strconv.ParseInt(r.FormValue("Count"), 10, 64); err != nil {
			} else {
				valInterface, cursor, err = zkey.ZScan(cursor, match, count)
				if err == nil {
					result = map[string]interface{}{"values": valInterface, "cursor": cursor}
				}
			}

		// --- LIST Operations ---
		case LRANGE:
			if !redisdb.IsAllowedListOp(svcCtx.Key, redisdb.LRange) {
				goto disallowedPermission
			}
			var start, stop int64
			if lKey, err = redisdb.GetHttpListKey(svcCtx.Key, svcCtx.RedisDataSource); err != nil {
			} else if start, err = strconv.ParseInt(r.FormValue("Start"), 10, 64); err != nil {
				result, err = "", errors.New("parse start error:"+err.Error())
			} else if stop, err = strconv.ParseInt(r.FormValue("Stop"), 10, 64); err != nil {
				result, err = "", errors.New("parse stop error:"+err.Error())
			} else {
				result, err = lKey.LRange(start, stop)
			}
		case LINDEX:
			if !redisdb.IsAllowedListOp(svcCtx.Key, redisdb.LIndex) {
				goto disallowedPermission
			}
			var index int64
			if lKey, err = redisdb.GetHttpListKey(svcCtx.Key, svcCtx.RedisDataSource); err != nil {
			} else if index, err = strconv.ParseInt(r.FormValue("Index"), 10, 64); err != nil {
				result, err = "", errors.New("parse index error:"+err.Error())
			} else {
				result, err = lKey.LIndex(index)
			}
		case LPOP:
			if !redisdb.IsAllowedListOp(svcCtx.Key, redisdb.LPop) {
				goto disallowedPermission
			}
			if lKey, err = redisdb.GetHttpListKey(svcCtx.Key, svcCtx.RedisDataSource); err == nil {
				result, err = lKey.LPop()
			}
		case RPOP:
			if !redisdb.IsAllowedListOp(svcCtx.Key, redisdb.RPop) {
				goto disallowedPermission
			}
			if lKey, err = redisdb.GetHttpListKey(svcCtx.Key, svcCtx.RedisDataSource); err == nil {
				result, err = lKey.RPop()
			}
		case LPUSH:
			if !redisdb.IsAllowedListOp(svcCtx.Key, redisdb.LPush) {
				goto disallowedPermission
			}
			if lKey, err = redisdb.GetHttpListKey(svcCtx.Key, svcCtx.RedisDataSource); err != nil {
			} else if result, err = svcCtx.ToValue(lKey, svcCtx.MsgpackBody(r, true)); err == nil {
				if err = lKey.LPush(result); err == nil {
					result = "true"
				} else {
					result = "false"
				}
			}
		case RPUSH:
			if !redisdb.IsAllowedListOp(svcCtx.Key, redisdb.RPush) {
				goto disallowedPermission
			}
			if lKey, err = redisdb.GetHttpListKey(svcCtx.Key, svcCtx.RedisDataSource); err != nil {
			} else if result, err = svcCtx.ToValue(lKey, svcCtx.MsgpackBody(r, true)); err == nil {
				if err = lKey.RPush(result); err == nil {
					result = "true"
				} else {
					result = "false"
				}
			}
		case LPUSHX:
			if !redisdb.IsAllowedListOp(svcCtx.Key, redisdb.LPushX) {
				goto disallowedPermission
			}
			if lKey, err = redisdb.GetHttpListKey(svcCtx.Key, svcCtx.RedisDataSource); err != nil {
			} else if result, err = svcCtx.ToValue(lKey, svcCtx.MsgpackBody(r, true)); err == nil {
				if err = lKey.LPushX(result); err == nil {
					result = "true"
				} else {
					result = "false"
				}
			}
		case RPUSHX:
			if !redisdb.IsAllowedListOp(svcCtx.Key, redisdb.RPushX) {
				goto disallowedPermission
			}
			if lKey, err = redisdb.GetHttpListKey(svcCtx.Key, svcCtx.RedisDataSource); err != nil {
			} else if result, err = svcCtx.ToValue(lKey, svcCtx.MsgpackBody(r, true)); err == nil {
				if err = lKey.RPushX(result); err == nil {
					result = "true"
				} else {
					result = "false"
				}
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
			if lKey, err = redisdb.GetHttpListKey(svcCtx.Key, svcCtx.RedisDataSource); err != nil {
				goto responseHttp
			} else if result, err = svcCtx.ToValue(lKey, svcCtx.MsgpackBody(r, true)); err == nil {
				if err = lKey.LRem(count, result); err == nil {
					result = "true"
				}
			}
		case LSET:
			if !redisdb.IsAllowedListOp(svcCtx.Key, redisdb.LSet) {
				goto disallowedPermission
			}
			var index int64
			if index, err = strconv.ParseInt(r.FormValue("Index"), 10, 64); err != nil {
				err = errors.New("parse index error:" + err.Error())
				goto responseHttp
			}
			if lKey, err = redisdb.GetHttpListKey(svcCtx.Key, svcCtx.RedisDataSource); err != nil {
				goto responseHttp
			} else if result, err = svcCtx.ToValue(lKey, svcCtx.MsgpackBody(r, true)); err == nil {
				if err = lKey.LSet(index, result); err == nil {
					result = "true"
				}
			}
		case LTRIM:
			if !redisdb.IsAllowedListOp(svcCtx.Key, redisdb.LTrim) {
				goto disallowedPermission
			}
			var start, stop int64
			if lKey, err = redisdb.GetHttpListKey(svcCtx.Key, svcCtx.RedisDataSource); err != nil {
			} else if start, err = strconv.ParseInt(r.FormValue("Start"), 10, 64); err != nil {
				result, err = "false", errors.New("parse start error:"+err.Error())
			} else if stop, err = strconv.ParseInt(r.FormValue("Stop"), 10, 64); err != nil {
				result, err = "false", errors.New("parse stop error:"+err.Error())
			} else if err = lKey.LTrim(start, stop); err == nil {
				result = "true"
			}

		// --- STREAM Operations ---
		case XRANGE, XRANGEN:
			if !redisdb.IsAllowedStreamOp(svcCtx.Key, redisdb.XRange) {
				goto disallowedPermission
			}
			var start, stop string
			var count int64
			start = r.FormValue("Start")
			stop = r.FormValue("Stop")
			if start == "" {
				result, err = "false", errors.New("no Start")
			} else if stop == "" {
				result, err = "false", errors.New("no Stop")
			} else if streamKey, err = redisdb.GetHttpStreamKey(svcCtx.Key, svcCtx.RedisDataSource); err == nil {
				if svcCtx.Cmd == XRANGEN {
					if count, err = strconv.ParseInt(r.FormValue("Count"), 10, 64); err != nil {
						result, err = "false", errors.New("parse N error:"+err.Error())
					} else {
						result, err = streamKey.XRange(start, stop, count)
					}
				} else {
					result, err = streamKey.XRange(start, stop, 0)
				}
			}

		case XREVRANGE, XREVRANGEN:
			if !redisdb.IsAllowedStreamOp(svcCtx.Key, redisdb.XRange) {
				goto disallowedPermission
			}
			var start, stop string
			var count int64
			start = r.FormValue("Start")
			stop = r.FormValue("Stop")
			if start == "" {
				result, err = "false", errors.New("no Start")
			} else if stop == "" {
				result, err = "false", errors.New("no Stop")
			} else if streamKey, err = redisdb.GetHttpStreamKey(svcCtx.Key, svcCtx.RedisDataSource); err == nil {
				if svcCtx.Cmd == XREVRANGEN {
					if count, err = strconv.ParseInt(r.FormValue("Count"), 10, 64); err != nil {
						result, err = "false", errors.New("parse N error:"+err.Error())
					} else {
						result, err = streamKey.XRevRange(start, stop, count)
					}
				} else {
					result, err = streamKey.XRevRange(start, stop, 0)
				}
			}

		case XREAD:
			if !redisdb.IsAllowedStreamOp(svcCtx.Key, redisdb.XRead) {
				goto disallowedPermission
			}
			var count int64
			var block time.Duration
			if streamKey, err = redisdb.GetHttpStreamKey(svcCtx.Key, svcCtx.RedisDataSource); err != nil {
			} else if count, err = strconv.ParseInt(r.FormValue("Count"), 10, 64); err != nil {
				result, err = "false", errors.New("parse count error:"+err.Error())
			} else if block, err = time.ParseDuration(r.FormValue("Block")); err != nil {
				result, err = "false", errors.New("parse block error:"+err.Error())
			} else {
				// XRead 的 Streams 参数构造
				result, err = streamKey.XRead([]string{svcCtx.Key, r.FormValue("ID")}, count, block)
			}
		case XADD:
			if !redisdb.IsAllowedStreamOp(svcCtx.Key, redisdb.XAdd) {
				goto disallowedPermission
			}
			if id := r.FormValue("ID"); id == "" {
				result, err = "false", errors.New("no ID")
			} else if streamKey, err = redisdb.GetHttpStreamKey(svcCtx.Key, svcCtx.RedisDataSource); err == nil {
				// 传递 MsgPack body 作为 Values
				if _, err = streamKey.XAdd(id, svcCtx.MsgpackBody(r, true)); err != nil {
					result = "false"
				} else {
					result = "true" // 或返回 ID
				}
			}
		case XDEL:
			if !redisdb.IsAllowedStreamOp(svcCtx.Key, redisdb.XDel) {
				goto disallowedPermission
			}
			if id := r.FormValue("ID"); id == "" {
				result, err = "false", errors.New("no ID")
			} else if streamKey, err = redisdb.GetHttpStreamKey(svcCtx.Key, svcCtx.RedisDataSource); err == nil {
				if err = streamKey.XDel(id); err != nil {
					result = "false"
				} else {
					result = "true"
				}
			}

		// --- STRING Operations ---
		case GET:
			if !redisdb.IsAllowedStringOp(svcCtx.Key, redisdb.Get) {
				goto disallowedPermission
			}
			if strKey, err = redisdb.GetHttpStringKey(svcCtx.Key, svcCtx.RedisDataSource); err == nil {
				result, err = strKey.Get(svcCtx.Field())
			}
		case SET:
			if !redisdb.IsAllowedStringOp(svcCtx.Key, redisdb.Set) {
				goto disallowedPermission
			}
			if strKey, err = redisdb.GetHttpStringKey(svcCtx.Key, svcCtx.RedisDataSource); err != nil {
			} else if result, err = svcCtx.ToValue(strKey, svcCtx.MsgpackBody(r, true)); err == nil {
				if err = strKey.Set(svcCtx.Field(), result, 0); err == nil {
					result = "true"
				} else {
					result = "false"
				}
			}
		case DEL:
			if !redisdb.IsAllowedStringOp(svcCtx.Key, redisdb.Del) {
				goto disallowedPermission
			}
			// Del 在各接口中未统一，暂时使用通用方式或 string key
			// 假设使用 StringKey.Set 来覆盖，或者 RdsClient.Del
			// 既然要移除 RdsClient，建议在 IHttpStringKey 加上 Del，或者使用 Hash Del
			// 这里保留 RdsClient 以确保兼容，因为 Del 是通用命令
			if err = svcCtx.RdsClient.Del(ctx, svcCtx.Key).Err(); err == nil {
				result = "true"
			} else {
				result = "false"
			}

		// --- HASH Operations ---
		case HGET:
			if !redisdb.IsAllowedHashOp(svcCtx.Key, redisdb.HGet) {
				goto disallowedPermission
			}
			if hkey, err = redisdb.GetHttpHashKey(svcCtx.Key, svcCtx.RedisDataSource); err == nil {
				result, err = hkey.HGet(svcCtx.Field())
			}
		case HGETALL:
			if !redisdb.IsAllowedHashOp(svcCtx.Key, redisdb.HGetAll) {
				goto disallowedPermission
			}
			if hkey, err = redisdb.GetHttpHashKey(svcCtx.Key, svcCtx.RedisDataSource); err == nil {
				result, err = hkey.HGetAll()
			}
		case HMGET:
			if !redisdb.IsAllowedHashOp(svcCtx.Key, redisdb.HMGET) {
				goto disallowedPermission
			}
			if hkey, err = redisdb.GetHttpHashKey(svcCtx.Key, svcCtx.RedisDataSource); err == nil {
				var fields []interface{} = sliceToInterface(strings.Split(svcCtx.Field(), ","))
				result, err = hkey.HMGET(fields...)
			}
		case HSET:
			if !redisdb.IsAllowedHashOp(svcCtx.Key, redisdb.HSet) {
				goto disallowedPermission
			}
			if hkey, err = redisdb.GetHttpHashKey(svcCtx.Key, svcCtx.RedisDataSource); err != nil {
			} else if result, err = svcCtx.ToValue(hkey, svcCtx.MsgpackBody(r, true)); err == nil {
				_, err = hkey.HSet(svcCtx.Field(), result)
			}
		case HMSET:
			if !redisdb.IsAllowedHashOp(svcCtx.Key, redisdb.HSet) {
				goto disallowedPermission
			}
			result = 0
			valuesInMsgpack := []string{}
			if hkey, err = redisdb.GetHttpHashKey(svcCtx.Key, svcCtx.RedisDataSource); err != nil {
			} else if err = msgpack.Unmarshal(svcCtx.MsgpackBody(r, true), &valuesInMsgpack); err != nil {
			} else if len(valuesInMsgpack) != len(svcCtx.Fields) {
				err = errors.New("fields count mismatch")
			} else if values, err = svcCtx.ToValues(hkey, valuesInMsgpack); err == nil {
				for i, field := range svcCtx.Fields {
					_, err = hkey.HSet(field, values[i])
					if err != nil {
						break
					}
				}
				if err == nil {
					result = svcCtx.Fields
				}
			}
		case HDEL:
			if !redisdb.IsAllowedHashOp(svcCtx.Key, redisdb.HDel) {
				goto disallowedPermission
			}
			// HDel 之前没在 IHttpHashKey 定义，如果没加，这里用 RdsClient
			// 假设你加上了 HDel(fields ...string)
			// result, err = hkey.HDel(svcCtx.Field())
			// 兜底：
			result, err = svcCtx.RdsClient.HDel(svcCtx.Ctx, svcCtx.Key, svcCtx.Field()).Result()
		case HKEYS:
			if !redisdb.IsAllowedHashOp(svcCtx.Key, redisdb.HKeys) {
				goto disallowedPermission
			}
			if hkey, err = redisdb.GetHttpHashKey(svcCtx.Key, svcCtx.RedisDataSource); err == nil {
				result, err = hkey.HKeys()
			}
		case HVALS:
			if !redisdb.IsAllowedHashOp(svcCtx.Key, redisdb.HVals) {
				goto disallowedPermission
			}
			if hkey, err = redisdb.GetHttpHashKey(svcCtx.Key, svcCtx.RedisDataSource); err == nil {
				result, err = hkey.HVals()
			}
		case HEXISTS:
			if !redisdb.IsAllowedHashOp(svcCtx.Key, redisdb.HExists) {
				goto disallowedPermission
			}
			if hkey, err = redisdb.GetHttpHashKey(svcCtx.Key, svcCtx.RedisDataSource); err == nil {
				result, err = hkey.HExists(svcCtx.Field())
			}
		case HRANDFIELD:
			if !redisdb.IsAllowedHashOp(svcCtx.Key, redisdb.HRandField) {
				goto disallowedPermission
			}
			var count int
			if hkey, err = redisdb.GetHttpHashKey(svcCtx.Key, svcCtx.RedisDataSource); err != nil {
			} else if count, err = strconv.Atoi(r.FormValue("Count")); err != nil {
				result, err = "", errors.New("parse count error:"+err.Error())
			} else {
				result, err = hkey.HRandField(count)
			}
		case HRANDFIELDWITHVALUES:
			if !redisdb.IsAllowedHashOp(svcCtx.Key, redisdb.HRandField) {
				goto disallowedPermission
			}
			var count int
			if hkey, err = redisdb.GetHttpHashKey(svcCtx.Key, svcCtx.RedisDataSource); err != nil {
			} else if count, err = strconv.Atoi(r.FormValue("Count")); err != nil {
				result, err = "", errors.New("parse count error:"+err.Error())
			} else {
				result, err = hkey.HRandFieldWithValues(count)
			}
		case HINCRBY:
			if !redisdb.IsAllowedHashOp(svcCtx.Key, redisdb.HIncrBy) {
				goto disallowedPermission
			}
			// 接口未定义，兜底
			var incr int64
			if incr, err = strconv.ParseInt(r.FormValue("Incr"), 10, 64); err != nil {
				result, err = "false", errors.New("parse Incr error:"+err.Error())
			} else if err = svcCtx.RdsClient.HIncrBy(svcCtx.Ctx, svcCtx.Key, svcCtx.Field(), incr).Err(); err == nil {
				result = "true"
			}
		case HINCRBYFLOAT:
			if !redisdb.IsAllowedHashOp(svcCtx.Key, redisdb.HIncrByFloat) {
				goto disallowedPermission
			}
			var incr float64
			if incr, err = strconv.ParseFloat(r.FormValue("Incr"), 64); err != nil {
				result, err = "false", errors.New("parse Incr error:"+err.Error())
			} else if err = svcCtx.RdsClient.HIncrByFloat(svcCtx.Ctx, svcCtx.Key, svcCtx.Field(), incr).Err(); err == nil {
				result = "true"
			}

		case SISMEMBER:
			if !redisdb.IsAllowedSetOp(svcCtx.Key, redisdb.SIsMember) {
				goto disallowedPermission
			}
			if skey, err = redisdb.GetHttpSetKey(svcCtx.Key, svcCtx.RedisDataSource); err != nil {
			} else {
				// SIsMember 在接口中接收 interface{}
				result, err = skey.SIsMember(r.FormValue("Member"))
			}

		// --- ZSET Operations ---
		case ZRANGE:
			if !redisdb.IsAllowedZSetOp(svcCtx.Key, redisdb.ZRange) {
				goto disallowedPermission
			}
			var start, stop int64
			if zkey, err = redisdb.GetHttpZSetKey(svcCtx.Key, svcCtx.RedisDataSource); err != nil {
			} else if start, err = strconv.ParseInt(r.FormValue("Start"), 10, 64); err != nil {
			} else if stop, err = strconv.ParseInt(r.FormValue("Stop"), 10, 64); err != nil {
			} else if r.FormValue("WITHSCORES") == "true" {
				var scores []float64
				var members interface{}
				if members, scores, err = zkey.ZRangeWithScores(start, stop); err == nil {
					result = map[string]interface{}{"members": members, "scores": scores}
				}
			} else {
				result, err = zkey.ZRange(start, stop)
			}

		case ZRANGEBYSCORE:
			if !redisdb.IsAllowedZSetOp(svcCtx.Key, redisdb.ZRangeByScore) {
				goto disallowedPermission
			}
			var offset, count int64
			var scores []float64
			var members interface{}
			if zkey, err = redisdb.GetHttpZSetKey(svcCtx.Key, svcCtx.RedisDataSource); err != nil {
			} else if Min := r.FormValue("Min"); Min == "" {
				result, err = "false", errors.New("no Min")
			} else if Max := r.FormValue("Max"); Max == "" {
				result, err = "false", errors.New("no Max")
			} else if offset, err = strconv.ParseInt(r.FormValue("Offset"), 10, 64); err != nil {
				result, err = "false", errors.New("parse offset error:"+err.Error())
			} else if count, err = strconv.ParseInt(r.FormValue("Count"), 10, 64); err != nil {
				result, err = "false", errors.New("parse count error:"+err.Error())
			} else {
				opt := &redis.ZRangeBy{Min: Min, Max: Max, Offset: offset, Count: count}
				if r.FormValue("WITHSCORES") == "true" {
					if members, scores, err = zkey.ZRangeByScoreWithScores(opt); err == nil {
						result = map[string]interface{}{"members": members, "scores": scores}
					}
				} else {
					result, err = zkey.ZRangeByScore(opt)
				}
			}

		case ZREVRANGE:
			if !redisdb.IsAllowedZSetOp(svcCtx.Key, redisdb.ZRevRange) {
				goto disallowedPermission
			}
			var start, stop int64
			if zkey, err = redisdb.GetHttpZSetKey(svcCtx.Key, svcCtx.RedisDataSource); err != nil {
			} else if start, err = strconv.ParseInt(r.FormValue("Start"), 10, 64); err != nil {
			} else if stop, err = strconv.ParseInt(r.FormValue("Stop"), 10, 64); err != nil {
			} else if r.FormValue("WITHSCORES") == "true" {
				var scores []float64
				var members interface{}
				if members, scores, err = zkey.ZRevRangeWithScores(start, stop); err == nil {
					result = map[string]interface{}{"members": members, "scores": scores}
				}
			} else {
				result, err = zkey.ZRevRange(start, stop)
			}

		case ZREVRANGEBYSCORE:
			if !redisdb.IsAllowedZSetOp(svcCtx.Key, redisdb.ZRevRangeByScore) {
				goto disallowedPermission
			}
			var offset, count int64
			var scores []float64
			var members interface{}
			if zkey, err = redisdb.GetHttpZSetKey(svcCtx.Key, svcCtx.RedisDataSource); err != nil {
			} else if Min, Max := r.FormValue("Min"), r.FormValue("Max"); Min == "" || Max == "" {
				result, err = "", errors.New("no Min or Max")
			} else if offset, err = strconv.ParseInt(r.FormValue("Offset"), 10, 64); err != nil {
				result, err = "", errors.New("parse offset error:"+err.Error())
			} else if count, err = strconv.ParseInt(r.FormValue("Count"), 10, 64); err != nil {
				result, err = "", errors.New("parse count error:"+err.Error())
			} else {
				opt := &redis.ZRangeBy{Min: Min, Max: Max, Offset: offset, Count: count}
				if r.FormValue("WITHSCORES") == "true" {
					if members, scores, err = zkey.ZRevRangeByScoreWithScores(opt); err == nil {
						result = map[string]interface{}{"members": members, "scores": scores}
					}
				} else {
					result, err = zkey.ZRevRangeByScore(opt)
				}
			}

		case ZRANK:
			if !redisdb.IsAllowedZSetOp(svcCtx.Key, redisdb.ZRank) {
				goto disallowedPermission
			}
			if zkey, err = redisdb.GetHttpZSetKey(svcCtx.Key, svcCtx.RedisDataSource); err == nil {
				result, err = zkey.ZRank(r.FormValue("Member"))
			}
		case ZCOUNT:
			if !redisdb.IsAllowedZSetOp(svcCtx.Key, redisdb.ZCount) {
				goto disallowedPermission
			}
			if zkey, err = redisdb.GetHttpZSetKey(svcCtx.Key, svcCtx.RedisDataSource); err == nil {
				result, err = zkey.ZCount(r.FormValue("Min"), r.FormValue("Max"))
			}
		case ZSCORE:
			if !redisdb.IsAllowedZSetOp(svcCtx.Key, redisdb.ZScore) {
				goto disallowedPermission
			}
			if zkey, err = redisdb.GetHttpZSetKey(svcCtx.Key, svcCtx.RedisDataSource); err == nil {
				result, err = zkey.ZScore(r.FormValue("Member"))
			}

		case ZADD:
			if !redisdb.IsAllowedZSetOp(svcCtx.Key, redisdb.ZAdd) {
				goto disallowedPermission
			}
			var Score float64
			if zkey, err = redisdb.GetHttpZSetKey(svcCtx.Key, svcCtx.RedisDataSource); err != nil {
			} else if result, err = svcCtx.ToValue(zkey, svcCtx.MsgpackBody(r, true)); err != nil {
			} else if Score, err = strconv.ParseFloat(r.FormValue("Score"), 64); err != nil {
				result, err = "false", errors.New("parameter Score should be float")
			} else if err = zkey.ZAdd(redis.Z{Score: Score, Member: result}); err == nil {
				result = "true"
			} else {
				result = "false"
			}

		case ZREM:
			if !redisdb.IsAllowedZSetOp(svcCtx.Key, redisdb.ZRem) {
				goto disallowedPermission
			}
			MemberStr := strings.Split(r.FormValue("Member"), ",")
			if len(MemberStr) == 0 {
				err = errors.New("no Member")
				goto responseHttp
			}
			Members := make([]interface{}, len(MemberStr))
			for i, v := range MemberStr {
				Members[i] = v
			}
			if zkey, err = redisdb.GetHttpZSetKey(svcCtx.Key, svcCtx.RedisDataSource); err != nil {
			} else if err = zkey.ZRem(Members...); err == nil {
				result = "true"
			} else {
				result = "false"
			}

		case ZREMRANGEBYSCORE:
			if !redisdb.IsAllowedZSetOp(svcCtx.Key, redisdb.ZRemRangeByScore) {
				goto disallowedPermission
			}
			// 接口未定义，兜底
			if Min, Max := r.FormValue("Min"), r.FormValue("Max"); Min == "" || Max == "" {
				result, err = "false", errors.New("no Min or Max")
			} else if err = svcCtx.RdsClient.ZRemRangeByScore(svcCtx.Ctx, svcCtx.Key, Min, Max).Err(); err == nil {
				result = "true"
			}

		case ZINCRBY:
			if !redisdb.IsAllowedZSetOp(svcCtx.Key, redisdb.ZIncrBy) {
				goto disallowedPermission
			}
			var incr float64
			var member string
			if member = r.FormValue("Member"); member == "" {
				result, err = "false", errors.New("no Member")
			} else if incr, err = strconv.ParseFloat(r.FormValue("Incr"), 64); err != nil {
				result, err = "false", errors.New("parse Incr error:"+err.Error())
			} else if zkey, err = redisdb.GetHttpZSetKey(svcCtx.Key, svcCtx.RedisDataSource); err == nil {
				var newScore float64
				if newScore, err = zkey.ZIncrBy(incr, member); err == nil {
					result = newScore
				}
			}

		// --- Vector/RediSearch (Example) ---
		case "SEARCH":
			if vectorKey, err = redisdb.GetHttpVectorSetKey(svcCtx.Key, svcCtx.RedisDataSource); err == nil {
				var count int64
				var docs interface{}
				if count, docs, err = vectorKey.Search(r.FormValue("Q")); err == nil {
					result = map[string]interface{}{"count": count, "docs": docs}
				}
			}

		// --- Common Keys Ops ---
		case TYPE:
			if !redisdb.IsAllowedCommon(svcCtx.Key, redisdb.Type) {
				goto disallowedPermission
			}
			result, err = svcCtx.RdsClient.Type(svcCtx.Ctx, svcCtx.Key).Result()
		case EXPIRE:
			if !redisdb.IsAllowedCommon(svcCtx.Key, redisdb.Expire) {
				goto disallowedPermission
			}
			var seconds int64
			if seconds, err = strconv.ParseInt(r.FormValue("Seconds"), 10, 64); err != nil {
				result, err = "false", errors.New("parse seconds error:"+err.Error())
			} else if err = svcCtx.RdsClient.Expire(svcCtx.Ctx, svcCtx.Key, time.Duration(seconds)*time.Second).Err(); err == nil {
				result = "true"
			}
		case EXPIREAT:
			if !redisdb.IsAllowedCommon(svcCtx.Key, redisdb.Expire) {
				goto disallowedPermission
			}
			var timestamp int64
			if timestamp, err = strconv.ParseInt(r.FormValue("Timestamp"), 10, 64); err != nil {
				result, err = "false", errors.New("parse seconds error:"+err.Error())
			} else if err = svcCtx.RdsClient.ExpireAt(svcCtx.Ctx, svcCtx.Key, time.Unix(timestamp, 0)).Err(); err == nil {
				result = "true"
			}
		case PERSIST:
			if !redisdb.IsAllowedCommon(svcCtx.Key, redisdb.Persist) {
				goto disallowedPermission
			}
			if err = svcCtx.RdsClient.Persist(svcCtx.Ctx, svcCtx.Key).Err(); err == nil {
				result = "true"
			}
		case TTL:
			if !redisdb.IsAllowedCommon(svcCtx.Key, redisdb.TTL) {
				goto disallowedPermission
			}
			result, err = svcCtx.RdsClient.TTL(svcCtx.Ctx, svcCtx.Key).Result()
		case PTTL:
			if !redisdb.IsAllowedCommon(svcCtx.Key, redisdb.TTL) {
				goto disallowedPermission
			}
			result, err = svcCtx.RdsClient.PTTL(svcCtx.Ctx, svcCtx.Key).Result()
		case RENAME:
			if !redisdb.IsAllowedCommon(svcCtx.Key, redisdb.Rename) {
				goto disallowedPermission
			}
			if newKey := r.FormValue("NewKey"); newKey == "" {
				result, err = "false", errors.New("no NewKey")
			} else if err = svcCtx.RdsClient.Rename(svcCtx.Ctx, svcCtx.Key, newKey).Err(); err == nil {
				result = "true"
			}
		case RENAMEX:
			if !redisdb.IsAllowedCommon(svcCtx.Key, redisdb.Rename) {
				goto disallowedPermission
			}
			if newKey := r.FormValue("NewKey"); newKey == "" {
				result, err = "false", errors.New("no NewKey")
			} else if err = svcCtx.RdsClient.RenameNX(svcCtx.Ctx, svcCtx.Key, newKey).Err(); err == nil {
				result = "true"
			}
		case EXISTS:
			if !redisdb.IsAllowedCommon(svcCtx.Key, redisdb.Exists) {
				goto disallowedPermission
			}
			result, err = svcCtx.RdsClient.Exists(svcCtx.Ctx, svcCtx.Key).Result()
		case TIME:
			if !redisdb.IsAllowedDBOp(redisdb.DBTime) {
				goto disallowedPermission
			}
			var tm time.Time
			if tm, err = svcCtx.RdsClient.Time(svcCtx.Ctx).Result(); err == nil {
				result = tm.UnixMilli()
			}
		case KEYS:
			if !redisdb.IsAllowedDBOp(redisdb.DBKeys) {
				goto disallowedPermission
			}
			result, err = svcCtx.RdsClient.Keys(svcCtx.Ctx, svcCtx.Key).Result()

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
					var dst *bytes.Buffer = bytes.NewBuffer([]byte{})
					if err = json.Compact(dst, bs); err == nil {
						bs = dst.Bytes()
					}
				}
			}
		}
		// Error handling
		if err != nil && httpStatus == http.StatusOK {
			httpStatus = http.StatusInternalServerError
		}
		// Content-Type
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
		WriteTimeout:      50 * time.Second,
		IdleTimeout:       15 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		logger.Error().Err(err).Msg("http server ListenAndServe error")
		return
	}
	logger.Info().Any("port", port).Any("path", path).Msg("doptime http server started!")
}

func init() {
	logger.Info().Any("port", cfghttp.Port).Any("path", cfghttp.Path).Msg("doptime http server is starting")
	go httpStart(cfghttp.Path, cfghttp.Port)
}

var ErrOperationNotPermited = errors.New("error operation permission denied. In develop stage, turn on DangerousAutoWhitelist in toml to auto permit")

var ErrBadCommand = errors.New("error bad command")
