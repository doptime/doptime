package rdsdb

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/doptime/doptime/config"
	"github.com/doptime/doptime/specification"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/vmihailenco/msgpack/v5"
)

type CtxInterface interface {
	// MsgpackUnmarshalValue(msgpack []byte) (rets interface{}, err error)
	// MsgpackUnmarshalKeyValues(msgpack []byte) (rets interface{}, err error)
	CheckDataSchema(msgpackBytes []byte) (val interface{}, err error)
	Validate() error
}

var hKeyMap cmap.ConcurrentMap[string, CtxInterface] = cmap.New[CtxInterface]()
var nonKey = NonKey[string, interface{}]()

func CtxWitchValueSchemaChecked(key, keyType string, RedisDataSource string, msgpackData []byte) (db *Ctx[string, interface{}], value interface{}, err error) {

	hashInterface, exists := hKeyMap.Get(key + ":" + RedisDataSource)
	if exists && msgpackData != nil {
		value, err = hashInterface.CheckDataSchema(msgpackData)
		if err != nil {
			return nil, nil, err
		}
	}
	if disallowed, found := specification.DisAllowedDataKeyNames[key]; found && disallowed {
		return nil, nil, fmt.Errorf("key name is disallowed: " + key)
	}
	_ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	ctx := Ctx[string, interface{}]{_ctx, RedisDataSource, nil, key, keyType, nonKey.MarshalValue, nonKey.UnmarshalValue, nonKey.UnmarshalValues}
	if ctx.Rds, exists = config.Rds[RedisDataSource]; !exists {
		return nil, nil, fmt.Errorf("rds item unconfigured: " + RedisDataSource)
	}

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
	if _, ok := config.Rds[ctx.RdsName]; !ok {
		return fmt.Errorf("rds item unconfigured: " + ctx.RdsName)
	}
	return nil
}

func (ctx *Ctx[k, v]) CheckDataSchema(msgpackBytes []byte) (val interface{}, err error) {
	if len(msgpackBytes) == 0 {
		return nil, fmt.Errorf("msgpackBytes is empty")
	}

	var vInstance v
	vType := reflect.TypeOf(vInstance)

	if vType.Kind() == reflect.Ptr {
		vType = vType.Elem()
	}

	if vType.Kind() == reflect.Struct {
		elem := reflect.New(vType).Interface()
		if err := msgpack.Unmarshal(msgpackBytes, elem); err != nil {
			return nil, err
		}
		return elem, nil
	}

	return nil, fmt.Errorf("type %v is not a struct or pointer to struct", vType)
}
