package api

import (
	"github.com/doptime/doptime/config"
	"github.com/doptime/doptime/specification"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

var ApiServices cmap.ConcurrentMap[string, ApiInterface] = cmap.New[ApiInterface]()

func GetApiByName(serviceName string) (apiInfo ApiInterface, ok bool) {
	var stdServiceName string
	if stdServiceName = specification.ApiName(serviceName); len(stdServiceName) == 0 {
		log.Error().Str("service misnamed", stdServiceName).Send()
		return nil, false
	}
	return ApiServices.Get(stdServiceName)
}

func GetServiceDB(serviceName string) (db *redis.Client) {
	var (
		err error
	)
	serviceInfo, _ := ApiServices.Get(serviceName)
	DataSource := serviceInfo.GetDataSource()
	if db, err = config.GetRdsClientByName(DataSource); err != nil {
		log.Panic().Str("DataSource not defined in enviroment. Please check the configuration", DataSource).Send()
	}
	return db
}
func apiServiceNames() (serviceNames []string) {
	for _, serviceInfo := range ApiServices.Items() {
		serviceNames = append(serviceNames, serviceInfo.GetName())
	}
	return serviceNames
}

var fun2Api cmap.ConcurrentMap[uintptr, ApiInterface] = cmap.NewWithCustomShardingFunction[uintptr, ApiInterface](func(key uintptr) uint32 {
	hash := uint32(2166136261)
	const prime32 = uint32(16777619)
	return ((hash*prime32)^uint32(key))*prime32 ^ uint32(key>>32)
})

func GetApiByFunc(f uintptr) (apiInfo ApiInterface, ok bool) {
	return fun2Api.Get(f)
}

var APIGroupByRdsToReceiveJob = cmap.New[[]string]()
