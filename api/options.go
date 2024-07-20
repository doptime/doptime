package api

// Option is parameter to create an API, RPC, or CallAt
type Option struct {
	ApiSourceRds  string
	ApiSourceHttp string
	PublishInfo   *PublishOptions
}

func mergeNewOptions(o *Option, options ...*Option) (out *Option) {
	if len(options) == 0 {
		return o
	}
	for _, option := range options {
		if len(option.ApiSourceRds) > 0 {
			o.ApiSourceRds = option.ApiSourceRds
		}
		if len(option.ApiSourceHttp) > 0 {
			o.ApiSourceHttp = option.ApiSourceHttp
		}
		if option.PublishInfo != nil {
			o.PublishInfo = option.PublishInfo
		}
	}
	return o
}
