package rdsdb

import (
	"context"
	"fmt"
	"reflect"

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

var hashKey = HashKey[string, interface{}](WithKey("_"))

// var zsetKey = ZSetKey[string, interface{}](WithKey("_"))
var listKey = ListKey[string, interface{}](WithKey("_"))

// var setKey = SetKey[string, interface{}](WithKey("_"))
var stringKey = StringKey[string, interface{}](WithKey("_"))

func HashCtxWitchValueSchemaChecked(key string, RedisDataSource string, msgpackData []byte) (db *CtxHash[string, interface{}], value interface{}, err error) {
	hashInterface, ok := hKeyMap.Get(key + ":" + RedisDataSource)
	if msgpackData != nil && ok {
		if err = hashInterface.Validate(); err != nil {
			return nil, nil, err
		} else if value, err = hashInterface.CheckDataSchema(msgpackData); err != nil {
			return nil, nil, err
		}
	}
	ctx := Ctx[string, interface{}]{hashKey.Context, RedisDataSource, hashKey.Rds, key, hashKey.KeyType, hashKey.Moder, hashKey.MarshalValue, hashKey.UnmarshalValue, hashKey.UnmarshalValues}
	if err = ctx.Validate(); err != nil {
		return nil, nil, err
	}
	return &CtxHash[string, interface{}]{ctx}, value, nil
}
func ListCtxWitchValueSchemaChecked(key string, RedisDataSource string, msgpackData []byte) (db *CtxList[string, interface{}], value interface{}, err error) {
	hashInterface, ok := hKeyMap.Get(key + ":" + RedisDataSource)
	if msgpackData != nil && ok {
		if err = hashInterface.Validate(); err != nil {
			return nil, nil, err
		} else if value, err = hashInterface.CheckDataSchema(msgpackData); err != nil {
			return nil, nil, err
		}
	}
	ctx := Ctx[string, interface{}]{listKey.Context, RedisDataSource, listKey.Rds, key, listKey.KeyType, listKey.Moder, listKey.MarshalValue, listKey.UnmarshalValue, listKey.UnmarshalValues}
	if err = ctx.Validate(); err != nil {
		return nil, nil, err
	}
	return &CtxList[string, interface{}]{ctx}, value, nil
}
func StringCtxWitchValueSchemaChecked(key string, RedisDataSource string, msgpackData []byte) (db *CtxString[string, interface{}], value interface{}, err error) {
	hashInterface, ok := hKeyMap.Get(key + ":" + RedisDataSource)
	if msgpackData != nil && ok {
		if err = hashInterface.Validate(); err != nil {
			return nil, nil, err
		} else if value, err = hashInterface.CheckDataSchema(msgpackData); err != nil {
			return nil, nil, err
		}
	}
	ctx := Ctx[string, interface{}]{stringKey.Context, RedisDataSource, stringKey.Rds, key, stringKey.KeyType, stringKey.Moder, stringKey.MarshalValue, stringKey.UnmarshalValue, stringKey.UnmarshalValues}
	if err = ctx.Validate(); err != nil {
		return nil, nil, err
	}
	return &CtxString[string, interface{}]{ctx}, value, nil
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
func (ctx *Ctx[k, v]) ModData(val interface{}) (err error) {
	_val, ok := val.(v)
	if !ok {
		return nil
	}
	if ctx.Moder != nil {
		return ctx.Moder.ApplyModifiers(context.Background(), &_val)
	}

	return nil
}
