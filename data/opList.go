package data

import (
	"time"

	"github.com/redis/go-redis/v9"
)

func (ctx *Ctx[k, v]) RPush(param ...v) (err error) {
	if val, err := ctx.toValueStrs(param); err != nil {
		return err
	} else {

		return ctx.Rds.RPush(ctx.Ctx, ctx.Key, val).Err()
	}
}
func (ctx *Ctx[k, v]) LPush(param ...v) (err error) {
	if val, err := ctx.toValueStrs(param); err != nil {
		return err
	} else {
		return ctx.Rds.LPush(ctx.Ctx, ctx.Key, val).Err()
	}
}
func (ctx *Ctx[k, v]) RPop() (ret v, err error) {
	cmd := ctx.Rds.RPop(ctx.Ctx, ctx.Key)
	if data, err := cmd.Bytes(); err != nil {
		return ret, err
	} else {
		return ctx.toValue(data)
	}
}
func (ctx *Ctx[k, v]) LPop() (ret v, err error) {
	cmd := ctx.Rds.LPop(ctx.Ctx, ctx.Key)
	if data, err := cmd.Bytes(); err != nil {
		return ret, err
	} else {
		return ctx.toValue(data)
	}
}
func (ctx *Ctx[k, v]) LRange(start, stop int64) (ret []v, err error) {
	cmd := ctx.Rds.LRange(ctx.Ctx, ctx.Key, start, stop)
	if data, err := cmd.Result(); err != nil {
		return ret, err
	} else {
		for _, val := range data {
			if v, err := ctx.toValue([]byte(val)); err != nil {
				return ret, err
			} else {
				ret = append(ret, v)
			}
		}
		return ret, nil
	}
}
func (ctx *Ctx[k, v]) LRem(count int64, param v) (err error) {
	if val, err := ctx.toValueStr(param); err != nil {
		return err
	} else {
		return ctx.Rds.LRem(ctx.Ctx, ctx.Key, count, val).Err()
	}
}
func (ctx *Ctx[k, v]) LSet(index int64, param v) (err error) {
	if val, err := ctx.toValueStr(param); err != nil {
		return err
	} else {
		return ctx.Rds.LSet(ctx.Ctx, ctx.Key, index, val).Err()
	}
}
func (ctx *Ctx[k, v]) BLPop(timeout time.Duration) (ret v, err error) {
	cmd := ctx.Rds.BLPop(ctx.Ctx, timeout, ctx.Key)
	if data, err := cmd.Result(); err != nil {
		return ret, err
	} else {
		return ctx.toValue([]byte(data[1]))
	}
}
func (ctx *Ctx[k, v]) BRPop(timeout time.Duration) (ret v, err error) {
	cmd := ctx.Rds.BRPop(ctx.Ctx, timeout, ctx.Key)
	if data, err := cmd.Result(); err != nil {
		return ret, err
	} else {
		return ctx.toValue([]byte(data[1]))
	}
}
func (ctx *Ctx[k, v]) BRPopLPush(destination string, timeout time.Duration) (ret v, err error) {
	cmd := ctx.Rds.BRPopLPush(ctx.Ctx, ctx.Key, destination, timeout)
	if data, err := cmd.Bytes(); err != nil {
		return ret, err
	} else {
		return ctx.toValue(data)
	}
}
func (ctx *Ctx[k, v]) LInsertBefore(pivot, param v) (err error) {
	var (
		pivotStr string
	)
	if val, err := ctx.toValueStr(param); err != nil {
		return err
	} else {
		if pivotStr, err = ctx.toValueStr(pivot); err != nil {
			return err
		} else {
			return ctx.Rds.LInsertBefore(ctx.Ctx, ctx.Key, pivotStr, val).Err()
		}
	}
}
func (ctx *Ctx[k, v]) LInsertAfter(pivot, param v) (err error) {
	var (
		pivotStr string
	)
	if val, err := ctx.toValueStr(param); err != nil {
		return err
	} else {
		if pivotStr, err = ctx.toValueStr(pivot); err != nil {
			return err
		} else {
			return ctx.Rds.LInsertAfter(ctx.Ctx, ctx.Key, pivotStr, val).Err()
		}
	}
}
func (ctx *Ctx[k, v]) Sort(sort *redis.Sort) (ret []v, err error) {
	cmd := ctx.Rds.Sort(ctx.Ctx, ctx.Key, sort)
	if data, err := cmd.Result(); err != nil {
		return ret, err
	} else {
		for _, val := range data {
			if v, err := ctx.toValue([]byte(val)); err != nil {
				return ret, err
			} else {
				ret = append(ret, v)
			}
		}
		return ret, nil
	}
}
func (ctx *Ctx[k, v]) LTrim(start, stop int64) (err error) {
	return ctx.Rds.LTrim(ctx.Ctx, ctx.Key, start, stop).Err()
}
func (ctx *Ctx[k, v]) LIndex(index int64) (ret v, err error) {
	cmd := ctx.Rds.LIndex(ctx.Ctx, ctx.Key, index)
	if data, err := cmd.Bytes(); err != nil {
		return ret, err
	} else {
		return ctx.toValue(data)
	}
}
func (ctx *Ctx[k, v]) LLen() (length int64, err error) {
	cmd := ctx.Rds.LLen(ctx.Ctx, ctx.Key)
	return cmd.Result()
}
