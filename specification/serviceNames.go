package specification

import (
	"math/big"
	"math/rand"
	"reflect"
	"strings"

	"github.com/doptime/logger"
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
func ApiName(nameOld string) (nameNew string) {
	nameNew = nameOld
	//remove postfix
	nameLowercase := "      " + strings.ToLower(nameOld)
	if p := nameLowercase[len(nameLowercase)-6:]; p == "output" {
		nameNew = nameOld[:len(nameOld)-6]
	} else if p := nameLowercase[len(nameLowercase)-5:]; p == "input" || p == "param" {
		nameNew = nameOld[:len(nameOld)-5]
	} else if p := nameLowercase[len(nameLowercase)-4:]; p == "data" {
		nameNew = nameOld[:len(nameOld)-4]
	} else if p := nameLowercase[len(nameLowercase)-3:]; p == "arg" || p == "req" || p == "src" || p == "out" {
		nameNew = nameOld[:len(nameOld)-3]
	} else if p := nameLowercase[len(nameLowercase)-2:]; p == "in" {
		nameNew = nameOld[:len(nameOld)-2]
	}

	//remove prefix
	nameLowercase = strings.ToLower(nameNew) + "      "
	if p := nameLowercase[:5]; p == "input" || p == "param" {
		nameNew = nameNew[5:]
	} else if p := nameLowercase[:4]; p == "api:" || p == "data" {
		nameNew = nameNew[4:]
	} else if p := nameLowercase[:3]; p == "arg" || p == "req" || p == "src" {
		nameNew = nameNew[3:]
	} else if p := nameLowercase[:2]; p == "in" {
		nameNew = nameNew[2:]
	}

	if _, ok := DisAllowedServiceNames[nameNew]; ok || len(nameNew) == 0 {
		nameNew = big.NewInt(rand.Int63()).String()
		logger.Error().Msg("Service created failed when calling ApiNamed, service name " + nameOld + " is invalid , it's renamed to " + nameNew)
	}

	//first byte of ServiceName should be lower case
	if nameNew[0] >= 'A' && nameNew[0] <= 'Z' {
		nameNew = string(nameNew[0]+32) + nameNew[1:]
	}
	//ensure ServiceKey start with "api:"
	return "api:" + nameNew
}

func ApiNameByType(i interface{}) (name string) {
	//get default ServiceName
	var _type reflect.Type
	//take name of type v as key
	for _type = reflect.TypeOf(i); _type.Kind() == reflect.Ptr || _type.Kind() == reflect.Array; _type = _type.Elem() {
	}
	return ApiName(_type.Name())

}
