package rdsdb

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/doptime/doptime/config"
	"github.com/doptime/doptime/dlog"
	"github.com/doptime/doptime/specification"
	"github.com/redis/go-redis/v9"
)

type Ctx[k comparable, v any] struct {
	Context         context.Context
	RdsName         string
	Rds             *redis.Client
	Key             string
	MarshalValue    func(value v) (valueStr string, err error)
	UnmarshalValue  func(valbytes []byte) (value v, err error)
	UnmarshalValues func(valStrs []string) (values []v, err error)
}

func (ctx *Ctx[k, v]) clone(newKey string) (newCtx Ctx[k, v]) {
	if len(newKey) == 0 {
		newKey = ctx.Key
	}
	return Ctx[k, v]{ctx.Context, ctx.RdsName, ctx.Rds, newKey, ctx.MarshalValue, ctx.UnmarshalValue, ctx.UnmarshalValues}
}

func NonKey[k comparable, v any](ops ...*DataOption) *Ctx[k, v] {
	ctx := &Ctx[k, v]{}
	if err := ctx.useOption(ops...); err != nil {
		dlog.Error().Err(err).Msg("data.New failed")
		return nil
	}
	return ctx
}
func (ctx *Ctx[k, v]) Time() (tm time.Time, err error) {
	cmd := ctx.Rds.Time(ctx.Context)
	return cmd.Result()
}

// sacn key by pattern
func (ctx *Ctx[k, v]) Scan(cursorOld uint64, match string, count int64) (keys []string, cursorNew uint64, err error) {
	var (
		cmd   *redis.ScanCmd
		_keys []string
	)
	//scan all keys
	for {

		if cmd = ctx.Rds.Scan(ctx.Context, cursorOld, match, count); cmd.Err() != nil {
			return nil, 0, cmd.Err()
		}
		if _keys, cursorNew, err = cmd.Result(); err != nil {
			return nil, 0, err
		}
		keys = append(keys, _keys...)
		if cursorNew == 0 {
			break
		}
	}
	return keys, cursorNew, nil
}

func (ctx *Ctx[k, v]) useOption(ops ...*DataOption) error {
	if len(ops) > 0 {
		ctx.Key = ops[0].Key
		ctx.RdsName = ops[0].DataSource
	}
	if len(ctx.Key) == 0 && !specification.GetValidDataKeyName((*v)(nil), &ctx.Key) {
		return fmt.Errorf("invalid keyname infered from type: " + reflect.TypeOf((*v)(nil)).String())
	}
	var exists bool
	if ctx.Rds, exists = config.Rds[ctx.RdsName]; !exists {
		return fmt.Errorf("Rds item unconfigured: " + ctx.RdsName)
	}
	ctx.Context = context.Background()
	ctx.MarshalValue = ctx.toValueStrFun()
	ctx.UnmarshalValue = ctx.toValueFunc()
	ctx.UnmarshalValues = ctx.toValuesFunc()
	return nil
}
