package data

import (
	"encoding/json"
	"reflect"

	"github.com/doptime/doptime/vars"
	"github.com/vmihailenco/msgpack/v5"
)

func (ctx *Ctx[k, v]) toKeyStr(key k) (keyStr string, err error) {
	if vv := reflect.ValueOf(key); vv.Kind() == reflect.Ptr && vv.IsNil() {
		return "", nil
	} else if !vv.IsValid() {
		return keyStr, vars.ErrInvalidField
	}
	//if key is a string, directly append to keyBytes
	if strkey, ok := interface{}(key).(string); ok {
		return strkey, nil
	}
	if keyBytes, err := json.Marshal(key); err != nil {
		return keyStr, err
	} else {
		return string(keyBytes), nil
	}
}
func (ctx *Ctx[k, v]) toValueStr(value v) (valueStr string, err error) {
	//marshal with msgpack
	//nil value is allowed
	if bytes, err := msgpack.Marshal(value); err != nil {
		return valueStr, err
	} else {
		return string(bytes), nil
	}
}

func (ctx *Ctx[k, v]) toValueStrs(values []v) (valueStrs []string, err error) {
	var bytes []byte
	for _, value := range values {
		if bytes, err = msgpack.Marshal(value); err != nil {
			return nil, err
		}
		valueStrs = append(valueStrs, string(bytes))
	}
	return valueStrs, nil
}
func (ctx *Ctx[k, v]) toKeyStrs(keys ...k) (KeyStrs []string, err error) {
	var keyStr string
	for _, key := range keys {
		if keyStr, err = ctx.toKeyStr(key); err != nil {
			return nil, err
		}
		KeyStrs = append(KeyStrs, keyStr)
	}
	return KeyStrs, nil
}

func (ctx *Ctx[k, v]) keyToInterfaceSlice(keys ...v) (KeyStrs []interface{}, err error) {
	var keyStr string
	for _, key := range keys {
		if keyStr, err = ctx.toValueStr(key); err != nil {
			return nil, err
		}
		KeyStrs = append(KeyStrs, keyStr)
	}
	return KeyStrs, nil
}
func (ctx *Ctx[k, v]) valueToInterfaceSlice(values ...v) (ValueStrs []interface{}, err error) {
	var valueStr string
	for _, value := range values {
		if valueStr, err = ctx.toValueStr(value); err != nil {
			return nil, err
		}
		ValueStrs = append(ValueStrs, valueStr)
	}
	return ValueStrs, nil
}
