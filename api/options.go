package api

// ApiOption is parameter to create an API, RPC, or CallAt
type ApiOption struct {
	ApiSourceRds  string
	ApiSourceHttp string
}

var Option *ApiOption

func (o *ApiOption) WithApiRds(ApiDataSourceRds string) (out *ApiOption) {
	if out = o; o == Option {
		out = &ApiOption{}
	}
	out.ApiSourceRds = ApiDataSourceRds
	return out
}
func (o *ApiOption) WithApiHttp(ApiSourceHttp string) (out *ApiOption) {
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
	var newOption *ApiOption = options[0]
	if len(newOption.ApiSourceRds) > 0 {
		o.ApiSourceRds = newOption.ApiSourceRds
	}
	if len(newOption.ApiSourceHttp) > 0 {
		o.ApiSourceHttp = newOption.ApiSourceHttp
	}
	return o
}
