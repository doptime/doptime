package api

import "github.com/doptime/doptime/data"

// ApiOption is parameter to create an API, RPC, or CallAt
type ApiOption struct {
	Name            string
	DataSource      string
	LoadParamFromDB []func(_mp map[string]interface{}) error
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

func (o *ApiOption) WithDataSource(DataSource string) (out *ApiOption) {
	if out = o; o == Option {
		out = &ApiOption{}
	}
	out.DataSource = DataSource
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
	if len(newOption.DataSource) > 0 {
		o.DataSource = newOption.DataSource
	}
	if len(newOption.LoadParamFromDB) > 0 {
		o.LoadParamFromDB = newOption.LoadParamFromDB
	}
	return o
}

func (o *ApiOption) WithParamHGet(key string, field string) (out *ApiOption) {
	if out = o; o == Option {
		out = &ApiOption{}
	}
	funLoader := func(_mp map[string]interface{}) (err error) {
		var k, f string = replaceKeyValueWithJwt(key, field, _mp)
		do := data.New[string, map[string]interface{}](data.Option.WithKey(k).WithDataSource(o.DataSource))
		out, err := do.HGet(f)
		if out == nil || err != nil {
			return err
		}
		DataMerger(_mp, out)
		return nil
	}
	out.LoadParamFromDB = append(o.LoadParamFromDB, funLoader)
	return out
}

func (o *ApiOption) WithParamGet(key string, field string) (out *ApiOption) {
	if out = o; o == Option {
		out = &ApiOption{}
	}
	funLoader := func(_mp map[string]interface{}) (err error) {
		var k, f string = replaceKeyValueWithJwt(key, field, _mp)
		do := data.New[string, map[string]interface{}](data.Option.WithKey(k).WithDataSource(o.DataSource))
		out, err := do.Get(f)
		if out == nil || err != nil {
			return err
		}
		DataMerger(_mp, out)
		return nil
	}
	out.LoadParamFromDB = append(o.LoadParamFromDB, funLoader)
	return out
}
