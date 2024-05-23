package specification

import (
	"math/big"
	"math/rand"
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
func ApiName(ServiceNameOriginal string) (ServiceName string) {
	//remove postfix
	nameLowercase := "      " + strings.ToLower(ServiceNameOriginal)
	if p := nameLowercase[len(nameLowercase)-6:]; p == "output" {
		ServiceName = ServiceNameOriginal[:len(ServiceName)-6]
	} else if p := nameLowercase[len(nameLowercase)-5:]; p == "input" || p == "param" {
		ServiceName = ServiceNameOriginal[:len(ServiceName)-5]
	} else if p := nameLowercase[len(nameLowercase)-4:]; p == "data" {
		ServiceName = ServiceNameOriginal[:len(ServiceName)-4]
	} else if p := nameLowercase[len(nameLowercase)-3:]; p == "arg" || p == "req" || p == "src" || p == "out" {
		ServiceName = ServiceNameOriginal[:len(ServiceName)-3]
	} else if p := nameLowercase[len(nameLowercase)-2:]; p == "in" {
		ServiceName = ServiceNameOriginal[:len(ServiceName)-2]
	}

	//remove prefix
	nameLowercase = strings.ToLower(ServiceName) + "      "
	if p := nameLowercase[:5]; p == "input" || p == "param" {
		ServiceName = ServiceName[5:]
	} else if p := nameLowercase[:4]; p == "api:" || p == "data" {
		ServiceName = ServiceName[4:]
	} else if p := nameLowercase[:3]; p == "arg" || p == "req" || p == "src" {
		ServiceName = ServiceName[3:]
	} else if p := nameLowercase[:2]; p == "in" {
		ServiceName = ServiceName[2:]
	}

	if _, ok := DisAllowedServiceNames[ServiceName]; ok || len(ServiceName) == 0 {
		ServiceName = big.NewInt(rand.Int63()).String()
		dlog.Error().Msg("Service created failed when calling ApiNamed, service name " + ServiceNameOriginal + " is invalid , it's renamed to " + ServiceName)
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
