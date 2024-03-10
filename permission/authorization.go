package permission

import (
	"time"

	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/rs/zerolog/log"
	"github.com/yangkequn/goflow/config"
	"github.com/yangkequn/goflow/data"
)

var rdsPermit = data.New[string, string](data.Option.WithKey("_permissions"))
var permitmap cmap.ConcurrentMap[string, bool] = cmap.New[bool]()

// this version of IsPermitted is design for fast searching & modifying
func IsPermitted(dataKey string, operation string) (ok bool) {
	var (
		autoPermit                            bool   = config.Cfg.Http.AutoAuth
		permitKey                             string = dataKey + "::" + operation
		permitKeyAllowed, permitKeyDisallowed string = permitKey + "::on", permitKey + "::off"
	)
	//blacklist first
	if _, ok := permitmap.Get(permitKeyDisallowed); ok {
		return false
	}
	if _, ok := permitmap.Get(permitKeyAllowed); ok {
		return true
	}
	if autoPermit {
		permitmap.Set(permitKeyAllowed, true)
		rdsPermit.HSet(permitKeyAllowed, time.Now().Format("2006-01-02 15:04:05"))
	}
	return autoPermit
}

var ConfigurationLoaded bool = false

func LoadPermissionTable() {
	var (
		keys       []string
		err        error
		_permitmap cmap.ConcurrentMap[string, bool] = cmap.New[bool]()
	)

	if keys, err = rdsPermit.HKeys(); !ConfigurationLoaded {
		if err != nil {
			log.Warn().AnErr("Step2.1: start permission loading from redis failed", err).Send()
		} else {
			log.Info().Msg("Step2.2: start permission loaded from redis")
		}
		ConfigurationLoaded = true
	}
	for _, key := range keys {
		_permitmap.Set(key, true)
	}
	permitmap = _permitmap
	go func() {
		time.Sleep(time.Minute)
		LoadPermissionTable()
	}()
}

func init() {
	LoadPermissionTable()
}
