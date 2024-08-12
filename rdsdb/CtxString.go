package rdsdb

import (
	"strings"
	"time"

	"github.com/doptime/doptime/dlog"
)

type CtxString[k comparable, v any] struct {
	Ctx[k, v]
}

func StringKey[k comparable, v any](ops ...opSetter) *CtxString[k, v] {
	ctx := &CtxString[k, v]{}
	ctx.KeyType = "string"
	op := Option{}.buildOptions(ops...)
	if err := ctx.applyOption(op); err != nil {
		dlog.Error().Err(err).Msg("data.New failed")
		return nil
	}
	return ctx
}

func (ctx *CtxString[k, v]) ConcatKey(fields ...interface{}) *CtxString[k, v] {
	return &CtxString[k, v]{ctx.Duplicate(ConcatedKeys(ctx.Key, fields...), ctx.RdsName)}
}
func (ctx *CtxString[k, v]) Get(Field k) (value v, err error) {
	FieldStr, err := ctx.toKeyStr(Field)
	if err != nil {
		return value, err
	}
	var keyFields []string
	if len(ctx.Key) > 0 {
		keyFields = append(keyFields, ctx.Key)
	}
	if len(FieldStr) > 0 {
		keyFields = append(keyFields, FieldStr)
	}

	cmd := ctx.Rds.Get(ctx.Context, strings.Join(keyFields, ":"))
	if err := cmd.Err(); err != nil {
		return value, err
	}
	data, err := cmd.Bytes()
	if err != nil {
		return value, err
	}
	return ctx.UnmarshalValue(data)
}

func (ctx *CtxString[k, v]) Set(key k, value v, expiration time.Duration) error {
	keyStr, err := ctx.toKeyStr(key)
	if err != nil {
		return err
	}
	valStr, err := ctx.MarshalValue(value)
	if err != nil {
		return err
	}
	return ctx.Rds.Set(ctx.Context, ctx.Key+":"+keyStr, valStr, expiration).Err()
}

func (ctx *CtxString[k, v]) Del(key k) error {
	keyStr, err := ctx.toKeyStr(key)
	if err != nil {
		return err
	}
	return ctx.Rds.Del(ctx.Context, ctx.Key+":"+keyStr).Err()
}

// get all keys that match the pattern, and return a map of key->value
func (ctx *CtxString[k, v]) GetAll(match string) (mapOut map[k]v, err error) {
	var (
		keys []string = []string{match}
		val  []byte
	)
	if keys, _, err = ctx.Scan(0, match, 1024*1024*1024); err != nil {
		return nil, err
	}
	mapOut = make(map[k]v, len(keys))
	var result error
	for _, key := range keys {
		if val, result = ctx.Rds.Get(ctx.Context, key).Bytes(); result != nil {
			err = result
			continue
		}
		//use default prefix to avoid confict of hash key
		//k is start with ctx.Key, remove it
		if len(ctx.Key) > 0 && (len(key) >= len(ctx.Key)) && key[:len(ctx.Key)] == ctx.Key {
			key = key[len(ctx.Key)+1:]
		}

		k, err := ctx.toKey([]byte(key))
		if err != nil {
			dlog.Info().AnErr("GetAll: key unmarshal error:", err).Msgf("Key: %s", ctx.Key)
			continue
		}
		v, err := ctx.UnmarshalValue(val)
		if err != nil {
			dlog.Info().AnErr("GetAll: value unmarshal error:", err).Msgf("Key: %s", ctx.Key)
			continue
		}
		mapOut[k] = v
	}
	return mapOut, err
}

// set each key value of _map to redis string type key value
func (ctx *CtxString[k, v]) SetAll(_map map[k]v) (err error) {
	//HSet each element of _map to redis
	pipe := ctx.Rds.Pipeline()
	for k, v := range _map {
		keyStr, err := ctx.toKeyStr(k)
		if err != nil {
			return err
		}
		valStr, err := ctx.MarshalValue(v)
		if err != nil {
			return err
		}

		pipe.Set(ctx.Context, ctx.Key+":"+keyStr, valStr, -1)
	}
	_, err = pipe.Exec(ctx.Context)
	return err
}
