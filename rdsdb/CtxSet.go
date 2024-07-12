package rdsdb

import (
	"github.com/doptime/doptime/dlog"
)

type CtxSet[k comparable, v any] struct {
	Ctx[k, v]
}

func SetKey[k comparable, v any](ops ...*DataOption) *CtxSet[k, v] {
	ctx := &CtxSet[k, v]{}
	if err := ctx.useOption(ops...); err != nil {
		dlog.Error().Err(err).Msg("data.New failed")
		return nil
	}
	if len(ops) > 0 && ops[0].RegisterWebData {
		ctx.RegisterWebData("set")
	}
	return ctx
}

func (ctx *CtxSet[k, v]) ConcatKey(fields ...interface{}) *CtxSet[k, v] {
	return &CtxSet[k, v]{ctx.clone(ConcatedKeys(ctx.Key, fields...))}
}

// append to Set
func (ctx *CtxSet[k, v]) SAdd(param v) (err error) {
	valStr, err := ctx.MarshalValue(param)
	if err != nil {
		return err
	}
	return ctx.Rds.SAdd(ctx.Context, ctx.Key, valStr).Err()
}

func (ctx *CtxSet[k, v]) SCard() (int64, error) {
	return ctx.Rds.SCard(ctx.Context, ctx.Key).Result()
}

func (ctx *CtxSet[k, v]) SRem(param v) error {
	valStr, err := ctx.MarshalValue(param)
	if err != nil {
		return err
	}
	return ctx.Rds.SRem(ctx.Context, ctx.Key, valStr).Err()
}
func (ctx *CtxSet[k, v]) SIsMember(param v) (bool, error) {
	valStr, err := ctx.MarshalValue(param)
	if err != nil {
		return false, err
	}
	return ctx.Rds.SIsMember(ctx.Context, ctx.Key, valStr).Result()
}

func (ctx *CtxSet[k, v]) SMembers() ([]v, error) {
	cmd := ctx.Rds.SMembers(ctx.Context, ctx.Key)
	if err := cmd.Err(); err != nil {
		return nil, err
	}
	return ctx.UnmarshalValues(cmd.Val())
}
func (ctx *CtxSet[k, v]) SScan(cursor uint64, match string, count int64) ([]v, uint64, error) {
	cmd := ctx.Rds.SScan(ctx.Context, ctx.Key, cursor, match, count)
	if err := cmd.Err(); err != nil {
		return nil, 0, err
	}
	Strs, cursor, err := cmd.Result()
	if err != nil {
		return nil, 0, err
	}
	values, err := ctx.UnmarshalValues(Strs)
	if err != nil {
		return nil, 0, err
	}
	return values, cursor, nil
}
