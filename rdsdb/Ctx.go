package rdsdb

import (
	"context"
	"fmt"
	"reflect"
	"time"
	"unicode"

	"github.com/doptime/doptime/config"
	"github.com/doptime/doptime/dlog"
	"github.com/doptime/doptime/specification"
	"github.com/redis/go-redis/v9"
)

type Ctx[k comparable, v any] struct {
	Context context.Context
	Rds     *redis.Client
	Key     string
}

func NonKey[k comparable, v any](ops ...*DataOption) *Ctx[k, v] {
	ctx := &Ctx[k, v]{}
	if err := ctx.LoadDataOption(ops...); err != nil {
		dlog.Error().Err(err).Msg("data.New failed")
		return nil
	}
	return ctx
}
func (ctx *Ctx[k, v]) setKeyTypeIdentifier() {
	var ValueIdentifier string
	typeOfV := reflect.TypeOf((*v)(nil)).Elem()
	switch typeOfV.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		ValueIdentifier = "i"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		ValueIdentifier = "u"
	case reflect.String:
		ValueIdentifier = "s"
	case reflect.Float32, reflect.Float64:
		ValueIdentifier = "f"
	case reflect.Bool:
		ValueIdentifier = "b"
	default:
		ValueIdentifier = ""
	}
	//convert to upper case with the first character
	if ValueIdentifier != "" {
		ctx.Key = ValueIdentifier + ctx.Key
		if len(ctx.Key) > 1 && unicode.IsLower(rune(ctx.Key[1])) {
			runes := []rune(ctx.Key)
			runes[1] = unicode.ToUpper(runes[1])
			ctx.Key = string(runes)
		}
	}
}
func (ctx *Ctx[k, v]) getKeyTypeIdentifier() (t rune) {
	if len(ctx.Key) < 2 {
		return 0
	}
	t = rune(ctx.Key[0])
	switch t {
	case 'i', 'u', 's', 'f', 'b':
		return t
	}
	return 0
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

func (ctx *Ctx[k, v]) LoadDataOption(ops ...*DataOption) error {
	var rdsName string
	if len(ops) > 0 {
		ctx.Key = ops[0].Key
		rdsName = ops[0].DataSource
	}
	if len(ctx.Key) == 0 && !specification.GetValidDataKeyName((*v)(nil), &ctx.Key) {
		return fmt.Errorf("invalid keyname infered from type: " + reflect.TypeOf((*v)(nil)).String())
	}
	var exists bool
	if ctx.Rds, exists = config.Rds[rdsName]; !exists {
		return fmt.Errorf("Rds item unconfigured: " + rdsName)
	}
	ctx.Context = context.Background()

	return nil
}
