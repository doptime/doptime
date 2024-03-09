package httpservice

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/vmihailenco/msgpack/v5"
	"github.com/yangkequn/goflow/api"
	"github.com/yangkequn/goflow/config"
	"github.com/yangkequn/goflow/data"
	"github.com/yangkequn/goflow/permission"
)

func (svcCtx *HttpContext) GetHandler() (ret interface{}, err error) {
	var (
		Min, Max  string
		tm        time.Time
		operation string
		members   []interface{} = []interface{}{}
		rds       *redis.Client
	)
	if rds, err = config.GetRdsClientByName(svcCtx.RedisDataSource); err != nil {
		return nil, err
	}

	if operation, err = svcCtx.KeyFieldAtJwt(); err != nil {
		return "", err
	}
	if !permission.IsPermitted(svcCtx.Key, operation) {
		// check operation permission
		return nil, fmt.Errorf(" operation %v not permitted", operation)
	}

	db := data.Ctx[string, interface{}]{Ctx: svcCtx.Ctx, Rds: rds, Key: svcCtx.Key}
	//case Is a member of a set
	switch svcCtx.Cmd {
	// all data that appears in the form or body is json format, will be stored in paramIn["JsonPack"]
	// this is used to support 3rd party api
	case "API":
		var (
			paramIn     map[string]interface{} = map[string]interface{}{}
			ServiceName string                 = svcCtx.Key
		)
		svcCtx.MergeJwtField(paramIn)
		//convert query fields to JsonPack. but ignore K field(api name )
		if svcCtx.Req.ParseForm(); len(svcCtx.Req.Form) > 0 {
			for key, value := range svcCtx.Req.Form {
				if paramIn[key] = value[0]; len(value) > 1 {
					paramIn[key] = value // Assign the single value directly
				}
			}
		}
		if msgPack := svcCtx.MsgpackBodyBytes(); len(msgPack) > 0 {
			if err = msgpack.Unmarshal(msgPack, &paramIn); err != nil {
				return nil, fmt.Errorf("msgpack.Unmarshal msgPack error %s", err)
			}
			//paramIn["MsgpackBody"] = msgPack
		} else if jsonBody := svcCtx.JsonBodyBytes(); len(jsonBody) > 0 {
			//convert to msgpack, so that fields can be renamed in ProcessOneJob
			if err = json.Unmarshal(jsonBody, &paramIn); err != nil {
				return nil, fmt.Errorf("msgpack.Unmarshal JsonBody error %s", err)
			}
		}
		return api.CallByHTTP(ServiceName, paramIn, svcCtx.Req)

	case "GET":
		return db.Get(svcCtx.Field)
	case "HGET":
		return db.HGet(svcCtx.Field)
	case "HGETALL":
		return db.HGetAll()
	case "HMGET":
		return db.HMGET(strings.Split(svcCtx.Field, ",")...)
	case "HKEYS":
		return db.HKeys()
	case "HEXISTS":
		return db.HExists(svcCtx.Field)
	case "HRANDFIELD":
		var count int
		if count, err = strconv.Atoi(svcCtx.Req.FormValue("Count")); err != nil {
			return "", errors.New("parse count error:" + err.Error())
		}
		return db.HRandField(count)
	case "HLEN":
		return db.HLen()
	case "HVALS":
		var values []interface{}
		if values, err = db.HVals(); err != nil {
			return "", err
		}
		return values, nil
	case "SISMEMBER":
		return db.SIsMember(svcCtx.Req.FormValue("Member"))
	case "TIME":
		if tm, err = db.Time(); err != nil {
			return "", err
		}
		return tm.UnixMilli(), nil
	case "ZRANGE":
		var (
			start, stop int64 = 0, -1
		)
		if start, err = strconv.ParseInt(svcCtx.Req.FormValue("Start"), 10, 64); err != nil {
			return "", errors.New("parse start error:" + err.Error())
		}
		if stop, err = strconv.ParseInt(svcCtx.Req.FormValue("Stop"), 10, 64); err != nil {
			return "", errors.New("parse stop error:" + err.Error())
		}
		// ZRANGE key start stop [WITHSCORES==true]
		if svcCtx.Req.FormValue("WITHSCORES") == "true" {
			var scores []float64
			if members, scores, err = db.ZRangeWithScores(start, stop); err != nil {
				return "", err
			}
			return json.Marshal(map[string]interface{}{"members": members, "scores": scores})
		}
		// ZRANGE key start stop [WITHSCORES==false]
		if members, err = db.ZRange(start, stop); err != nil {
			return "", err
		}
		return json.Marshal(members)
	case "ZRANGEBYSCORE":
		var (
			offset, count int64 = 0, -1
			scores        []float64
		)
		if Min, Max = svcCtx.Req.FormValue("Min"), svcCtx.Req.FormValue("Max"); Min == "" || Max == "" {
			return "", errors.New("no Min or Max")
		}
		//ZRANGEBYSCORE key min max [WITHSCORES==true]
		if svcCtx.Req.FormValue("WITHSCORES") == "true" {
			if members, scores, err = db.ZRangeByScoreWithScores(&redis.ZRangeBy{Min: Min, Max: Max, Offset: offset, Count: count}); err != nil {
				return "", err
			}
			//marshal result to json
			return json.Marshal(map[string]interface{}{"members": members, "scores": scores})
		}
		//ZRANGEBYSCORE key min max [WITHSCORES==false]
		if members, err = db.ZRangeByScore(&redis.ZRangeBy{Min: Min, Max: Max, Offset: offset, Count: count}); err != nil {
			return "", err
		}
		return json.Marshal(members)
	case "ZREVRANGE":
		var (
			start, stop int64 = 0, -1
		)
		if start, err = strconv.ParseInt(svcCtx.Req.FormValue("Start"), 10, 64); err != nil {
			return "", errors.New("parse start error:" + err.Error())
		}
		if stop, err = strconv.ParseInt(svcCtx.Req.FormValue("Stop"), 10, 64); err != nil {
			return "", errors.New("parse stop error:" + err.Error())
		}
		// ZREVRANGE key start stop [WITHSCORES==true]
		if svcCtx.Req.FormValue("WITHSCORES") == "true" {
			var scores []float64
			if members, scores, err = db.ZRevRangeWithScores(start, stop); err != nil {
				return "", err
			}
			return json.Marshal(map[string]interface{}{"members": members, "scores": scores})
		}
		// ZREVRANGE key start stop [WITHSCORES==false]
		if members, err = db.ZRevRange(start, stop); err != nil {
			return "", err
		}
		return json.Marshal(members)
	case "ZREVRANGEBYSCORE":
		var (
			offset, count int64 = 0, -1
			scores        []float64
		)
		if Min, Max = svcCtx.Req.FormValue("Min"), svcCtx.Req.FormValue("Max"); Min == "" || Max == "" {
			return "", errors.New("no Min or Max")
		}
		//ZREVRANGEBYSCORE key max min [WITHSCORES==true]
		if svcCtx.Req.FormValue("WITHSCORES") == "true" {
			if members, scores, err = db.ZRevRangeByScoreWithScores(&redis.ZRangeBy{Min: Min, Max: Max, Offset: offset, Count: count}); err != nil {
				return "", err
			}
			//marshal result to json
			return json.Marshal(map[string]interface{}{"members": members, "scores": scores})
		}
		//ZREVRANGEBYSCORE key max min [WITHSCORES==false]
		if members, err = db.ZRevRangeByScore(&redis.ZRangeBy{Min: Min, Max: Max, Offset: offset, Count: count}); err != nil {
			return "", err
		}
		return json.Marshal(members)
	case "ZCARD":
		return db.ZCard()
	case "ZRANK":
		return db.ZRank(svcCtx.Req.FormValue("Member"))
	case "ZCOUNT":
		return db.ZCount(svcCtx.Req.FormValue("Min"), svcCtx.Req.FormValue("Max"))
	case "ZSCORE":
		return db.ZScore(svcCtx.Req.FormValue("Member"))
	//case default
	default:
		return nil, ErrBadCommand
	}

}
