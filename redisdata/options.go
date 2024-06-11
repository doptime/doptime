// package do stands for data options
package redisdata

// DataOption is parameter to create an API, RPC, or CallAt
type DataOption struct {
	Key        string
	DataSource string
}

var Option *DataOption = nil

// WithKey purpose of ApiNamed is to allow different API to have the same input type
func (o *DataOption) WithKey(key string) (out *DataOption) {
	if out = o; o == Option {
		out = &DataOption{DataSource: "default"}
	}
	out.Key = key
	return out
}
func (o *DataOption) WithRds(dataSource string) (out *DataOption) {
	if out = o; o == Option {
		out = &DataOption{DataSource: "default"}
	}
	out.DataSource = dataSource
	return out
}
