package rdsdb

import (
	"encoding/json"
	"reflect"
	"strconv"

	"github.com/doptime/doptime/vars"
	"github.com/vmihailenco/msgpack/v5"
)

func (ctx *Ctx[k, v]) toKeyStr(key interface{}) (keyStr string, err error) {
	if vv := reflect.ValueOf(key); vv.Kind() == reflect.Ptr && vv.IsNil() {
		return "", nil
	} else if !vv.IsValid() {
		return keyStr, vars.ErrInvalidField
	}
	//if key is a string, directly append to keyBytes
	if strkey, ok := key.(string); ok {
		return strkey, nil
	} else if strkey, ok := key.(int); ok {
		return strconv.FormatInt(int64(strkey), 10), nil
	} else if strkey, ok := key.(int8); ok {
		return strconv.FormatInt(int64(strkey), 10), nil
	} else if strkey, ok := key.(int16); ok {
		return strconv.FormatInt(int64(strkey), 10), nil
	} else if strkey, ok := key.(int32); ok {
		return strconv.FormatInt(int64(strkey), 10), nil
	} else if strkey, ok := key.(int64); ok {
		return strconv.FormatInt(strkey, 10), nil
	} else if strkey, ok := key.(uint); ok {
		return strconv.FormatUint(uint64(strkey), 10), nil
	} else if strkey, ok := key.(uint8); ok {
		return strconv.FormatUint(uint64(strkey), 10), nil
	} else if strkey, ok := key.(uint16); ok {
		return strconv.FormatUint(uint64(strkey), 10), nil
	} else if strkey, ok := key.(uint32); ok {
		return strconv.FormatUint(uint64(strkey), 10), nil
	} else if strkey, ok := key.(uint64); ok {
		return strconv.FormatUint(strkey, 10), nil
	} else if strkey, ok := key.(float32); ok {
		return strconv.FormatFloat(float64(strkey), 'f', -1, 32), nil
	} else if strkey, ok := key.(float64); ok {
		return strconv.FormatFloat(strkey, 'f', -1, 64), nil
	} else if strkey, ok := key.(bool); ok {
		return strconv.FormatBool(strkey), nil
	} else if keyBytes, err := json.Marshal(key); err != nil {
		return keyStr, err
	} else {
		return string(keyBytes), nil
	}
}
func (ctx *Ctx[k, v]) toValueStrFun() func(value v) (valueStr string, err error) {
	var typeofv = reflect.TypeOf((*v)(nil)).Elem().Kind()
	switch typeofv {
	//type string
	case reflect.String:
		return func(value v) (valueStr string, err error) {
			return interface{}(value).(string), nil
		}
		//type int
	case reflect.Int:
		return func(value v) (valueStr string, err error) {
			return strconv.FormatInt(int64(interface{}(value).(int)), 10), nil
		}

	case reflect.Int8:
		return func(value v) (valueStr string, err error) {
			return strconv.FormatInt(int64(interface{}(value).(int8)), 10), nil
		}
	case reflect.Int16:
		return func(value v) (valueStr string, err error) {
			return strconv.FormatInt(int64(interface{}(value).(int16)), 10), nil
		}
	case reflect.Int32:
		return func(value v) (valueStr string, err error) {
			return strconv.FormatInt(int64(interface{}(value).(int32)), 10), nil
		}
	case reflect.Int64:
		return func(value v) (valueStr string, err error) {
			return strconv.FormatInt(interface{}(value).(int64), 10), nil
		}

		//case uint
	case reflect.Uint:
		return func(value v) (valueStr string, err error) {
			return strconv.FormatUint(uint64(interface{}(value).(uint)), 10), nil
		}

	case reflect.Uint8:
		return func(value v) (valueStr string, err error) {
			return strconv.FormatUint(uint64(interface{}(value).(uint8)), 10), nil
		}
	case reflect.Uint16:
		return func(value v) (valueStr string, err error) {
			return strconv.FormatUint(uint64(interface{}(value).(uint16)), 10), nil
		}
	case reflect.Uint32:
		return func(value v) (valueStr string, err error) {
			return strconv.FormatUint(uint64(interface{}(value).(uint32)), 10), nil
		}
	case reflect.Uint64:
		return func(value v) (valueStr string, err error) {
			return strconv.FormatUint(interface{}(value).(uint64), 10), nil
		}

		//case float
	case reflect.Float32:
		return func(value v) (valueStr string, err error) {
			return strconv.FormatFloat(float64(interface{}(value).(float32)), 'f', -1, 32), nil
		}

	case reflect.Float64:
		return func(value v) (valueStr string, err error) {
			return strconv.FormatFloat(interface{}(value).(float64), 'f', -1, 64), nil
		}

	case reflect.Bool:
		return func(value v) (valueStr string, err error) {
			return strconv.FormatBool(interface{}(value).(bool)), nil
		}
	default:
		return func(value v) (valueStr string, err error) {
			bytes, err := msgpack.Marshal(value)
			if err == nil {
				return string(bytes), nil
			}
			return valueStr, err
		}
	}
}

func (ctx *Ctx[k, v]) toKeyStrs(keys ...interface{}) (KeyStrs []string, err error) {
	var keyStr string
	for _, key := range keys {
		if keyStr, err = ctx.toKeyStr(key); err != nil {
			return nil, err
		}
		KeyStrs = append(KeyStrs, keyStr)
	}
	return KeyStrs, nil
}

func (ctx *Ctx[k, v]) toValueStrsSlice(values ...v) (ValueStrs []interface{}, err error) {
	var valueStr string
	for _, value := range values {
		if valueStr, err = ctx.MarshalValue(value); err != nil {
			return nil, err
		}
		ValueStrs = append(ValueStrs, valueStr)
	}
	return ValueStrs, nil
}
