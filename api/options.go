package api

// should allow pay to api ProviderAccountEmail ˝
type PublishOptions struct {
	//set to 0 to disable
	RateByCall float64

	//set to 0 to disable
	RateByRequestMB  float64
	RateByResponseMB float64

	//set to 0 to disable
	RateByRequestToken  float64
	RateByResponseToken float64

	ActiveAt      int64
	ProviderToken string
}

// Option is parameter to create an API, RPC, or CallAt
type Option struct {
	ApiSourceRds string
	PublishInfo  *PublishOptions
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
