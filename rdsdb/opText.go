package rdsdb

import (
	"time"

	"github.com/redis/go-redis/v9"
)

func (ctx *Ctx[k, v]) Keys() (out []k, err error) {
	var (
		cmd  *redis.StringSliceCmd
		keys []string
	)
	cmd = ctx.Rds.Keys(ctx.Context, ctx.Key+":*")
	if keys, err = cmd.Result(); err != nil {
		return nil, err
	}
	return ctx.toKeys(keys)
}

// for the reason of protection, both ctx.Key & Key are required. the avoid set Hash table to the wrong type , and thus leading to data loss.
func (ctx *Ctx[k, v]) Set(key k, param v, expiration time.Duration) (err error) {
	var (
		keyStr string
		valStr string
	)
	if keyStr, err = ctx.toKeyStr(key); err != nil {
		return err
	}
	if valStr, err = ctx.MarshalValue(param); err != nil {
		return err
	} else {
		status := ctx.Rds.Set(ctx.Context, ctx.Key+":"+keyStr, valStr, expiration)
		return status.Err()
	}
}

// for the reason of protection, both ctx.Key & Key are required. the avoid set Hash table to the wrong type , and thus leading to data loss.
func (ctx *Ctx[k, v]) Del(key k) (err error) {
	var (
		keyStr string
	)
	if keyStr, err = ctx.toKeyStr(key); err != nil {
		return err
	}
	status := ctx.Rds.Del(ctx.Context, ctx.Key+":"+keyStr)
	return status.Err()
}
