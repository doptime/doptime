package api

// Option is parameter to create an API, RPC, or CallAt
type Option struct {
	ApiSourceRds string
	ApiKey       string
}
type optionSetter func(*Option)

func WithApiRds(rdsSource string) optionSetter {
	return func(o *Option) {
		o.ApiSourceRds = rdsSource
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
