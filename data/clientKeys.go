package data

import (
	"github.com/redis/go-redis/v9"
)

// sacn key by pattern
func (ctx *Ctx[k, v]) Scan(cursorOld uint64, match string, count int64) (keys []string, cursorNew uint64, err error) {
	var (
		cmd   *redis.ScanCmd
		_keys []string
	)
	//scan all keys
	for {

		if cmd = ctx.Rds.Scan(ctx.Ctx, cursorOld, match, count); cmd.Err() != nil {
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
