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

	// newkey = hashInterface.CloneToRedisKey(req.Key, req.RedisDataSource)
	// if newkey.ValidDataKey() != nil {
	// 	return nil, fmt.Errorf("key name is invalid: %s", req.Key)
	// }
	return nil, nil
}

func (req *DoptimeReqCtx) ToValue(key redisdb.IHttpKey, msgpack []byte) (value interface{}, err error) {
	value, err = key.DeserializeValue(msgpack)
	mapper.Decode(req.Params, &value)
	return value, err
}
func (req *DoptimeReqCtx) ToValues(key redisdb.IHttpKey, msgpack []string) (value []interface{}, err error) {
	value, err = key.DeserializeValues(msgpack)
	for i, v := range value {
		//这个地方有错误，fields 的值会不断变化
		req.Params["@field"] = req.Fields[i]
		err = mapper.Decode(req.Params, &v)
		if err != nil {
			return value, err
		}
	}

	return value, err
}
