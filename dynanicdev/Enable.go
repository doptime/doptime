package dynamicdev

import "github.com/doptime/doptime/dlog"

func Enable() {
	dlog.Debug().Msg("dynamicdev enabled")
}
func init() {
	Enable()

}
