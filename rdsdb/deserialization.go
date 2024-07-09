package rdsdb

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"github.com/doptime/doptime/dlog"
	"github.com/vmihailenco/msgpack/v5"
)

func (ctx *Ctx[k, v]) toKeys(valStr []string) (keys []k, err error) {
	if _, ok := interface{}(valStr).([]k); ok {
		return interface{}(valStr).([]k), nil
	}
	if keys = make([]k, len(valStr)); len(valStr) == 0 {
		return keys, nil
	}
	keyStruct := reflect.TypeOf((*k)(nil)).Elem()
	isElemPtr := keyStruct.Kind() == reflect.Ptr

	//save all data to mapOut
	for i, val := range valStr {
		if isElemPtr {
			keys[i] = reflect.New(keyStruct.Elem()).Interface().(k)
			err = json.Unmarshal([]byte(val), keys[i])
		} else {
			//if key is type of string, just return string
			if keyStruct.Kind() == reflect.String {
				keys[i] = interface{}(string(val)).(k)
			} else {
				err = json.Unmarshal([]byte(val), &keys[i])
			}
		}
		if err != nil {
			dlog.Info().AnErr("HKeys: field unmarshal error:", err).Msgf("Key: %s", ctx.Key)
			continue
		}
	}
	return keys, nil
}

// unmarhsal using msgpack
func (ctx *Ctx[k, v]) toValuesFunc() func(valStrs []string) (value []v, err error) {
	valueStruct := reflect.TypeOf((*v)(nil)).Elem()
	var typeofv = valueStruct.Kind()
	isElemPtr := valueStruct.Kind() == reflect.Ptr

	switch typeofv {
	case reflect.Uint64:
		return func(valStrs []string) (values []v, err error) {
			values = make([]v, len(valStrs))
			for i, val := range valStrs {
				var vint int64
				if vint, err = strconv.ParseInt(val, 10, 64); err != nil {
					break
				}
				values[i] = interface{}(uint64(vint)).(v)
			}
			return values, err
		}
	case reflect.Uint:
		return func(valStrs []string) (values []v, err error) {
			values = make([]v, len(valStrs))
			for i, val := range valStrs {
				var vint int64
				if vint, err = strconv.ParseInt(val, 10, 64); err != nil {
					break
				}
				values[i] = interface{}(uint(vint)).(v)
			}
			return values, err
		}
	case reflect.Uint8:
		return func(valStrs []string) (values []v, err error) {
			values = make([]v, len(valStrs))
			for i, val := range valStrs {
				var vint int64
				if vint, err = strconv.ParseInt(val, 10, 64); err != nil {
					break
				}
				values[i] = interface{}(uint8(vint)).(v)
			}
			return values, err
		}
	case reflect.Uint16:
		return func(valStrs []string) (values []v, err error) {
			values = make([]v, len(valStrs))
			for i, val := range valStrs {
				var vint int64
				if vint, err = strconv.ParseInt(val, 10, 64); err != nil {
					break
				}
				values[i] = interface{}(uint16(vint)).(v)
			}
			return values, err
		}
	case reflect.Uint32:
		return func(valStrs []string) (values []v, err error) {
			values = make([]v, len(valStrs))
			for i, val := range valStrs {
				var vint int64
				if vint, err = strconv.ParseInt(val, 10, 64); err != nil {
					break
				}
				values[i] = interface{}(uint32(vint)).(v)
			}
			return values, err
		}

	case reflect.Int64:
		return func(valStrs []string) (values []v, err error) {
			values = make([]v, len(valStrs))
			for i, val := range valStrs {
				var vint int64
				if vint, err = strconv.ParseInt(val, 10, 64); err != nil {
					break
				}
				values[i] = interface{}(int64(vint)).(v)
			}
			return values, err
		}
	case reflect.Int:
		return func(valStrs []string) (values []v, err error) {
			values = make([]v, len(valStrs))
			for i, val := range valStrs {
				var vint int64
				if vint, err = strconv.ParseInt(val, 10, 64); err != nil {
					break
				}
				values[i] = interface{}(int(vint)).(v)
			}
			return values, err
		}
	case reflect.Int8:
		return func(valStrs []string) (values []v, err error) {
			values = make([]v, len(valStrs))
			for i, val := range valStrs {
				var vint int64
				if vint, err = strconv.ParseInt(val, 10, 64); err != nil {
					break
				}
				values[i] = interface{}(int8(vint)).(v)
			}
			return values, err
		}
	case reflect.Int16:
		return func(valStrs []string) (values []v, err error) {
			values = make([]v, len(valStrs))
			for i, val := range valStrs {
				var vint int64
				if vint, err = strconv.ParseInt(val, 10, 64); err != nil {
					break
				}
				values[i] = interface{}(int16(vint)).(v)
			}
			return values, err
		}
	case reflect.Int32:
		return func(valStrs []string) (values []v, err error) {
			values = make([]v, len(valStrs))
			for i, val := range valStrs {
				var vint int64
				if vint, err = strconv.ParseInt(val, 10, 64); err != nil {
					break
				}
				values[i] = interface{}(int32(vint)).(v)
			}
			return values, err
		}

	case reflect.Float64:
		return func(valStrs []string) (values []v, err error) {
			values = make([]v, len(valStrs))
			for i, val := range valStrs {
				var vfloat float64
				if vfloat, err = strconv.ParseFloat(val, 64); err != nil {
					break
				}
				values[i] = interface{}(vfloat).(v)
			}
			return values, err
		}
	case reflect.Float32:
		return func(valStrs []string) (values []v, err error) {
			values = make([]v, len(valStrs))
			for i, val := range valStrs {
				var vfloat float64
				if vfloat, err = strconv.ParseFloat(val, 64); err != nil {
					break
				}
				values[i] = interface{}(float32(vfloat)).(v)
			}
			return values, err
		}
	case reflect.Bool:
		return func(valStrs []string) (values []v, err error) {
			values = make([]v, len(valStrs))
			for i, val := range valStrs {
				var vb bool
				if vb, err = strconv.ParseBool(val); err != nil {
					break
				}
				values[i] = interface{}(vb).(v)
			}
			return values, err
		}
	case reflect.String:
		return func(valStrs []string) (values []v, err error) {
			values = make([]v, len(valStrs))
			for i, val := range valStrs {
				values[i] = interface{}(string(val)).(v)
			}
			return values, err
		}
	default:
		//continue with msgpack unmarshal
		if isElemPtr {
			return func(valStrs []string) (values []v, err error) {
				values = make([]v, len(valStrs))

				for i, val := range valStrs {
					values[i] = reflect.New(valueStruct.Elem()).Interface().(v)
					if err = msgpack.Unmarshal([]byte(val), values[i]); err != nil {
						break
					}
				}
				return values, err
			}
		} else {
			return func(valStrs []string) (values []v, err error) {
				values = make([]v, len(valStrs))
				for i, val := range valStrs {
					if err = msgpack.Unmarshal([]byte(val), &values[i]); err != nil {
						break
					}
				}
				return values, err
			}
		}
	}
}
func (ctx *Ctx[k, v]) toValueFunc() func(valbytes []byte) (value v, err error) {
	vTypeKind := reflect.TypeOf((*v)(nil)).Elem().Kind()
	switch vTypeKind {
	case reflect.Int64:
		return func(valbytes []byte) (value v, err error) {
			vint, err := strconv.ParseInt(string(valbytes), 10, 64)
			return interface{}(vint).(v), err
		}
	case reflect.Int:
		return func(valbytes []byte) (value v, err error) {
			vint, err := strconv.ParseInt(string(valbytes), 10, 64)
			return interface{}(int(vint)).(v), err
		}
	case reflect.Int8:
		return func(valbytes []byte) (value v, err error) {
			vint, err := strconv.ParseInt(string(valbytes), 10, 64)
			return interface{}(int8(vint)).(v), err
		}
	case reflect.Int16:
		return func(valbytes []byte) (value v, err error) {
			vint, err := strconv.ParseInt(string(valbytes), 10, 64)
			return interface{}(int16(vint)).(v), err
		}
	case reflect.Int32:
		return func(valbytes []byte) (value v, err error) {
			vint, err := strconv.ParseInt(string(valbytes), 10, 64)
			return interface{}(int32(vint)).(v), err
		}
	case reflect.Uint64:
		return func(valbytes []byte) (value v, err error) {
			vint, err := strconv.ParseInt(string(valbytes), 10, 64)
			return interface{}(uint64(vint)).(v), err
		}
	case reflect.Uint:
		return func(valbytes []byte) (value v, err error) {
			vint, err := strconv.ParseInt(string(valbytes), 10, 64)
			return interface{}(uint(vint)).(v), err
		}
	case reflect.Uint8:
		return func(valbytes []byte) (value v, err error) {
			vint, err := strconv.ParseInt(string(valbytes), 10, 64)
			return interface{}(uint8(vint)).(v), err
		}
	case reflect.Uint16:
		return func(valbytes []byte) (value v, err error) {
			vint, err := strconv.ParseInt(string(valbytes), 10, 64)
			return interface{}(uint16(vint)).(v), err
		}
	case reflect.Uint32:
		return func(valbytes []byte) (value v, err error) {
			vint, err := strconv.ParseInt(string(valbytes), 10, 64)
			return interface{}(uint32(vint)).(v), err
		}
	case reflect.Float64:
		return func(valbytes []byte) (value v, err error) {
			vfloat, err := strconv.ParseFloat(string(valbytes), 64)
			return interface{}(vfloat).(v), err
		}
	case reflect.Float32:
		return func(valbytes []byte) (value v, err error) {
			vfloat, err := strconv.ParseFloat(string(valbytes), 64)
			return interface{}(float32(vfloat)).(v), err
		}
	case reflect.Bool:
		return func(valbytes []byte) (value v, err error) {
			bval, err := strconv.ParseBool(string(valbytes))
			return interface{}(bval).(v), err
		}
	case reflect.String:
		return func(valbytes []byte) (value v, err error) {
			return interface{}(string(valbytes)).(v), err
		}
	case reflect.Interface:
		valueStruct := reflect.TypeOf((*v)(nil)).Elem()
		isElemPtr := valueStruct.Kind() == reflect.Ptr
		if isElemPtr {
			return func(valbytes []byte) (value v, err error) {
				value = reflect.New(valueStruct.Elem()).Interface().(v)
				if err = msgpack.Unmarshal(valbytes, value); err == nil {
					return value, nil
				}
				return interface{}(string(valbytes)).(v), nil
			}
		} else {
			return func(valbytes []byte) (value v, err error) {
				if err = msgpack.Unmarshal(valbytes, &value); err == nil {
					return value, nil
				}
				return interface{}(string(valbytes)).(v), nil
			}
		}
	default:
		valueStruct := reflect.TypeOf((*v)(nil)).Elem()
		isElemPtr := valueStruct.Kind() == reflect.Ptr
		if isElemPtr {
			return func(valbytes []byte) (value v, err error) {
				value = reflect.New(valueStruct.Elem()).Interface().(v)
				if err = msgpack.Unmarshal(valbytes, value); err == nil {
					return value, nil
				}
				return value, fmt.Errorf("fail convert redis data to value")
			}
		} else {
			return func(valbytes []byte) (value v, err error) {
				if err = msgpack.Unmarshal(valbytes, &value); err == nil {
					return value, nil
				}
				return value, fmt.Errorf("fail convert redis data to value")
			}
		}
	}

}

func (ctx *Ctx[k, v]) toKey(valBytes []byte) (key k, err error) {
	keyStruct := reflect.TypeOf((*k)(nil)).Elem()
	isElemPtr := keyStruct.Kind() == reflect.Ptr
	if isElemPtr {
		key = reflect.New(keyStruct.Elem()).Interface().(k)
		return key, json.Unmarshal(valBytes, key)
	} else {
		//if key is type of string, just return string
		if keyStruct.Kind() == reflect.String {
			return reflect.ValueOf(string(valBytes)).Interface().(k), nil
		}
		err = json.Unmarshal(valBytes, &key)
		return key, err
	}
}
