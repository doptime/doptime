package utils

import (
	"fmt"
	"reflect"
)

func GetValidDataKeyName(value interface{}) (Key string, err error) {
	if len(Key) == 0 {
		//get default ServiceName
		var _type reflect.Type
		//take name of type v as key
		for _type = reflect.TypeOf(value); _type.Kind() == reflect.Ptr || _type.Kind() == reflect.Array; _type = _type.Elem() {
		}
		Key = _type.Name()
	}
	if invalid := IsInvalidStructName(Key); invalid || len(Key) == 0 {
		err = fmt.Errorf("invalid keyname infered from type: %s", Key)
		return "", err
	}
	return Key, nil
}
