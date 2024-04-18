package httpserve

import (
	"github.com/redis/go-redis/v9"
)

func (svcCtx *HttpContext) PutHandler(rds *redis.Client) (data interface{}, err error) {
	//use remote service map to handle request
	var (
		bytes []byte
	)

	switch svcCtx.Cmd {
	case "SET":
		if svcCtx.Key == "" || svcCtx.Field == "" {
			return "false", ErrEmptyKeyOrField
		}
		if bytes, err = svcCtx.MsgpackBody(); err != nil {
			return "false", err
		}
		cmd := rds.Set(svcCtx.Ctx, svcCtx.Key+":"+svcCtx.Field, bytes, 0)
		if err = cmd.Err(); err != nil {
			return "false", err
		}
		return "true", nil

	case "HSET":
		//error if empty Key or Field
		if svcCtx.Key == "" || svcCtx.Field == "" {
			return "false", ErrEmptyKeyOrField
		}
		if bytes, err = svcCtx.MsgpackBody(); err != nil {
			return "false", err
		}
		cmd := rds.HSet(svcCtx.Ctx, svcCtx.Key, svcCtx.Field, bytes)
		if err = cmd.Err(); err != nil {
			return "false", err
		}
		return "true", nil
	case "RPUSH":
		//error if empty Key or Field
		if svcCtx.Key == "" {
			return "false", ErrEmptyKeyOrField
		}
		if bytes, err = svcCtx.MsgpackBody(); err != nil {
			return "false", err
		}
		cmd := rds.RPush(svcCtx.Ctx, svcCtx.Key, bytes)
		if err = cmd.Err(); err != nil {
			return "false", err
		}
		return "true", nil
	default:
		return nil, ErrBadCommand
	}
}
