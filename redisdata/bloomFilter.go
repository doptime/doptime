package redisdata

import (
	"log"
	"math/rand"

	"github.com/bits-and-blooms/bloom/v3"
)

func (ctx *Ctx[k, v]) BuildBloomFilterHKeys(capacity int, falsePosition float64) (err error) {
	//get type of key, if not hash, then return error
	var keys []string
	if keys, err = ctx.Rds.HKeys(ctx.Context, ctx.Key).Result(); err != nil {
		return err
	}
	return ctx.BuildBloomFilterByKeys(keys, capacity, falsePosition)
}
func (ctx *Ctx[k, v]) BuildBloomFilterByKeys(keys []string, capacity int, falsePosition float64) (err error) {
	if capacity < 0 {
		capacity = int(float64(len(keys))*1.2) + 1024*1024 + int(rand.Uint32()%10000)
	}
	if falsePosition <= 0 || falsePosition >= 1 {
		falsePosition = 0.0000001 + rand.Float64()/10000000
	}
	ctx.BloomFilterKeys = bloom.NewWithEstimates(uint(capacity), falsePosition)
	//if type of k is string, then AddString is faster than Add
	for _, it := range keys {
		ctx.BloomFilterKeys.AddString(it)
	}
	return nil
}
func (ctx *Ctx[k, v]) TestHKey(key k) (exist bool) {
	var (
		keyStr string
		err    error
	)
	if ctx.BloomFilterKeys == nil {
		log.Fatal("BloomKeys is nil, please BuildKeysBloomFilter first")
	}
	if keyStr, err = ctx.toKeyStr(key); err != nil {
		log.Fatalf("TestKey -> toKeyStr error: %v", err.Error())
	}
	return ctx.BloomFilterKeys.TestString(keyStr)
}
func (ctx *Ctx[k, v]) TestKey(key k) (exist bool) {
	var (
		keyStr string
		err    error
	)
	if ctx.BloomFilterKeys == nil {
		log.Fatal("BloomKeys is nil, please BuildKeysBloomFilter first")
	}
	if keyStr, err = ctx.toKeyStr(key); err != nil {
		log.Fatalf("TestKey -> toKeyStr error: %v", err.Error())
	}
	return ctx.BloomFilterKeys.TestString(ctx.Key + ":" + keyStr)
}
func (ctx *Ctx[k, v]) AddBloomKey(key k) (err error) {
	var (
		keyStr string
	)
	if ctx.BloomFilterKeys == nil {
		log.Fatal("BloomKeys is nil, please BuildKeysBloomFilter first")
	}
	if keyStr, err = ctx.toKeyStr(key); err != nil {
		return err
	}
	ctx.BloomFilterKeys.AddString(keyStr)
	return nil
}
