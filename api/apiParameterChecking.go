package api

import (
	"reflect"

	"github.com/go-playground/validator/v10"
)

func needValidate(v reflect.Type) func(s interface{}) (err error) {
	var (
		validate              *validator.Validate = validator.New()
		isStruct, hasValidTag bool
	)
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if isStruct = v.Kind() == reflect.Struct; isStruct {
		// check if contains tag "validate"
		for i := 0; i < v.NumField(); i++ {
			if v.Field(i).Tag.Get("validate") != "" {
				hasValidTag = true
				break
			}
		}
	}
	if isStruct && hasValidTag {
		return validate.Struct
	}
	return func(s interface{}) (err error) {
		return nil
	}
}
