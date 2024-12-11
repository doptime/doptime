package utils

import (
	"math/big"
	"math/rand"
	"reflect"
	"strings"

	"github.com/doptime/logger"
)

// 判断服务名是否被禁止
func IsInvalidStructName(serviceName string) bool {
	const DisAllowedServiceNames = ",string,int32,int64,float32,float64,int,uint,float,bool,byte,rune,complex64,complex128,map,"
	return strings.Contains(DisAllowedServiceNames, ","+serviceName+",")
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
	} else if p := nameLowercase[len(nameLowercase)-3:]; p == "arg" || p == "req" || p == "src" || p == "out" {
		nameNew = nameOld[:len(nameOld)-3]
	} else if p := nameLowercase[len(nameLowercase)-2:]; p == "in" {
		nameNew = nameOld[:len(nameOld)-2]
	}

	//remove prefix
	nameLowercase = strings.ToLower(nameNew) + "      "
	if p := nameLowercase[:5]; p == "input" || p == "param" {
		nameNew = nameNew[5:]
	} else if p := nameLowercase[:4]; p == "api:" {
		nameNew = nameNew[4:]
	} else if p := nameLowercase[:3]; p == "arg" || p == "req" || p == "src" {
		nameNew = nameNew[3:]
	} else if p := nameLowercase[:2]; p == "in" {
		nameNew = nameNew[2:]
	}

	if invalid := IsInvalidStructName(nameNew); invalid || len(nameNew) == 0 {
		nameNew = big.NewInt(rand.Int63()).String()
		logger.Error().Msg("Service created failed when calling ApiNamed, service name " + nameOld + " is invalid , it's renamed to " + nameNew)
	}

	//ensure ServiceKey start with "api:"
	return "api:" + strings.ToLower(nameNew)
}

func ApiNameByType(i interface{}) (name string) {
	//get default ServiceName
	var _type reflect.Type
	//take name of type v as key
	for _type = reflect.TypeOf(i); _type.Kind() == reflect.Ptr || _type.Kind() == reflect.Array; _type = _type.Elem() {
	}
	return ApiName(_type.Name())

}
