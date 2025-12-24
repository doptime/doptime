package httpserve

import (
	"fmt"

	"github.com/doptime/redisdb"
)

func CtxWithValueSchemaChecked(key, keyType string, RedisDataSource string, msgpackData []byte) (newkey *redisdb.RedisKey[string, interface{}], value interface{}, err error) {

	if !redisdb.IsValidKeyType(keyType) {
		return nil, nil, fmt.Errorf("%s", "key type is invalid: "+keyType)
	}
	keyScope := redisdb.KeyScope(key)
	hashInterface, exists := redisdb.RediskeyForWeb.Get(keyScope + ":" + RedisDataSource)
	if hashInterface == nil && !exists {
		return nil, nil, fmt.Errorf("key schema is unconfigured: %s", keyScope)
	}
	if hashInterface.ValidDataKey() != nil {
		return nil, nil, hashInterface.ValidDataKey()
	}

	if len(msgpackData) > 0 {
		value, err = hashInterface.DeserializeToInterface(msgpackData)
		if err != nil {
			return nil, nil, err
		} else if exists {
			hashInterface.TimestampFiller(value)
		}
	}

	newkey = hashInterface.CloneToRedisKey(key, RedisDataSource)
	if newkey.ValidDataKey() != nil {
		return nil, nil, fmt.Errorf("key name is invalid: %s", key)
	}
	return newkey, value, nil
}

func HashCtxWitchValueSchemaChecked(key string, RedisDataSource string, msgpackData []byte) (db *redisdb.HashKey[string, interface{}], value interface{}, err error) {
	var ctx *redisdb.RedisKey[string, interface{}]
	ctx, value, err = CtxWithValueSchemaChecked(key, "hash", RedisDataSource, msgpackData)
	if err != nil {
		return nil, nil, err
	}
	return &redisdb.HashKey[string, interface{}]{RedisKey: *ctx}, value, nil
}
func StringCtxWitchValueSchemaChecked(key string, RedisDataSource string, msgpackData []byte) (db *redisdb.StringKey[string, interface{}], value interface{}, err error) {
	var ctx *redisdb.RedisKey[string, interface{}]
	ctx, value, err = CtxWithValueSchemaChecked(key, "string", RedisDataSource, msgpackData)
	if err != nil {
		return nil, nil, err
	}
	return &redisdb.StringKey[string, interface{}]{RedisKey: *ctx}, value, nil
}
func ListCtxWitchValueSchemaChecked(key string, RedisDataSource string, msgpackData []byte) (db *redisdb.ListKey[interface{}], value interface{}, err error) {
	var ctx *redisdb.RedisKey[string, interface{}]
	ctx, value, err = CtxWithValueSchemaChecked(key, "list", RedisDataSource, msgpackData)
	if err != nil {
		return nil, nil, err
	}
	return &redisdb.ListKey[interface{}]{RedisKey: *ctx}, value, nil
}
func ZSetCtxWitchValueSchemaChecked(key string, RedisDataSource string, msgpackData []byte) (db *redisdb.ZSetKey[string, interface{}], value interface{}, err error) {
	var ctx *redisdb.RedisKey[string, interface{}]
	ctx, value, err = CtxWithValueSchemaChecked(key, "zset", RedisDataSource, msgpackData)
	if err != nil {
		return nil, nil, err
	}
	return &redisdb.ZSetKey[string, interface{}]{RedisKey: *ctx}, value, nil
}

func SetCtxWitchValueSchemaChecked(key string, RedisDataSource string, msgpackData []byte) (db *redisdb.SetKey[string, interface{}], value interface{}, err error) {
	var ctx *redisdb.RedisKey[string, interface{}]
	ctx, value, err = CtxWithValueSchemaChecked(key, "set", RedisDataSource, msgpackData)
	if err != nil {
		return nil, nil, err
	}
	return &redisdb.SetKey[string, interface{}]{RedisKey: *ctx}, value, nil
}
