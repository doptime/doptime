package data

import "github.com/redis/go-redis/v9"

// append to Set
func (ctx *Ctx[k, v]) SAdd(param v) (err error) {
	valStr, err := ctx.toValueStr(param)
	if err != nil {
		return err
	}
	status := ctx.Rds.SAdd(ctx.Ctx, ctx.Key, valStr)
	return status.Err()
}
func (ctx *Ctx[k, v]) SRem(param v) (err error) {
	valStr, err := ctx.toValueStr(param)
	if err != nil {
		return err
	}
	status := ctx.Rds.SRem(ctx.Ctx, ctx.Key, valStr)
	return status.Err()
}
func (ctx *Ctx[k, v]) SIsMember(param v) (isMember bool, err error) {
	valStr, err := ctx.toValueStr(param)
	if err != nil {
		return false, err
	}
	status := ctx.Rds.SIsMember(ctx.Ctx, ctx.Key, valStr)
	return status.Result()
}
func (ctx *Ctx[k, v]) SMembers() (members []v, err error) {
	var cmd *redis.StringSliceCmd
	if cmd = ctx.Rds.SMembers(ctx.Ctx, ctx.Key); cmd.Err() != nil {
		return nil, cmd.Err()
	}
	return ctx.toValues(cmd.Val()...)
}
