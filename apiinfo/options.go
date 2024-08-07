package apiinfo

// should allow pay to api ProviderAccountEmail ˝
type PublishSetting struct {
	//set to 0 to disable
	RateByCall float64

	//set to 0 to disable
	RateByRequestMB  float64
	RateByResponseMB float64

	//set to 0 to disable
	RateByRequestToken  float64
	RateByResponseToken float64

	ActiveAt int64
	ApiUrl   string

	//JWT will expire, while API Token never
	ProviderAPIToken string
}

type OptSetter func(o *PublishSetting)

func WithRateByCall(rate float64) OptSetter {
	return func(o *PublishSetting) {
		o.RateByCall = rate
	}
}
func WithRateByMB(RateByRequestMB, RateByResponseMB float64) OptSetter {
	return func(o *PublishSetting) {
		o.RateByRequestMB = RateByRequestMB
		o.RateByResponseMB = RateByResponseMB
	}
}
func WithRateByToken(RateByRequestToken, RateByResponseToken float64) OptSetter {
	return func(o *PublishSetting) {
		o.RateByRequestToken = RateByRequestToken
		o.RateByResponseToken = RateByResponseToken
	}
}
func WithApiUrl(url string) OptSetter {
	return func(o *PublishSetting) {
		o.ApiUrl = url
	}
}

func WithProviderAPIToken(token string) OptSetter {
	return func(o *PublishSetting) {
		o.ProviderAPIToken = token
	}
}

// mergeNewOptions applies a list of option functions to an Option object.
func (o *PublishSetting) MergeNewOptions(options ...OptSetter) *PublishSetting {
	for _, opt := range options {
		opt(o)
	}
	return o
}
