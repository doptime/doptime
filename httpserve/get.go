package httpserve

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/doptime/doptime/data"
	"github.com/redis/go-redis/v9"
)

func (svcCtx *HttpContext) GetHandler() (ret interface{}, err error) {
	var (
		Min, Max string
		tm       time.Time
		members  []interface{} = []interface{}{}
	)

	db := data.New[string, interface{}](data.Option.WithKey(svcCtx.Key).WithRds(svcCtx.RedisDataSource))
	//case Is a member of a set
	switch svcCtx.Cmd {
	// all data that appears in the form or body is json format, will be stored in paramIn["JsonPack"]
	// this is used to support 3rd party api
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
			start, stop      int64 = 0, -1
			memberScoreSlice []interface{}
		)
		if start, err = strconv.ParseInt(svcCtx.Req.FormValue("Start"), 10, 64); err != nil {
			return "", errors.New("parse start error:" + err.Error())
		}
		if stop, err = strconv.ParseInt(svcCtx.Req.FormValue("Stop"), 10, 64); err != nil {
			return "", errors.New("parse stop error:" + err.Error())
		}
		// ZREVRANGE key start stop [WITHSCORES==true]
		if svcCtx.Req.FormValue("WITHSCORES") == "true" {
			cmd := db.Rds.ZRevRangeWithScores(context.Background(), db.Key, start, stop)
			if rlts, err := cmd.Result(); err != nil {
				return "", cmd.Err()
			} else {
				for _, rlt := range rlts {
					memberScoreSlice = append(memberScoreSlice, rlt.Member)
					memberScoreSlice = append(memberScoreSlice, rlt.Score)
				}
			}
			return memberScoreSlice, nil
		}
		// ZREVRANGE key start stop [WITHSCORES==false]
		return db.ZRevRange(start, stop)
	case "ZREVRANGEBYSCORE":
		var (
			offset, count    int64 = 0, -1
			memberScoreSlice []interface{}
		)
		if Min, Max = svcCtx.Req.FormValue("Min"), svcCtx.Req.FormValue("Max"); Min == "" || Max == "" {
			return "", errors.New("no Min or Max")
		}
		if offset, err = strconv.ParseInt(svcCtx.Req.FormValue("Offset"), 10, 64); err != nil {
			return "", errors.New("parse offset error:" + err.Error())
		}
		if count, err = strconv.ParseInt(svcCtx.Req.FormValue("Count"), 10, 64); err != nil {
			return "", errors.New("parse count error:" + err.Error())
		}
		//ZREVRANGEBYSCORE key max min [WITHSCORES==true]
		if svcCtx.Req.FormValue("WITHSCORES") == "true" {

			cmd := db.Rds.ZRevRangeByScoreWithScores(context.Background(), db.Key, &redis.ZRangeBy{Min: Min, Max: Max, Offset: offset, Count: count})
			if rlts, err := cmd.Result(); err != nil {
				return "", cmd.Err()
			} else {
				for _, rlt := range rlts {
					memberScoreSlice = append(memberScoreSlice, rlt.Member)
					memberScoreSlice = append(memberScoreSlice, rlt.Score)
				}
			}
			return memberScoreSlice, nil
		}
		//ZREVRANGEBYSCORE key max min [WITHSCORES==false]
		return db.ZRevRangeByScore(&redis.ZRangeBy{Min: Min, Max: Max, Offset: offset, Count: count})
	case "ZCARD":
		return db.ZCard()
	case "ZRANK":
		return db.ZRank(svcCtx.Req.FormValue("Member"))
	case "ZCOUNT":
		return db.ZCount(svcCtx.Req.FormValue("Min"), svcCtx.Req.FormValue("Max"))
	case "ZSCORE":
		return db.ZScore(svcCtx.Req.FormValue("Member"))
	case "SCAN":
		var (
			cursor uint64
			count  int64
			keys   []string
			match  string
		)
		if cursor, err = strconv.ParseUint(svcCtx.Req.FormValue("Cursor"), 10, 64); err != nil {
			return "", errors.New("parse cursor error:" + err.Error())
		}
		if match = svcCtx.Req.FormValue("Match"); match == "" {
			return "", errors.New("no Match")
		}
		if count, err = strconv.ParseInt(svcCtx.Req.FormValue("Count"), 10, 64); err != nil {
			return "", errors.New("parse count error:" + err.Error())
		}
		if keys, cursor, err = db.Scan(cursor, match, count); err != nil {
			return "", err
		}
		return json.Marshal(map[string]interface{}{"keys": keys, "cursor": cursor})

	//case default
	default:
		return nil, ErrBadCommand
	}

}
