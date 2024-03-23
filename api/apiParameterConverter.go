package api

import (
	"reflect"
	"strconv"

	"github.com/mitchellh/mapstructure"
)

func mapToStructDecodHook(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	switch {
	case f.Kind() == reflect.String && t.Kind() == reflect.Int64:
		return strconv.ParseInt(data.(string), 10, 64)
	case f.Kind() == reflect.String && t.Kind() == reflect.Int:
		return strconv.Atoi(data.(string))
	case f.Kind() == reflect.String && t.Kind() == reflect.Int32:
		i, err := strconv.ParseInt(data.(string), 10, 32)
		return int32(i), err

	case f.Kind() == reflect.String && t.Kind() == reflect.Float64:
		return strconv.ParseFloat(data.(string), 64)
	case f.Kind() == reflect.String && t.Kind() == reflect.Float32:
		f, err := strconv.ParseFloat(data.(string), 32)
		if err != nil {
			return nil, err
		}
		return float32(f), nil

	case f.Kind() == reflect.Int64 && t.Kind() == reflect.String:
		return strconv.FormatInt(data.(int64), 10), nil
	case f.Kind() == reflect.Int && t.Kind() == reflect.String:
		return strconv.Itoa(data.(int)), nil
	case f.Kind() == reflect.Int32 && t.Kind() == reflect.String:
		return strconv.FormatInt(int64(data.(int32)), 10), nil

	case f.Kind() == reflect.Float64 && t.Kind() == reflect.String:
		return strconv.FormatFloat(data.(float64), 'f', -1, 64), nil
	case f.Kind() == reflect.Float32 && t.Kind() == reflect.String:
		return strconv.FormatFloat(float64(data.(float32)), 'f', -1, 32), nil
	case f.Kind() == reflect.Bool && t.Kind() == reflect.String:
		return strconv.FormatBool(data.(bool)), nil
	default:
		return data, nil
	}
}
func mapToStructDecoder(pIn interface{}) (decoder *mapstructure.Decoder, err error) {
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
