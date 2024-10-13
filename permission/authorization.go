package permission

import (
	"time"

	"github.com/doptime/config/cfghttp"
	"github.com/doptime/doptime/dlog"
	"github.com/doptime/redisdb"
	cmap "github.com/orcaman/concurrent-map/v2"
)

var rdsPermit = redisdb.HashKey[string, string](redisdb.WithKey("_permissions"))
var permitmap cmap.ConcurrentMap[string, bool] = cmap.New[bool]()

// this version of IsPermitted is design for fast searching & modifying
func IsPermitted(operation string) (ok bool) {
	var (
		autoPermit                            bool   = cfghttp.AutoDataAuth
		permitKey                             string = operation
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
		_permitmap cmap.ConcurrentMap[string, bool] = cmap.New[bool]()
	)
	keys, err := rdsPermit.HKeys()
	// show log if it is the first time to load
	for ; !ConfigurationLoaded; ConfigurationLoaded = true {
		if err != nil {
			dlog.Warn().AnErr("Step2.1: start permission loading from redis failed", err).Send()
		} else {
			dlog.Info().Msg("Step2.2: start permission loaded from redis")
		}
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
