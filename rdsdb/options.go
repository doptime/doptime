// package do stands for data options
package rdsdb

// Option is parameter to create an API, RPC, or CallAt
type Option struct {
	Key             string
	KeyType         string
	DataSource      string
	RegisterWebData bool
	Modifiers       map[string]ModifierFunc
}
type opSetter func(*Option)

func WithKey(key string) opSetter {
	return func(o *Option) {
		o.Key = key
	}
}

func WithRds(dataSource string) opSetter {
	return func(o *Option) {
		o.DataSource = dataSource
	}
}

func WithRegisterWebData(registerWebData bool) opSetter {
	return func(o *Option) {
		o.RegisterWebData = registerWebData
	}
}
func WithModifier(extraModifiers map[string]ModifierFunc) opSetter {
	return func(o *Option) {
		o.Modifiers = extraModifiers
	}
}

func (opt Option) applyOptions(options ...opSetter) *Option {
	for _, option := range options {
		option(&opt)
	}
	return &opt
}
