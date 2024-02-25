package api

// Options is parameter to create an API, RPC, or CallAt
type Options struct {
	ApiName        string
	DataSourceName string
}

// set a option property
type With func(*Options)

// Key purpose of ApiNamed is to allow different API to have the same input type
func WithName(name string) With {
	return func(opts *Options) {
		opts.ApiName = name
	}
}
func WithDS(DataSourceName string) With {
	return func(opts *Options) {
		opts.DataSourceName = DataSourceName
	}
}
func mergeOptions(options ...With) (o *Options) {
	o = &Options{}
	for _, option := range options {
		option(o)
	}
	return o
}