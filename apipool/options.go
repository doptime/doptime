package apipool

import "github.com/doptime/doptime/apiinfo"

type OptSetter func(o *apiinfo.PublishSetting)

func WithRateByCall(rate float64) OptSetter {
	return func(o *apiinfo.PublishSetting) {
		o.RateByCall = rate
	}
}
func WithRateByMB(RateByRequestMB, RateByResponseMB float64) OptSetter {
	return func(o *apiinfo.PublishSetting) {
		o.RateByRequestMB = RateByRequestMB
		o.RateByResponseMB = RateByResponseMB
	}
}
func WithRateByToken(RateByRequestToken, RateByResponseToken float64) OptSetter {
	return func(o *apiinfo.PublishSetting) {
		o.RateByRequestToken = RateByRequestToken
		o.RateByResponseToken = RateByResponseToken
	}
}

func WithApiUrl(url string) OptSetter {
	return func(o *apiinfo.PublishSetting) {
		o.ApiUrl = url
	}
}
func WithProviderToken(ProviderToken string) OptSetter {
	return func(o *apiinfo.PublishSetting) {
		o.ProviderToken = ProviderToken
	}
}
