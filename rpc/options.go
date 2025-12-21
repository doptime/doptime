package rpc

import (
	"reflect"
	"strings"
)

// Option is parameter to create an API, RPC, or CallAt
type Option struct {
	ApiSourceRds  string
	ApiSourceHttp string
	ApiKey        string
}
type optionSetter func(*Option)

func WithApiRds(rdsSource string) optionSetter {
	return func(o *Option) {
		o.ApiSourceRds = rdsSource
	}
}
func WithApiHttp(httpSource string) optionSetter {
	return func(o *Option) {
		o.ApiSourceHttp = httpSource
	}
}
func WithApiDoptime(apiKey string) optionSetter {
	return func(o *Option) {
		o.ApiSourceHttp = "https://api.doptime.com/"
		o.ApiKey = apiKey
	}
}

func WithApiKey(apiKey string) optionSetter {
	return func(o *Option) {
		o.ApiKey = apiKey
	}
}

func (o Option) mergeNewOptions(optionSetters ...optionSetter) (out *Option) {
	for _, setter := range optionSetters {
		setter(&o)
	}
	return &o
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
		fieldName, tagLowercase := vType.Field(i).Name, strings.ToLower(vType.Field(i).Tag.Get("json"))
		if strings.HasPrefix(fieldName, "Header") || strings.Contains(tagLowercase, "header") {
			return true
		}
	}
	return false
}
