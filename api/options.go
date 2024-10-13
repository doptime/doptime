package api

// Option is parameter to create an API, RPC, or CallAt
type Option struct {
	ApiSourceRds  string
	ApiSourceHttp string
	ApiKey        string
}
type optionSetter func(*Option)

func WithApiRds(rdsSource string) optionSetter {
	return func(o *Option) {
		o.ApiSourceRds = rdsSource
	}
}
func WithApiHttp(httpSource string) optionSetter {
	return func(o *Option) {
		o.ApiSourceHttp = httpSource
	}
}
func WithApiDoptime(apiKey string) optionSetter {
	return func(o *Option) {
		o.ApiSourceHttp = "https://api.doptime.com/"
		o.ApiKey = apiKey
	}
}

func WithApiKey(apiKey string) optionSetter {
	return func(o *Option) {
		o.ApiKey = apiKey
	}
}

func (o Option) mergeNewOptions(optionSetters ...optionSetter) (out *Option) {
	for _, setter := range optionSetters {
		setter(&o)
	}
	return &o
}
