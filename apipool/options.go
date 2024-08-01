package apipool

import "github.com/doptime/doptime/apiinfo"

func WithRateByCall(rate float64) apiinfo.OptSetter {
	return func(o *apiinfo.PublishSetting) {
		o.RateByCall = rate
	}
}
func WithRateByMB(RateByRequestMB, RateByResponseMB float64) apiinfo.OptSetter {
	return func(o *apiinfo.PublishSetting) {
		o.RateByRequestMB = RateByRequestMB
		o.RateByResponseMB = RateByResponseMB
	}
}
func WithRateByToken(RateByRequestToken, RateByResponseToken float64) apiinfo.OptSetter {
	return func(o *apiinfo.PublishSetting) {
		o.RateByRequestToken = RateByRequestToken
		o.RateByResponseToken = RateByResponseToken
	}
}

func WithApiUrl(url string) apiinfo.OptSetter {
	return func(o *apiinfo.PublishSetting) {
		o.ApiUrl = url
	}
}

func WithProviderAPIToken(token string) apiinfo.OptSetter {
	return func(o *apiinfo.PublishSetting) {
		o.ProviderAPIToken = token
	}
}
