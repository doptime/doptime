package httpapi

import (
	"github.com/doptime/doptime/utils"
	"github.com/doptime/logger"

	cmap "github.com/orcaman/concurrent-map/v2"
)

var ApiViaHttp cmap.ConcurrentMap[string, ApiInterface] = cmap.New[ApiInterface]()

func GetApiByName(serviceName string) (apiInfo ApiInterface, ok bool) {
	var stdServiceName string
	if stdServiceName = utils.ApiName(serviceName); len(stdServiceName) == 0 {
		logger.Error().Str("service misnamed", stdServiceName).Send()
		return nil, false
	}
	return ApiViaHttp.Get(stdServiceName)
}
func apiServiceNames() (serviceNames []string) {
	for _, serviceInfo := range ApiViaHttp.Items() {
		serviceNames = append(serviceNames, serviceInfo.GetName())
	}
	return serviceNames
}

var Fun2Api cmap.ConcurrentMap[uintptr, ApiInterface] = cmap.NewWithCustomShardingFunction[uintptr, ApiInterface](func(key uintptr) uint32 {
	hash := uint32(2166136261)
	const prime32 = uint32(16777619)
	return ((hash*prime32)^uint32(key))*prime32 ^ uint32(key>>32)
})

func GetApiByFunc(f uintptr) (apiInfo ApiInterface, ok bool) {
	return Fun2Api.Get(f)
}
