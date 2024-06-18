package rdsdb

import (
	"fmt"
	"math/rand"

	"github.com/bits-and-blooms/bloom/v3"
)

func BuildBloomFilterByKeys(keys []string, capacity int, falsePosition float64) (BloomFilterKeys *bloom.BloomFilter) {
	if capacity < 0 {
		capacity = int(float64(len(keys))*1.2) + 1024*1024 + int(rand.Uint32()%10000)
	}
	if falsePosition <= 0 || falsePosition >= 1 {
		falsePosition = 0.0000001 + rand.Float64()/10000000
	}
	BloomFilterKeys = bloom.NewWithEstimates(uint(capacity), falsePosition)
	//if type of k is string, then AddString is faster than Add
	for _, it := range keys {
		BloomFilterKeys.AddString(it)
	}
	return BloomFilterKeys
}

func (ctx *CtxString[k, v]) BuildBloomFilter(keys []string, capacity int, falsePosition float64) (err error) {
	ctx.BloomFilterKeys = BuildBloomFilterByKeys(keys, capacity, falsePosition)
	return nil
}
func (ctx *CtxString[k, v]) TestKey(key k) (exist bool, err error) {
	var (
		keyStr string
	)
	if ctx.BloomFilterKeys == nil {
		return false, fmt.Errorf("BloomKeys is nil, please BuildBloomFilter first")
	}
	if keyStr, err = ctx.toKeyStr(key); err != nil {
		return false, fmt.Errorf("TestKey -> toKeyStr error: %v", err.Error())
	}
	return ctx.BloomFilterKeys.TestString(keyStr), nil
}

func (ctx *CtxHash[k, v]) BuildBloomFilter(capacity int, falsePosition float64) (err error) {
	//get type of key, if not hash, then return error
	var keys []string
	if keys, err = ctx.Rds.HKeys(ctx.Context, ctx.Key).Result(); err != nil {
		return err
	}
	ctx.BloomFilterKeys = BuildBloomFilterByKeys(keys, capacity, falsePosition)
	return nil
}
func (ctx *CtxHash[k, v]) TestKey(key k) (exist bool, err error) {
	var (
		keyStr string
	)
	if ctx.BloomFilterKeys == nil {
		return false, fmt.Errorf("BloomKeys is nil, please BuildBloomFilter first")
	}
	if keyStr, err = ctx.toKeyStr(key); err != nil {
		return false, fmt.Errorf("TestKey -> toKeyStr error: %v", err.Error())
	}
	return ctx.BloomFilterKeys.TestString(keyStr), nil
}
