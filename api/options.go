package api

// ApiOption is parameter to create an API, RPC, or CallAt
type ApiOption struct {
	Name          string
	ApiSourceRds  string
	ApiSourceHttp string
}

var Option *ApiOption

// Key purpose of ApiNamed is to allow different API to have the same input type
func (o *ApiOption) WithName(apiName string) (out *ApiOption) {
	if out = o; o == Option {
		out = &ApiOption{}
	}
	out.Name = apiName
	return out
}

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
	if len(newOption.Name) > 0 {
		o.Name = newOption.Name
	}
	if len(newOption.ApiSourceRds) > 0 {
		o.ApiSourceRds = newOption.ApiSourceRds
	}
	if len(newOption.ApiSourceHttp) > 0 {
		o.ApiSourceHttp = newOption.ApiSourceHttp
	}
	return o
}
