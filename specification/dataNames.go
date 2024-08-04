package specification

import (
	"fmt"
	"reflect"
)

var DisAllowedDataKeyNames = map[string]bool{
	"":           true,
	"string":     true,
	"int32":      true,
	"int64":      true,
	"float32":    true,
	"float64":    true,
	"int":        true,
	"uint":       true,
	"float":      true,
	"bool":       true,
	"byte":       true,
	"rune":       true,
	"complex64":  true,
	"complex128": true,
}

func GetValidDataKeyName(value interface{}) (Key string, err error) {
	if len(Key) == 0 {
		//get default ServiceName
		var _type reflect.Type
		//take name of type v as key
		for _type = reflect.TypeOf(value); _type.Kind() == reflect.Ptr || _type.Kind() == reflect.Array; _type = _type.Elem() {
		}
		Key = _type.Name()
	}
	if _, ok := DisAllowedDataKeyNames[Key]; ok {
		err = fmt.Errorf("invalid keyname infered from type: " + Key)
		return "", err
	}
	return Key, nil
}
