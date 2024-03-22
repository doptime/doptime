package api

import (
	"reflect"
	"strings"

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

func HeaderFieldsUsed(vType reflect.Type) bool {
	//use reflect to detect if the param has a field start with "Header", or tag of that field contains "Header",if true return true else return false

	// case double pointer decoding
	for ; vType.Kind() == reflect.Ptr; vType = vType.Elem() {
	}
	if vType.Kind() != reflect.Struct {
		return false
	}

	for i := 0; i < vType.NumField(); i++ {
		fieldName, tagLowercase := vType.Field(i).Name, strings.ToLower(vType.Field(i).Tag.Get("mapstructure"))

		if strings.HasPrefix(fieldName, "Header") || strings.Contains(tagLowercase, "header") {
			return true
		}
	}
	return false
}

func WithJwtFields(vt reflect.Type) bool {
	for ; vt.Kind() == reflect.Ptr; vt = vt.Elem() {
	}
	if vt.Kind() != reflect.Struct {
		return false
	}
	// check if contains tag "validate"
	for i := 0; i < vt.NumField(); i++ {
		fieldName, tagLowercase := vt.Field(i).Name, strings.ToLower(vt.Field(i).Tag.Get("mapstructure"))
		if strings.HasPrefix(fieldName, "Jwt") || strings.Contains(tagLowercase, "jwt") {
			return true
		}
	}
	return false
}
