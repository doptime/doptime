package httpapi

import (
	"fmt"
	"time"

	"github.com/doptime/doptime/utils"
	"github.com/doptime/logger"
)

var ApiCounter utils.Counter = utils.Counter{}

func reportApiStates() {
	//wait till all apis are loaded
	if ApiViaHttp.Count() == 0 {
		logger.Info().Msg("waiting for apis to load")
	}
	for i, lastCnt := 0, 0; (ApiViaHttp.Count() == 0 || lastCnt != ApiViaHttp.Count()) && i < 100; i++ {
		lastCnt = ApiViaHttp.Count()
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
			if num, _ := ApiCounter.Get(serviceName); num > 0 {
				logger.Info().Any("serviceName", serviceName).Any("proccessed", num).Msg("Tasks processed.")
			}
			ApiCounter.DeleteAndGetLastValue(serviceName)
		}
	}
}
func init() {
	go reportApiStates()
}
