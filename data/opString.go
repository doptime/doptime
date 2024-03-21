package data

import (
	"github.com/rs/zerolog/log"

	"github.com/vmihailenco/msgpack/v5"
)

// get all keys that match the pattern, and return a map of key->value
func (ctx *Ctx[k, v]) GetAll(match string) (mapOut map[k]v, err error) {
	var (
		keys []string = []string{match}
		val  []byte
	)
	if keys, err = ctx.Scan(match, 0, 1024*1024*1024); err != nil {
		return nil, err
	}
	mapOut = make(map[k]v, len(keys))
	var result error
	for _, key := range keys {
		if val, result = ctx.Rds.Get(ctx.Ctx, key).Bytes(); result != nil {
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
			log.Info().AnErr("GetAll: key unmarshal error:", err).Msgf("Key: %s", ctx.Key)
			continue
		}
		v, err := ctx.toValue(val)
		if err != nil {
			log.Info().AnErr("GetAll: value unmarshal error:", err).Msgf("Key: %s", ctx.Key)
			continue
		}
		mapOut[k] = v
	}
	return mapOut, err
}

// set each key value of _map to redis string type key value
func (ctx *Ctx[k, v]) SetAll(_map map[k]v) (err error) {
	//HSet each element of _map to redis
	var (
		result error
		bytes  []byte
		keyStr string
	)
	pipe := ctx.Rds.Pipeline()
	for k, v := range _map {
		if bytes, err = msgpack.Marshal(v); err != nil {
			return err
		}
		if keyStr, err = ctx.toKeyStr(k); err != nil {
			return err
		}

		pipe.Set(ctx.Ctx, ctx.Key+":"+keyStr, bytes, -1)
	}
	pipe.Exec(ctx.Ctx)
	return result
}
