package rdsdb

import (
	"encoding/json"
	"reflect"
	"strconv"

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
	switch typeOfV := reflect.TypeOf(value); typeOfV.Kind() {
	//type string
	case reflect.String:
		return interface{}(value).(string), nil
		//type int
	case reflect.Int:
		return strconv.FormatInt(int64(interface{}(value).(int)), 10), nil
	case reflect.Int8:
		return strconv.FormatInt(int64(interface{}(value).(int8)), 10), nil
	case reflect.Int16:
		return strconv.FormatInt(int64(interface{}(value).(int16)), 10), nil
	case reflect.Int32:
		return strconv.FormatInt(int64(interface{}(value).(int32)), 10), nil
	case reflect.Int64:
		return strconv.FormatInt(interface{}(value).(int64), 10), nil

		//case uint
	case reflect.Uint:
		return strconv.FormatUint(uint64(interface{}(value).(uint)), 10), nil
	case reflect.Uint8:
		return strconv.FormatUint(uint64(interface{}(value).(uint8)), 10), nil
	case reflect.Uint16:
		return strconv.FormatUint(uint64(interface{}(value).(uint16)), 10), nil
	case reflect.Uint32:
		return strconv.FormatUint(uint64(interface{}(value).(uint32)), 10), nil
	case reflect.Uint64:
		return strconv.FormatUint(interface{}(value).(uint64), 10), nil

		//case float
	case reflect.Float32:
		return strconv.FormatFloat(float64(interface{}(value).(float32)), 'f', -1, 32), nil
	case reflect.Float64:
		return strconv.FormatFloat(interface{}(value).(float64), 'f', -1, 64), nil
	case reflect.Bool:
		return strconv.FormatBool(interface{}(value).(bool)), nil
	default:
		bytes, err := msgpack.Marshal(value)
		if err != nil {
			return valueStr, err
		}
		return string(bytes), nil
	}
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
