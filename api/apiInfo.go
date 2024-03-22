package api

import (
	"context"
	"net/http"
	"sync"

	"github.com/doptime/doptime/config"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type ApiInfo struct {
	// Name is the name of the service
	Name       string
	DataSource string
	WithHeader bool
	WithJwt    bool
	Ctx        context.Context
	// ApiFuncWithMsgpackedParam is the function of the service
	ApiFuncWithMsgpackedParam func(s []byte) (ret interface{}, err error)
}

func (apiInfo *ApiInfo) MergeHeader(req *http.Request, paramIn map[string]interface{}) {
	//copy fields from req to paramIn
	for key, value := range req.Header {
		if len(value) > 1 {
			paramIn["Header"+key] = value
		} else {
			paramIn["Header"+key] = value[0]
		}
	}
	// copy ip address from req to paramIn
	paramIn["Header"+"RemoteAddr"] = req.RemoteAddr
	paramIn["Header"+"Host"] = req.Host
	paramIn["Header"+"Method"] = req.Method
	paramIn["Header"+"Path"] = req.URL.Path
	paramIn["Header"+"RawQuery"] = req.URL.RawQuery

}

var ApiServices cmap.ConcurrentMap[string, *ApiInfo] = cmap.New[*ApiInfo]()

func apiServiceNames() (serviceNames []string) {
	for _, serviceInfo := range ApiServices.Items() {
		serviceNames = append(serviceNames, serviceInfo.Name)
	}
	return serviceNames
}
func GetServiceDB(serviceName string) (db *redis.Client) {
	var (
		err error
	)
	serviceInfo, _ := ApiServices.Get(serviceName)
	DataSource := serviceInfo.DataSource
	if db, err = config.GetRdsClientByName(DataSource); err != nil {
		log.Panic().Str("DataSource not defined in enviroment. Please check the configuration", DataSource).Send()
	}
	return db
}

var fun2ApiInfoMap = &sync.Map{}
var APIGroupByDataSource = cmap.New[[]string]()
