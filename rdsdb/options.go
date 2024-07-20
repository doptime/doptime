// package do stands for data options
package rdsdb

// Option is parameter to create an API, RPC, or CallAt
type Option struct {
	Key             string
	DataSource      string
	RegisterWebData bool
}
