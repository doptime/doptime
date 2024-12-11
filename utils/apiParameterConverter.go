package utils

import (
	"reflect"
	"strconv"

	"github.com/mitchellh/mapstructure"
)

func mapToStructDecodHook(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	switch {
	// case string to ints
	case f.Kind() == reflect.String && t.Kind() == reflect.Int:
		return strconv.Atoi(data.(string))
	case f.Kind() == reflect.String && t.Kind() == reflect.Int64:
		return strconv.ParseInt(data.(string), 10, 64)
	case f.Kind() == reflect.String && t.Kind() == reflect.Int32:
		i, err := strconv.ParseInt(data.(string), 10, 32)
		return int32(i), err
	case f.Kind() == reflect.String && t.Kind() == reflect.Int16:
		i, err := strconv.ParseInt(data.(string), 10, 16)
		return int16(i), err
	case f.Kind() == reflect.String && t.Kind() == reflect.Int8:
		i, err := strconv.ParseInt(data.(string), 10, 8)
		return int8(i), err

	//case string to uints
	case f.Kind() == reflect.String && t.Kind() == reflect.Uint:
		i, err := strconv.ParseUint(data.(string), 10, 64)
		return uint(i), err
	case f.Kind() == reflect.String && t.Kind() == reflect.Uint64:
		return strconv.ParseUint(data.(string), 10, 64)
	case f.Kind() == reflect.String && t.Kind() == reflect.Uint32:
		i, err := strconv.ParseUint(data.(string), 10, 32)
		return uint32(i), err
	case f.Kind() == reflect.String && t.Kind() == reflect.Uint16:
		i, err := strconv.ParseUint(data.(string), 10, 16)
		return uint16(i), err
	case f.Kind() == reflect.String && t.Kind() == reflect.Uint8:
		i, err := strconv.ParseUint(data.(string), 10, 8)
		return uint8(i), err

	//case string to floats
	case f.Kind() == reflect.String && t.Kind() == reflect.Float64:
		return strconv.ParseFloat(data.(string), 64)
	case f.Kind() == reflect.String && t.Kind() == reflect.Float32:
		f, err := strconv.ParseFloat(data.(string), 32)
		return float32(f), err

	//case ints to strings
	case f.Kind() == reflect.Int && t.Kind() == reflect.String:
		return strconv.Itoa(data.(int)), nil
	case f.Kind() == reflect.Int64 && t.Kind() == reflect.String:
		return strconv.FormatInt(data.(int64), 10), nil
	case f.Kind() == reflect.Int32 && t.Kind() == reflect.String:
		return strconv.FormatInt(int64(data.(int32)), 10), nil
	case f.Kind() == reflect.Int16 && t.Kind() == reflect.String:
		return strconv.FormatInt(int64(data.(int16)), 10), nil
	case f.Kind() == reflect.Int8 && t.Kind() == reflect.String:
		return strconv.FormatInt(int64(data.(int8)), 10), nil

	//case uints to strings
	case f.Kind() == reflect.Uint && t.Kind() == reflect.String:
		return strconv.FormatUint(uint64(data.(uint)), 10), nil
	case f.Kind() == reflect.Uint64 && t.Kind() == reflect.String:
		return strconv.FormatUint(data.(uint64), 10), nil
	case f.Kind() == reflect.Uint32 && t.Kind() == reflect.String:
		return strconv.FormatUint(uint64(data.(uint32)), 10), nil
	case f.Kind() == reflect.Uint16 && t.Kind() == reflect.String:
		return strconv.FormatUint(uint64(data.(uint16)), 10), nil
	case f.Kind() == reflect.Uint8 && t.Kind() == reflect.String:
		return strconv.FormatUint(uint64(data.(uint8)), 10), nil

	//case floats to strings
	case f.Kind() == reflect.Float64 && t.Kind() == reflect.String:
		return strconv.FormatFloat(data.(float64), 'f', -1, 64), nil
	case f.Kind() == reflect.Float32 && t.Kind() == reflect.String:
		return strconv.FormatFloat(float64(data.(float32)), 'f', -1, 32), nil

	//case bool to strings
	case f.Kind() == reflect.Bool && t.Kind() == reflect.String:
		return strconv.FormatBool(data.(bool)), nil
	default:
		return data, nil
	}
}
func MapToStructDecoder(pIn interface{}) (decoder *mapstructure.Decoder, err error) {
	// mapstructure support type conversion
	config := &mapstructure.DecoderConfig{
		Metadata:   nil,
		Result:     pIn,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(mapToStructDecodHook),
	}

	if decoder, err = mapstructure.NewDecoder(config); err != nil {
		return nil, err
	}
	return decoder, nil
}
