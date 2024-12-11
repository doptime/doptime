package api

import "github.com/doptime/logger"

func init() {
	logger.Info().Msg("Receive Rpc started..")
	go rpcCallAtTasksLoad()
	go rpcReceive()
}
