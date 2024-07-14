package api

// ApiOption is parameter to create an API, RPC, or CallAt
type ApiOption struct {
	ApiSourceRds  string
	ApiSourceHttp string
	VendorInfo    *VendorInfo
}

var Option *ApiOption

func (o *ApiOption) WithSourceRds(ApiSourceRds string) (out *ApiOption) {
	if out = o; o == Option {
		out = &ApiOption{}
	}
	out.ApiSourceRds = ApiSourceRds
	return out
}
func (o *ApiOption) WithSourceHttp(ApiSourceHttp string) (out *ApiOption) {
	if out = o; o == Option {
		out = &ApiOption{}
	}
	out.ApiSourceHttp = ApiSourceHttp
	return out
}
func mergeNewOptions(o *ApiOption, options ...*ApiOption) (out *ApiOption) {
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
		if option.VendorInfo != nil {
			o.VendorInfo = option.VendorInfo
		}
	}
	return o
}
