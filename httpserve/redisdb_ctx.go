package httpserve

import (
	"fmt"
	"strings"

	"github.com/doptime/doptime/utils/mapper"
	"github.com/doptime/redisdb"
)

func (req *DoptimeReqCtx) CtxWithValueSchemaChecked(keyType redisdb.KeyType) (newkey *redisdb.RedisKey[string, interface{}], err error) {

	keyScope := strings.ToLower(redisdb.KeyScope(req.Key))
	hashInterface, exists := redisdb.RediskeyInterfaceForWebVisit.Get(keyScope + ":" + req.RedisDataSource)
	if hashInterface == nil && !exists {
		return nil, fmt.Errorf("key schema is unconfigured: %s", keyScope)
	}
	if hashInterface.ValidDataKey() != nil {
		return nil, hashInterface.ValidDataKey()
	}
	if hashInterface.GetKeyType() != keyType {
		return nil, fmt.Errorf("key type is mismatched: %s != %s", string(hashInterface.GetKeyType()), keyType)
	}

	newkey = hashInterface.CloneToRedisKey(req.Key, req.RedisDataSource)
	if newkey.ValidDataKey() != nil {
		return nil, fmt.Errorf("key name is invalid: %s", req.Key)
	}
	return newkey, nil
}

func (req *DoptimeReqCtx) ToValue(key *redisdb.RedisKey[string, interface{}], msgpack []byte) (value interface{}, err error) {
	value, err = key.DeserializeToInterface(msgpack)
	mapper.Decode(map[string]interface{}(req.Params), &value)
	return value, err
}
func (req *DoptimeReqCtx) ToValues(key *redisdb.RedisKey[string, interface{}], msgpack []string) (value []interface{}, err error) {
	value, err = key.DeserializeToInterfaceSlice(msgpack)
	for i, v := range value {
		//这个地方有错误，fields 的值会不断变化
		req.Params["@field"] = req.Fields[i]
		err = mapper.Decode(map[string]interface{}(req.Params), &v)
		if err != nil {
			return value, err
		}
	}

	return value, err
}

func (req *DoptimeReqCtx) HashCtxFromSchema() (db *redisdb.HashKey[string, interface{}], err error) {
	var ctx *redisdb.RedisKey[string, interface{}]
	ctx, err = req.CtxWithValueSchemaChecked(redisdb.KeyTypeHash)
	if err != nil {
		return nil, err
	}
	return &redisdb.HashKey[string, interface{}]{RedisKey: *ctx}, nil
}
func (req *DoptimeReqCtx) StringCtxFromSchema() (db *redisdb.StringKey[string, interface{}], err error) {
	var ctx *redisdb.RedisKey[string, interface{}]
	ctx, err = req.CtxWithValueSchemaChecked(redisdb.KeyTypeString)
	if err != nil {
		return nil, err
	}
	return &redisdb.StringKey[string, interface{}]{RedisKey: *ctx}, nil
}
func (req *DoptimeReqCtx) ListCtxFromSchema() (db *redisdb.ListKey[interface{}], err error) {
	var ctx *redisdb.RedisKey[string, interface{}]
	ctx, err = req.CtxWithValueSchemaChecked(redisdb.KeyTypeList)
	if err != nil {
		return nil, err
	}
	return &redisdb.ListKey[interface{}]{RedisKey: *ctx}, nil
}
func (req *DoptimeReqCtx) ZSetCtxFromSchema() (db *redisdb.ZSetKey[string, interface{}], err error) {
	var ctx *redisdb.RedisKey[string, interface{}]
	ctx, err = req.CtxWithValueSchemaChecked(redisdb.KeyTypeZSet)
	if err != nil {
		return nil, err
	}
	return &redisdb.ZSetKey[string, interface{}]{RedisKey: *ctx}, nil
}

func (req *DoptimeReqCtx) SetCtxFromSchema() (db *redisdb.SetKey[string, interface{}], err error) {
	var ctx *redisdb.RedisKey[string, interface{}]
	ctx, err = req.CtxWithValueSchemaChecked(redisdb.KeyTypeSet)
	if err != nil {
		return nil, err
	}
	return &redisdb.SetKey[string, interface{}]{RedisKey: *ctx}, nil
}
