package httpserve

import (
	"fmt"

	"github.com/doptime/doptime/utils/mapper"
	"github.com/doptime/redisdb"
)

func (req *DoptimeReqCtx) CtxWithValueSchemaChecked(keyType redisdb.KeyType, msgpackData []byte) (newkey *redisdb.RedisKey[string, interface{}], value interface{}, err error) {

	keyScope := redisdb.KeyScope(req.Key)
	hashInterface, exists := redisdb.RediskeyForWeb.Get(keyScope + ":" + req.RedisDataSource)
	if hashInterface == nil && !exists {
		return nil, nil, fmt.Errorf("key schema is unconfigured: %s", keyScope)
	}
	if hashInterface.ValidDataKey() != nil {
		return nil, nil, hashInterface.ValidDataKey()
	}
	if hashInterface.GetKeyType() != keyType {
		return nil, nil, fmt.Errorf("key type is mismatched: %s != %s", string(hashInterface.GetKeyType()), keyType)
	}

	if len(msgpackData) > 0 {
		value, err = hashInterface.DeserializeToInterface(msgpackData)
		if err != nil {
			return nil, nil, err
		}
	}

	mapper.Decode(map[string]interface{}(req.JwtClaims), &value)

	newkey = hashInterface.CloneToRedisKey(req.Key, req.RedisDataSource)
	if newkey.ValidDataKey() != nil {
		return nil, nil, fmt.Errorf("key name is invalid: %s", req.Key)
	}
	return newkey, value, nil
}

func (req *DoptimeReqCtx) HashCtxWitchValueSchemaChecked(msgpackData []byte) (db *redisdb.HashKey[string, interface{}], value interface{}, err error) {
	var ctx *redisdb.RedisKey[string, interface{}]
	ctx, value, err = req.CtxWithValueSchemaChecked(redisdb.KeyTypeHash, msgpackData)
	if err != nil {
		return nil, nil, err
	}
	return &redisdb.HashKey[string, interface{}]{RedisKey: *ctx}, value, nil
}
func (req *DoptimeReqCtx) StringCtxWitchValueSchemaChecked(msgpackData []byte) (db *redisdb.StringKey[string, interface{}], value interface{}, err error) {
	var ctx *redisdb.RedisKey[string, interface{}]
	ctx, value, err = req.CtxWithValueSchemaChecked(redisdb.KeyTypeString, msgpackData)
	if err != nil {
		return nil, nil, err
	}
	return &redisdb.StringKey[string, interface{}]{RedisKey: *ctx}, value, nil
}
func (req *DoptimeReqCtx) ListCtxWitchValueSchemaChecked(msgpackData []byte) (db *redisdb.ListKey[interface{}], value interface{}, err error) {
	var ctx *redisdb.RedisKey[string, interface{}]
	ctx, value, err = req.CtxWithValueSchemaChecked(redisdb.KeyTypeList, msgpackData)
	if err != nil {
		return nil, nil, err
	}
	return &redisdb.ListKey[interface{}]{RedisKey: *ctx}, value, nil
}
func (req *DoptimeReqCtx) ZSetCtxWitchValueSchemaChecked(msgpackData []byte) (db *redisdb.ZSetKey[string, interface{}], value interface{}, err error) {
	var ctx *redisdb.RedisKey[string, interface{}]
	ctx, value, err = req.CtxWithValueSchemaChecked(redisdb.KeyTypeZSet, msgpackData)
	if err != nil {
		return nil, nil, err
	}
	return &redisdb.ZSetKey[string, interface{}]{RedisKey: *ctx}, value, nil
}

func (req *DoptimeReqCtx) SetCtxWitchValueSchemaChecked(msgpackData []byte) (db *redisdb.SetKey[string, interface{}], value interface{}, err error) {
	var ctx *redisdb.RedisKey[string, interface{}]
	ctx, value, err = req.CtxWithValueSchemaChecked(redisdb.KeyTypeSet, msgpackData)
	if err != nil {
		return nil, nil, err
	}
	return &redisdb.SetKey[string, interface{}]{RedisKey: *ctx}, value, nil
}
