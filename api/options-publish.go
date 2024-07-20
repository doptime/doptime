package api

type TokenBasedPricing struct {
	RatePerRequestToken  float64
	RatePerResponseToken float64
}
type TrafficBasedPricing struct {
	RatePerRequestMB  float64
	RatePerResponseMB float64
}
type CallBasedPricing struct {
	RatePerCall float64
}

// should allow pay to api ProviderAccountEmail ˝
type PublishOptions struct {
	PricingByCall  *CallBasedPricing
	PricingByToken *TokenBasedPricing
	PricingByKB    *TrafficBasedPricing
	ActiveAt       int64
	ProviderToken  string
}
