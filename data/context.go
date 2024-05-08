package data

import (
	"context"
	"time"

	"github.com/bits-and-blooms/bloom/v3"
	"github.com/doptime/doptime/config"
	"github.com/doptime/doptime/dlog"
	"github.com/doptime/doptime/specification"
	"github.com/redis/go-redis/v9"
)

type Ctx[k comparable, v any] struct {
	Ctx             context.Context
	Rds             *redis.Client
	Key             string
	BloomFilterKeys *bloom.BloomFilter
}

func New[k comparable, v any](ops ...*DataOption) *Ctx[k, v] {
	var (
		rds    *redis.Client
		option *DataOption = &DataOption{}
		exists bool
	)
	if len(ops) > 0 {
		option = ops[0]
	}
	//panic if Key is empty
	if !specification.GetValidDataKeyName((*v)(nil), &option.Key) {
		dlog.Panic().Str("Key is empty in Data.New", option.Key).Send()
	}
	if rds, exists = config.Rds[option.DataSource]; !exists {
		dlog.Error().Str("DataSource not defined in enviroment while calling Data.New", option.DataSource).Send()
		return nil
	}
	ctx := &Ctx[k, v]{Ctx: context.Background(), Rds: rds, Key: option.Key}
	dlog.Debug().Str("data New create end!", option.Key).Send()
	return ctx
}
func (ctx *Ctx[k, v]) Time() (tm time.Time, err error) {
	cmd := ctx.Rds.Time(ctx.Ctx)
	return cmd.Result()
}
