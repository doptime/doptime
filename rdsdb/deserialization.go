package rdsdb

import (
	"encoding/json"
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
	typeofv := reflect.TypeOf(_v)
	valueStruct := reflect.TypeOf((*v)(nil)).Elem()
	isElemPtr := valueStruct.Kind() == reflect.Ptr

	typeId := ctx.getKeyTypeIdentifier()

	if typeId == 's' {
		if typeofv.Kind() == reflect.String || typeofv.Kind() == reflect.Interface {
			for i, val := range valStr {
				values[i], _ = interface{}(string(val)).(v)
			}
		}

	} else if typeId == 'i' {
		switch typeofv.Kind() {
		case reflect.Int64:
			for i, val := range valStr {
				if vint, err := strconv.ParseInt(string(val), 10, 64); err == nil {
					values[i] = interface{}(vint).(v)
				}
			}
		case reflect.Int:
			for i, val := range valStr {
				if vint, err := strconv.ParseInt(string(val), 10, 64); err == nil {
					values[i] = interface{}(int(vint)).(v)
				}
			}
		case reflect.Int8:
			for i, val := range valStr {
				if vint, err := strconv.ParseInt(string(val), 10, 64); err == nil {
					values[i] = interface{}(int8(vint)).(v)
				}
			}
		case reflect.Int16:
			for i, val := range valStr {
				if vint, err := strconv.ParseInt(string(val), 10, 64); err == nil {
					values[i] = interface{}(int16(vint)).(v)
				}
			}
		case reflect.Int32:
			for i, val := range valStr {
				if vint, err := strconv.ParseInt(string(val), 10, 64); err == nil {
					values[i] = interface{}(int32(vint)).(v)
				}
			}
		}

	} else if typeId == 'u' {
		switch typeofv.Kind() {
		case reflect.Uint64:
			for i, val := range valStr {
				if vint, err := strconv.ParseInt(string(val), 10, 64); err == nil {
					values[i] = interface{}(uint64(vint)).(v)
				}
			}
		case reflect.Uint:
			for i, val := range valStr {
				if vint, err := strconv.ParseInt(string(val), 10, 64); err == nil {
					values[i] = interface{}(uint(vint)).(v)
				}
			}
		case reflect.Uint8:
			for i, val := range valStr {
				if vint, err := strconv.ParseInt(string(val), 10, 64); err == nil {
					values[i] = interface{}(uint8(vint)).(v)
				}
			}
		case reflect.Uint16:
			for i, val := range valStr {
				if vint, err := strconv.ParseInt(string(val), 10, 64); err == nil {
					values[i] = interface{}(uint16(vint)).(v)
				}
			}
		case reflect.Uint32:
			for i, val := range valStr {
				if vint, err := strconv.ParseInt(string(val), 10, 64); err == nil {
					values[i] = interface{}(uint32(vint)).(v)
				}
			}
		}

	} else if typeId == 'f' {
		switch typeofv.Kind() {
		case reflect.Float64:
			for i, val := range valStr {
				if vfloat, err := strconv.ParseFloat(string(val), 64); err == nil {
					values[i] = interface{}(vfloat).(v)
				}
			}
		case reflect.Float32:
			for i, val := range valStr {
				if vfloat, err := strconv.ParseFloat(string(val), 64); err == nil {
					values[i] = interface{}(float32(vfloat)).(v)
				}
			}
		}
	} else if typeId == 'b' {
		for i, val := range valStr {
			if bval, err := strconv.ParseBool(string(val)); err == nil {
				values[i] = interface{}(bval).(v)
			}
		}
	}
	//continue with msgpack unmarshal
	if isElemPtr {
		for i := range values {
			values[i] = reflect.New(valueStruct.Elem()).Interface().(v)
		}
	}
	for i, val := range valStr {
		if err := msgpack.Unmarshal([]byte(val), &values[i]); err != nil {
			return nil, err
		}
	}
	return values, nil
}
func (ctx *Ctx[k, v]) toValue(valbytes []byte) (value v, err error) {
	typeId := ctx.getKeyTypeIdentifier()
	if typeId == 's' {
		return interface{}(string(valbytes)).(v), nil
	} else if typeId == 'i' {
		vint, err := strconv.ParseInt(string(valbytes), 10, 64)
		if err == nil {
			switch reflect.TypeOf(value).Kind() {
			case reflect.Int64:
				return interface{}(vint).(v), nil
			case reflect.Int:
				return interface{}(int(vint)).(v), nil
			case reflect.Int8:
				return interface{}(int8(vint)).(v), nil
			case reflect.Int16:
				return interface{}(int16(vint)).(v), nil
			case reflect.Int32:
				return interface{}(int32(vint)).(v), nil
			}
		}

	} else if typeId == 'u' {
		vint, err := strconv.ParseInt(string(valbytes), 10, 64)
		if err == nil {
			switch reflect.TypeOf(value).Kind() {
			case reflect.Uint64:
				return interface{}(vint).(v), nil
			case reflect.Uint:
				return interface{}(uint(vint)).(v), nil
			case reflect.Uint8:
				return interface{}(uint8(vint)).(v), nil
			case reflect.Uint16:
				return interface{}(uint16(vint)).(v), nil
			case reflect.Uint32:
				return interface{}(uint32(vint)).(v), nil
			}
		}

	} else if typeId == 'f' {
		vfloat, err := strconv.ParseFloat(string(valbytes), 64)
		if err == nil {
			switch reflect.TypeOf(value).Kind() {
			case reflect.Float64:
				return interface{}(vfloat).(v), nil
			case reflect.Float32:
				return interface{}(float32(vfloat)).(v), nil
			}
		}
	} else if typeId == 'b' {
		bval, err := strconv.ParseBool(string(valbytes))
		if err == nil {
			//both type bool and interface{} are supported
			return interface{}(bval).(v), nil
		}
	}
	//continue with msgpack unmarshal
	valueStruct := reflect.TypeOf((*v)(nil)).Elem()
	isElemPtr := valueStruct.Kind() == reflect.Ptr
	if isElemPtr {
		value = reflect.New(valueStruct.Elem()).Interface().(v)
		return value, msgpack.Unmarshal(valbytes, value)
	} else {
		err = msgpack.Unmarshal(valbytes, &value)
		return value, err
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
