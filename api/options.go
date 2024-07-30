package api

import "github.com/doptime/doptime/apiinfo"

// Option is parameter to create an API, RPC, or CallAt
type Option struct {
	ApiSourceRds string
	PublishInfo  *apiinfo.PublishSetting
}

func mergeNewOptions(o *Option, options ...*Option) (out *Option) {
	if len(options) == 0 {
		return o
	}
	for _, option := range options {
		if len(option.ApiSourceRds) > 0 {
			o.ApiSourceRds = option.ApiSourceRds
		}
		if option.PublishInfo != nil {
			o.PublishInfo = option.PublishInfo
		}
	}
	return o
}
