package httpserve

import (
	"net/http"
	"strings"

	"github.com/redis/go-redis/v9"
)

func (svcCtx *DoptimeReqCtx) DelHandler(r *http.Request, rds *redis.Client) (result interface{}, err error) {
	switch svcCtx.Cmd {
	case "HDEL":
		//error if empty Key or Field
		if svcCtx.Field == "" {
			return "false", ErrEmptyKeyOrField
		}
		cmd := rds.HDel(svcCtx.Ctx, svcCtx.Key, svcCtx.Field)
		if err = cmd.Err(); err == nil {
			return "true", nil
		}
		return "false", err
	case "DEL":
		cmd := rds.HDel(svcCtx.Ctx, svcCtx.Key, "del")
		if err = cmd.Err(); err == nil {
			return "true", nil
		}
		return "false", err
	case "ZREM":
		var MemberStr = strings.Split(r.FormValue("Member"), ",")
		//convert Member to []interface{}
		var Member = make([]interface{}, len(MemberStr))
		for i, v := range MemberStr {
			Member[i] = v
		}
		if err = rds.ZRem(svcCtx.Ctx, svcCtx.Key, Member...).Err(); err == nil {
			return "true", nil
		}
		return "false", err
	case "ZREMRANGEBYSCORE":
		var Min = r.FormValue("Min")
		var Max = r.FormValue("Max")
		if err = rds.ZRemRangeByScore(svcCtx.Ctx, svcCtx.Key, Min, Max).Err(); err == nil {
			return "true", nil
		}
		return "false", err
	default:
		return nil, ErrBadCommand
	}

}
