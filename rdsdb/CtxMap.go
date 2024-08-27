package rdsdb

import (
	"context"
	"fmt"
	"strings"

	"github.com/doptime/doptime/config"
	"github.com/doptime/doptime/specification"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/vmihailenco/msgpack/v5"
)

type CtxInterface interface {
	// MsgpackUnmarshalValue(msgpack []byte) (rets interface{}, err error)
	// MsgpackUnmarshalKeyValues(msgpack []byte) (rets interface{}, err error)
	CheckDataSchema(msgpackBytes []byte) (val interface{}, err error)
	ApplyModifiers(val interface{}) error
	Validate() error
}

var hKeyMap cmap.ConcurrentMap[string, CtxInterface] = cmap.New[CtxInterface]()
var nonKey = NonKey[string, interface{}]()

func CtxWitchValueSchemaChecked(key, keyType string, RedisDataSource string, msgpackData []byte) (db *Ctx[string, interface{}], value interface{}, err error) {
	keyItems := strings.Split(key, "@")
	for i, item := range keyItems {
		if len(item) > 1 && strings.Contains(item, ":") {
			keyItems[i] = strings.Split(item, ":")[0]
		}
	}
	key = strings.Join(keyItems, "@")

	hashInterface, exists := hKeyMap.Get(key + ":" + RedisDataSource)
	if hashInterface != nil && exists && msgpackData != nil {
		value, err = hashInterface.CheckDataSchema(msgpackData)
		if err != nil {
			return nil, nil, err
		}
	}
	if disallowed, found := specification.DisAllowedDataKeyNames[key]; found && disallowed {
		return nil, nil, fmt.Errorf("key name is disallowed: " + key)
	}
	ctx := Ctx[string, interface{}]{context.Background(), RedisDataSource, nil, key, keyType, nonKey.MarshalValue, nonKey.UnmarshalValue, nonKey.UnmarshalValues}
	if ctx.Rds, exists = config.Rds.Get(RedisDataSource); !exists {
		return nil, nil, fmt.Errorf("rds item unconfigured: " + RedisDataSource)
	}
	hKeyMap.Set(key+":"+RedisDataSource, &ctx)

	if value == nil && msgpackData != nil {
		value, err = ctx.CheckDataSchema(msgpackData)
		if err != nil {
			return nil, nil, err
		}
	}
	return &ctx, value, nil
}

func HashCtxWitchValueSchemaChecked(key string, RedisDataSource string, msgpackData []byte) (db *CtxHash[string, interface{}], value interface{}, err error) {
	var ctx *Ctx[string, interface{}]
	ctx, value, err = CtxWitchValueSchemaChecked(key, "hash", RedisDataSource, msgpackData)
	if err != nil {
		return nil, nil, err
	}
	return &CtxHash[string, interface{}]{*ctx}, value, nil
}
func ListCtxWitchValueSchemaChecked(key string, RedisDataSource string, msgpackData []byte) (db *CtxList[string, interface{}], value interface{}, err error) {
	var ctx *Ctx[string, interface{}]
	ctx, value, err = CtxWitchValueSchemaChecked(key, "list", RedisDataSource, msgpackData)
	if err != nil {
		return nil, nil, err
	}
	return &CtxList[string, interface{}]{*ctx}, value, nil
}
func StringCtxWitchValueSchemaChecked(key string, RedisDataSource string, msgpackData []byte) (db *CtxString[string, interface{}], value interface{}, err error) {
	var ctx *Ctx[string, interface{}]
	ctx, value, err = CtxWitchValueSchemaChecked(key, "string", RedisDataSource, msgpackData)
	if err != nil {
		return nil, nil, err
	}
	return &CtxString[string, interface{}]{*ctx}, value, nil
}

func (ctx *Ctx[k, v]) Validate() error {
	if disallowed, found := specification.DisAllowedDataKeyNames[ctx.Key]; found && disallowed {
		return fmt.Errorf("key name is disallowed: " + ctx.Key)
	}
	if _, ok := config.Rds.Get(ctx.RdsName); !ok {
		return fmt.Errorf("rds item unconfigured: " + ctx.RdsName)
	}
	return nil
}

func (ctx *Ctx[k, v]) CheckDataSchema(msgpackBytes []byte) (val interface{}, err error) {
	if len(msgpackBytes) == 0 {
		return nil, fmt.Errorf("msgpackBytes is empty")
	}

	var vInstance v

	if err = msgpack.Unmarshal(msgpackBytes, &vInstance); err != nil {
		return nil, err
	}

	return vInstance, nil
}
