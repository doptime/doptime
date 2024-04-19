package specification

import (
	"reflect"
	"strings"

	"github.com/doptime/doptime/dlog"
)

var DisAllowedServiceNames = map[string]bool{
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
	"map":        true,
}

// return the api name of the service
// name with format "api:serviceName". first letter of serviceName should be lower case, and start with "api:"
// two possible source of the service name:
// 1. the type name of the first parameter of the function
// 2. the name give by the user
// do not panic, because it may be called by web client. otherwise the server can be maliciously closed by the client
func ApiName(ServiceName string) string {
	var ServiceNameLowercase string
	//remove  prefix. "api:" is the case of encoded service name. other wise for the case of parameter type name
	if ServiceNameLowercase = strings.ToLower(ServiceName); len(ServiceNameLowercase) == 0 {
		dlog.Warn().Msg("Service created failed when calling ApiNamed, service name is empty")
		return ""
	}
	for _, prefix := range []string{"api:", "input", "in", "req", "arg", "param", "src", "data"} {
		if strings.HasPrefix(ServiceNameLowercase, prefix) {
			ServiceName = ServiceName[len(prefix):]
			break
		}
	}
	//remove postfix
	if ServiceNameLowercase = strings.ToLower(ServiceName); len(ServiceNameLowercase) == 0 {
		dlog.Warn().Msg("Service created failed when calling ApiNamed, service name is empty")
		return ""
	}
	for _, postfix := range []string{"input", "in", "req", "arg", "param", "src", "data"} {
		if strings.HasSuffix(ServiceName, postfix) {
			ServiceName = ServiceName[:len(ServiceName)-len(postfix)]
			break
		}
	}

	if _, ok := DisAllowedServiceNames[ServiceName]; ok {
		dlog.Error().Str("Service created failed when calling ApiNamed, service name disallowed", ServiceName).Send()
		return ""
	}
	//first byte of ServiceName should be lower case
	if ServiceName[0] >= 'A' && ServiceName[0] <= 'Z' {
		ServiceName = string(ServiceName[0]+32) + ServiceName[1:]
	}
	//ensure ServiceKey start with "api:"
	return "api:" + ServiceName
}

func ApiNameByType(i interface{}) (name string) {
	//get default ServiceName
	var _type reflect.Type
	//take name of type v as key
	for _type = reflect.TypeOf(i); _type.Kind() == reflect.Ptr || _type.Kind() == reflect.Array; _type = _type.Elem() {
	}
	return ApiName(_type.Name())

}
