package apipool

type Option struct {
	ProviderToken string

	//rating options
	//set to 0 to disable
	RateByCall float64

	//set to 0 to disable
	RateByRequestMB  float64
	RateByResponseMB float64

	//set to 0 to disable
	RateByRequestToken  float64
	RateByResponseToken float64
}

// WithRateByCall sets the rate per call.
func WithRateByCall(rate float64) func(*Option) {
	return func(o *Option) {
		o.RateByCall = rate
	}
}

// WithRateByRequestMB sets the rate per MB for requests.
func WithRateByRequestMB(rate float64) func(*Option) {
	return func(o *Option) {
		o.RateByRequestMB = rate
	}
}

// WithRateByResponseMB sets the rate per MB for responses.
func WithRateByResponseMB(rate float64) func(*Option) {
	return func(o *Option) {
		o.RateByResponseMB = rate
	}
}

// WithRateByRequestToken sets the rate per token for requests.
func WithRateByRequestToken(rate float64) func(*Option) {
	return func(o *Option) {
		o.RateByRequestToken = rate
	}
}

// WithRateByResponseToken sets the rate per token for responses.
func WithRateByResponseToken(rate float64) func(*Option) {
	return func(o *Option) {
		o.RateByResponseToken = rate
	}
}

// mergeNewOptions applies a list of option functions to an Option object.
func mergeNewOptions(o *Option, options ...func(*Option)) *Option {
	for _, opt := range options {
		opt(o)
	}
	return o
}
