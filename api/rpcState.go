package api

import (
	"fmt"
	"time"

	"github.com/doptime/doptime/utils"
	"github.com/doptime/logger"
)

var apiCounter utils.Counter = utils.Counter{}

func reportApiStates() {
	//wait till all apis are loaded
	if ApiServices.Count() == 0 {
		logger.Info().Msg("waiting for apis to load")
	}
	for i, lastCnt := 0, 0; (ApiServices.Count() == 0 || lastCnt != ApiServices.Count()) && i < 100; i++ {
		lastCnt = ApiServices.Count()
		fmt.Print(".")
		time.Sleep(time.Millisecond * 100)
	}

	// all keys of ServiceMap to []string serviceNames
	var serviceNames []string = apiServiceNames()
	logger.Info().Any("cnt", len(serviceNames)).Strs("apis are load:", serviceNames).Send()
	for {
		time.Sleep(time.Second * 60)
		serviceNames = apiServiceNames()
		for _, serviceName := range serviceNames {
			if num, _ := apiCounter.Get(serviceName); num > 0 {
				logger.Info().Any("serviceName", serviceName).Any("proccessed", num).Msg("Tasks processed.")
			}
			apiCounter.DeleteAndGetLastValue(serviceName)
		}
	}
}
func init() {
	go reportApiStates()
	go StarApis()
}

func StarApis() {
	logger.Info().Msg("Step Last: API is starting")
	rpcCallAtTasksLoad()
	rpcReceive()
}
