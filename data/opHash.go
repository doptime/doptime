package data

import (
	"encoding/json"
	"reflect"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

func (ctx *Ctx[k, v]) HGet(field k) (value v, err error) {

	var (
		cmd      *redis.StringCmd
		valBytes []byte
		fieldStr string
	)
	if fieldStr, err = ctx.toKeyStr(field); err != nil {
		return value, err
	}

	if cmd = ctx.Rds.HGet(ctx.Ctx, ctx.Key, fieldStr); cmd.Err() != nil {
		return value, cmd.Err()
	}
	if valBytes, err = cmd.Bytes(); err != nil {
		return value, err
	}
	return ctx.toValue(valBytes)
}

// HSet accepts values in following formats:
//
//   - HSet("myhash", "key1", "value1", "key2", "value2")
//
//   - HSet("myhash", map[string]interface{}{"key1": "value1", "key2": "value2"})
func (ctx *Ctx[k, v]) HSet(values ...interface{}) (err error) {
	var (
		KeyValuesStrs []string
	)
	if KeyValuesStrs, err = ctx.toKeyValueStrs(values...); err != nil {
		return err
	}
	status := ctx.Rds.HSet(ctx.Ctx, ctx.Key, KeyValuesStrs)
	return status.Err()
}

func (ctx *Ctx[k, v]) HExists(field k) (ok bool, err error) {

	var (
		cmd      *redis.BoolCmd
		fieldStr string
	)
	if fieldStr, err = ctx.toKeyStr(field); err != nil {
		return false, err
	}
	cmd = ctx.Rds.HExists(ctx.Ctx, ctx.Key, fieldStr)
	return cmd.Result()

}
func (ctx *Ctx[k, v]) HGetAll() (mapOut map[k]v, err error) {
	var (
		cmd *redis.MapStringStringCmd
		key k
		val v
	)
	mapOut = make(map[k]v)
	if cmd = ctx.Rds.HGetAll(ctx.Ctx, ctx.Key); cmd.Err() != nil {
		return mapOut, cmd.Err()
	}
	//append all data to mapOut
	for k, v := range cmd.Val() {
		if key, err = ctx.toKey([]byte(k)); err != nil {
			log.Info().AnErr("HGetAll: key unmarshal error:", err).Msgf("Key: %s", ctx.Key)
			continue
		}
		if val, err = ctx.toValue([]byte(v)); err != nil {
			log.Info().AnErr("HGetAll: value unmarshal error:", err).Msgf("Key: %s", ctx.Key)
			continue
		}
		mapOut[key] = val
	}
	return mapOut, err
}
func (ctx *Ctx[k, v]) HRandField(count int) (fields []k, err error) {
	var (
		cmd *redis.StringSliceCmd
	)
	if cmd = ctx.Rds.HRandField(ctx.Ctx, ctx.Key, count); cmd.Err() != nil {
		return nil, cmd.Err()
	}
	return ctx.toKeys(cmd.Val())
}

func (ctx *Ctx[k, v]) HMGET(fields ...k) (values []v, err error) {
	var (
		cmd          *redis.SliceCmd
		fieldsString []string
		rawValues    []string
	)
	if fieldsString, err = ctx.toKeyStrs(fields...); err != nil {
		return nil, err
	}
	if cmd = ctx.Rds.HMGet(ctx.Ctx, ctx.Key, fieldsString...); cmd.Err() != nil {
		return nil, cmd.Err()
	}
	rawValues = make([]string, len(cmd.Val()))
	for i, val := range cmd.Val() {
		if val == nil {
			continue
		}
		rawValues[i] = val.(string)
	}
	return ctx.toValues(rawValues...)
}

func (ctx *Ctx[k, v]) HLen() (length int64, err error) {
	cmd := ctx.Rds.HLen(ctx.Ctx, ctx.Key)
	return cmd.Val(), cmd.Err()
}
func (ctx *Ctx[k, v]) HDel(fields ...k) (err error) {
	var (
		cmd       *redis.IntCmd
		fieldStrs []string
		bytes     []byte
	)
	if len(fields) == 0 {
		return nil
	}
	//if k is  string, then use HDEL directly
	if reflect.TypeOf(fields[0]).Kind() == reflect.String {
		fieldStrs = interface{}(fields).([]string)
	} else {
		//if k is not string, then marshal k to string
		fieldStrs = make([]string, len(fields))
		for i, field := range fields {
			if bytes, err = json.Marshal(field); err != nil {
				return err
			}
			fieldStrs[i] = string(bytes)
		}
	}
	cmd = ctx.Rds.HDel(ctx.Ctx, ctx.Key, fieldStrs...)
	return cmd.Err()
}
func (ctx *Ctx[k, v]) HKeys() (fields []k, err error) {
	var (
		cmd *redis.StringSliceCmd
	)
	if cmd = ctx.Rds.HKeys(ctx.Ctx, ctx.Key); cmd.Err() != nil {
		return nil, cmd.Err()
	}
	return ctx.toKeys(cmd.Val())
}
func (ctx *Ctx[k, v]) HVals() (values []v, err error) {
	var cmd *redis.StringSliceCmd
	if cmd = ctx.Rds.HVals(ctx.Ctx, ctx.Key); cmd.Err() != nil {
		return nil, cmd.Err()
	}
	return ctx.toValues(cmd.Val()...)
}
func (ctx *Ctx[k, v]) HIncrBy(field k, increment int64) (err error) {
	var (
		cmd      *redis.IntCmd
		fieldStr string
	)
	if fieldStr, err = ctx.toKeyStr(field); err != nil {
		return err
	}
	cmd = ctx.Rds.HIncrBy(ctx.Ctx, ctx.Key, fieldStr, increment)
	return cmd.Err()
}

func (ctx *Ctx[k, v]) HIncrByFloat(field k, increment float64) (err error) {
	var (
		cmd      *redis.FloatCmd
		fieldStr string
	)
	if fieldStr, err = ctx.toKeyStr(field); err != nil {
		return err
	}
	cmd = ctx.Rds.HIncrByFloat(ctx.Ctx, ctx.Key, fieldStr, increment)
	return cmd.Err()

}
func (ctx *Ctx[k, v]) HSetNX(field k, value v) (err error) {
	var (
		cmd      *redis.BoolCmd
		fieldStr string
		valStr   string
	)
	if fieldStr, err = ctx.toKeyStr(field); err != nil {
		return err
	}
	if valStr, err = ctx.toValueStr(value); err != nil {
		return err
	}
	cmd = ctx.Rds.HSetNX(ctx.Ctx, ctx.Key, fieldStr, valStr)
	return cmd.Err()
}
