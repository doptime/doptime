package rdsdb

import (
	"context"
	"fmt"
	"time"

	"github.com/doptime/doptime/config"
	"github.com/doptime/doptime/dlog"
	"github.com/doptime/doptime/specification"
	"github.com/redis/go-redis/v9"
)

type Ctx[k comparable, v any] struct {
	Context context.Context
	RdsName string
	Rds     *redis.Client

	Key     string
	KeyType string

	Moder           *StructModifiers[v]
	MarshalValue    func(value v) (valueStr string, err error)
	UnmarshalValue  func(valbytes []byte) (value v, err error)
	UnmarshalValues func(valStrs []string) (values []v, err error)
}

func (ctx *Ctx[k, v]) Duplicate(newKey, RdsSourceName string) (newCtx Ctx[k, v]) {
	return Ctx[k, v]{ctx.Context, RdsSourceName, ctx.Rds, newKey, ctx.KeyType, ctx.Moder, ctx.MarshalValue, ctx.UnmarshalValue, ctx.UnmarshalValues}
}

func NonKey[k comparable, v any](ops ...opSetter) *Ctx[k, v] {
	ctx := &Ctx[k, v]{KeyType: "nonkey"}
	op := Option{}.buildOptions(ops...)
	if err := ctx.applyOption(op); err != nil {
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

func (ctx *Ctx[k, v]) applyOption(opt *Option) (err error) {
	ctx.Key = opt.Key
	ctx.RdsName = opt.DataSource

	if opt.RegisterWebData {
		ctx.RegisterWebData()
	}
	if len(ctx.Key) == 0 {
		ctx.Key, err = specification.GetValidDataKeyName((*v)(nil))
		if err != nil {
			return err
		}
	}
	var exists bool
	if ctx.Rds, exists = config.Rds[ctx.RdsName]; !exists {
		return fmt.Errorf("rds item unconfigured: " + ctx.RdsName)
	}
	ctx.Context = context.Background()
	ctx.MarshalValue = ctx.toValueStrFun()
	ctx.UnmarshalValue = ctx.toValueFunc()
	ctx.UnmarshalValues = ctx.toValuesFunc()
	ctx.Moder = RegisterStructModifiers[v](opt.Modifiers)
	return nil
}

func (ctx *Ctx[k, v]) toKeyValueStrs(keyValue ...interface{}) (keyValStrs []string, err error) {
	var (
		key              k
		value            v
		strkey, strvalue string
	)
	if len(keyValue) == 0 {
		return keyValStrs, fmt.Errorf("key value is nil")
	}
	// if key value is a map, convert it to key value slice
	if kvMap, ok := keyValue[0].(map[k]v); ok {
		for key, value := range kvMap {
			if strkey, err = ctx.toKeyStr(key); err != nil {
				return nil, err
			}
			if strvalue, err = ctx.MarshalValue(value); err != nil {
				return nil, err
			}
			keyValStrs = append(keyValStrs, strkey, strvalue)
		}
	} else if l := len(keyValue); l%2 == 0 {
		for i := 0; i < l; i += 2 {
			if strkey, err = ctx.toKeyStr(key); err != nil {
				return nil, err
			}
			if strvalue, err = ctx.MarshalValue(value); err != nil {
				return nil, err
			}
			keyValStrs = append(keyValStrs, strkey, strvalue)
		}
	} else {
		return nil, fmt.Errorf("invalid type key value while converting to strings")
	}
	return keyValStrs, nil
}
func (ctx *Ctx[k, v]) MsgpackUnmarshalValue(msgpack []byte) (rets interface{}, err error) {
	return nil, nil
}

func (ctx *Ctx[k, v]) MsgpackUnmarshalKeyValues(msgpack []byte) (rets interface{}, err error) {
	return nil, nil
}
