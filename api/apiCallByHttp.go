package api

import (
	"github.com/doptime/doptime/specification"
	"github.com/rs/zerolog/log"
)

func GetApiInfoByName(serviceName string) (apiInfo *ApiInfo, ok bool) {
	var stdServiceName string
	if stdServiceName = specification.ApiName(serviceName); len(stdServiceName) == 0 {
		log.Error().Str("service misnamed", stdServiceName).Send()
		return nil, false
	}
	return ApiServices.Get(stdServiceName)
}

func (apiInfo *ApiInfo) CallByHTTP(paramIn map[string]interface{}) (ret interface{}, err error) {
	var (
		buf []byte
	)
	//if function is stored locally, call it directly. This is alias monolithic mode
	if buf, err = specification.MarshalApiInput(paramIn); err != nil {
		return nil, err
	}
	return apiInfo.ApiFuncWithMsgpackedParam(buf)
}
