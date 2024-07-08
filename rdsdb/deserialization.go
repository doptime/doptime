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
func (ctx *Ctx[k, v]) toValues(valStr ...string) (values []v, err error) {
	if values = make([]v, len(valStr)); len(valStr) == 0 {
		return values, nil
	}
	var _v v
	kindOfv := reflect.TypeOf(_v).Kind()
	valueStruct := reflect.TypeOf((*v)(nil)).Elem()
	isElemPtr := valueStruct.Kind() == reflect.Ptr

	switch kindOfv {
	case reflect.Uint64, reflect.Interface:
		for i, val := range valStr {
			if vint, err := strconv.ParseInt(string(val), 10, 64); err != nil {
				values[i] = interface{}(uint64(vint)).(v)
			} else if i == len(valStr)-1 {
				return values, nil
			} else {
				break
			}
		}
	case reflect.Uint:
		for i, val := range valStr {
			if vint, err := strconv.ParseInt(string(val), 10, 64); err == nil {
				values[i] = interface{}(uint(vint)).(v)
			} else if i == len(valStr)-1 {
				return values, nil
			} else {
				break
			}
		}
	case reflect.Uint8:
		for i, val := range valStr {
			if vint, err := strconv.ParseInt(string(val), 10, 64); err == nil {
				values[i] = interface{}(uint8(vint)).(v)
			} else {
				break
			}
		}
	case reflect.Uint16:
		for i, val := range valStr {
			if vint, err := strconv.ParseInt(string(val), 10, 64); err == nil {
				values[i] = interface{}(uint16(vint)).(v)
			} else if i == len(valStr)-1 {
				return values, nil
			} else {
				break
			}
		}
	case reflect.Uint32:
		for i, val := range valStr {
			if vint, err := strconv.ParseInt(string(val), 10, 64); err == nil {
				values[i] = interface{}(uint32(vint)).(v)
			} else if i == len(valStr)-1 {
				return values, nil
			} else {
				break
			}
		}
	}

	switch kindOfv {
	case reflect.Int64, reflect.Interface:
		for i, val := range valStr {
			if vint, err := strconv.ParseInt(string(val), 10, 64); err == nil {
				values[i] = interface{}(vint).(v)
			} else if i == len(valStr)-1 {
				return values, nil
			} else {
				break
			}
		}
	case reflect.Int:
		for i, val := range valStr {
			if vint, err := strconv.ParseInt(string(val), 10, 64); err == nil {
				values[i] = interface{}(int(vint)).(v)
			} else if i == len(valStr)-1 {
				return values, nil
			} else {
				break
			}
		}
	case reflect.Int8:
		for i, val := range valStr {
			if vint, err := strconv.ParseInt(string(val), 10, 64); err == nil {
				values[i] = interface{}(int8(vint)).(v)
			} else if i == len(valStr)-1 {
				return values, nil
			} else {
				break
			}
		}
	case reflect.Int16:
		for i, val := range valStr {
			if vint, err := strconv.ParseInt(string(val), 10, 64); err == nil {
				values[i] = interface{}(int16(vint)).(v)
			} else {
				break
			}
		}
	case reflect.Int32:
		for i, val := range valStr {
			if vint, err := strconv.ParseInt(string(val), 10, 64); err == nil {
				values[i] = interface{}(int32(vint)).(v)
			} else {
				break
			}
		}
	}

	switch kindOfv {
	case reflect.Float64, reflect.Interface:
		for i, val := range valStr {
			if vfloat, err := strconv.ParseFloat(string(val), 64); err == nil {
				values[i] = interface{}(vfloat).(v)
			} else if i == len(valStr)-1 {
				return values, nil
			} else {
				break
			}
		}
	case reflect.Float32:
		for i, val := range valStr {
			if vfloat, err := strconv.ParseFloat(string(val), 64); err == nil {
				values[i] = interface{}(float32(vfloat)).(v)
			} else if i == len(valStr)-1 {
				return values, nil
			} else {
				break
			}
		}
	}

	switch kindOfv {
	case reflect.Bool, reflect.Interface:
		for i, val := range valStr {
			if bval, err := strconv.ParseBool(string(val)); err == nil {
				values[i] = interface{}(bval).(v)
			} else if i == len(valStr)-1 {
				return values, nil
			} else {
				break
			}
		}
	}

	//continue with msgpack unmarshal
	if isElemPtr {
		for i, val := range valStr {
			values[i] = reflect.New(valueStruct.Elem()).Interface().(v)
			if err = msgpack.Unmarshal([]byte(val), values[i]); err != nil {
				break
			} else if i == len(valStr)-1 {
				return values, nil
			}
		}
	} else {
		for i, val := range valStr {
			if err := msgpack.Unmarshal([]byte(val), &values[i]); err != nil {
				break
			} else if i == len(valStr)-1 {
				return values, nil
			}
		}
	}
	return values, fmt.Errorf("fail convert redis data to value")
}
func (ctx *Ctx[k, v]) toValue(valbytes []byte) (value v, err error) {
	vTypeKind := reflect.TypeOf(value).Kind()
	switch vTypeKind {
	case reflect.Int64, reflect.Interface:
		if vint, err := strconv.ParseInt(string(valbytes), 10, 64); err == nil {
			return interface{}(vint).(v), nil
		}
	case reflect.Int:
		if vint, err := strconv.ParseInt(string(valbytes), 10, 64); err == nil {
			return interface{}(int(vint)).(v), nil
		}
	case reflect.Int8:
		if vint, err := strconv.ParseInt(string(valbytes), 10, 64); err == nil {
			return interface{}(int8(vint)).(v), nil
		}
	case reflect.Int16:
		if vint, err := strconv.ParseInt(string(valbytes), 10, 64); err == nil {
			return interface{}(int16(vint)).(v), nil
		}
	case reflect.Int32:
		if vint, err := strconv.ParseInt(string(valbytes), 10, 64); err == nil {
			return interface{}(int32(vint)).(v), nil
		}
	}

	switch vTypeKind {
	case reflect.Uint64, reflect.Interface:
		if vint, err := strconv.ParseInt(string(valbytes), 10, 64); err == nil {
			return interface{}(uint64(vint)).(v), nil
		}
	case reflect.Uint:
		if vint, err := strconv.ParseInt(string(valbytes), 10, 64); err == nil {
			return interface{}(uint(vint)).(v), nil
		}
	case reflect.Uint8:
		if vint, err := strconv.ParseInt(string(valbytes), 10, 64); err == nil {
			return interface{}(uint8(vint)).(v), nil
		}
	case reflect.Uint16:
		if vint, err := strconv.ParseInt(string(valbytes), 10, 64); err == nil {
			return interface{}(uint16(vint)).(v), nil
		}
	case reflect.Uint32:
		if vint, err := strconv.ParseInt(string(valbytes), 10, 64); err == nil {
			return interface{}(uint32(vint)).(v), nil
		}
	}

	switch vTypeKind {
	case reflect.Float64, reflect.Interface:
		if vfloat, err := strconv.ParseFloat(string(valbytes), 64); err == nil {
			return interface{}(vfloat).(v), nil
		}
	case reflect.Float32:
		if vfloat, err := strconv.ParseFloat(string(valbytes), 64); err == nil {
			return interface{}(float32(vfloat)).(v), nil
		}
	}

	switch vTypeKind {
	case reflect.Bool, reflect.Interface:
		if bval, err := strconv.ParseBool(string(valbytes)); err == nil {
			return interface{}(bval).(v), nil
		}
	}

	switch vTypeKind {
	case reflect.String:
		return interface{}(string(valbytes)).(v), nil
	}
	//try deserialization with msgpack

	valueStruct := reflect.TypeOf((*v)(nil)).Elem()
	isElemPtr := valueStruct.Kind() == reflect.Ptr
	if isElemPtr {
		value = reflect.New(valueStruct.Elem()).Interface().(v)
		if err = msgpack.Unmarshal(valbytes, value); err == nil {
			return value, nil
		}
	} else {
		if err = msgpack.Unmarshal(valbytes, &value); err == nil {
			return value, nil
		}
	}
	return value, fmt.Errorf("fail convert redis data to value")
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
