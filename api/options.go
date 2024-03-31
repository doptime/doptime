package api

import (
	"context"
	"net/http"
	"reflect"
)

// ApiOption is parameter to create an API, RPC, or CallAt
type Api[i any, o any] struct {
	Name       string
	DataSource string
	WithHeader bool
	WithJwt    bool
	IsRpc      bool
	Ctx        context.Context
	// ApiFuncWithMsgpackedParam is the function of the service
	ApiFuncWithMsgpackedParam func(s []byte) (ret interface{}, err error)
	F                         func(InParameter i) (ret o, err error)
	Validate                  func(pIn interface{}) error
}
type ApiInterface interface {
	GetName() string
	GetDataSource() string
	GetWithJwt() bool
	ProcessOneMap(_map map[string]interface{}) (ret interface{}, err error)
	MergeHeader(req *http.Request, paramIn map[string]interface{})
}

func (a *Api[i, o]) GetName() string {
	return a.Name
}
func (a *Api[i, o]) MergeHeader(req *http.Request, paramIn map[string]interface{}) {
	if !a.WithHeader {
		return
	}
	//copy fields from req to paramIn
	for key, value := range req.Header {
		if len(value) > 1 {
			paramIn["Header"+key] = value
		} else {
			paramIn["Header"+key] = value[0]
		}
	}
	// copy ip address from req to paramIn
	paramIn["Header"+"RemoteAddr"] = req.RemoteAddr
	paramIn["Header"+"Host"] = req.Host
	paramIn["Header"+"Method"] = req.Method
	paramIn["Header"+"Path"] = req.URL.Path
	paramIn["Header"+"RawQuery"] = req.URL.RawQuery

}
func (a *Api[i, o]) GetDataSource() string {
	return a.DataSource
}
func (a *Api[i, o]) GetWithJwt() bool {
	return a.WithJwt
}
func (a *Api[i, o]) ProcessOneMap(_map map[string]interface{}) (ret interface{}, err error) {
	var (
		in  i
		pIn interface{}
		//datapack DataPacked
	)
	// case double pointer decoding
	if vType := reflect.TypeOf((*i)(nil)).Elem(); vType.Kind() == reflect.Ptr {
		pIn = reflect.New(vType.Elem()).Interface()
		in = pIn.(i)
	} else {
		pIn = reflect.New(vType).Interface()
		in = *pIn.(*i)
	}

	if decoder, errMapTostruct := mapToStructDecoder(pIn); errMapTostruct != nil {
		return nil, errMapTostruct
	} else if err = decoder.Decode(_map); err != nil {
		return nil, err
	}
	//validate the input if it is struct and has tag "validate"
	if err = a.Validate(pIn); err != nil {
		return nil, err
	}
	return a.F(in)
}

// // Key purpose of ApiNamed is to allow different API to have the same input type
// func (rpc *Api[i, o]) ReName(apiName string) (out *Api[i, o]) {
// 	//remove the old name from the concurrent map
// 	ApiServices.Remove(rpc.Name)
// 	rpc.Name = specification.ApiName(apiName)
// 	//register the new name
// 	ApiServices.Set(rpc.Name, rpc)

// 	//if that is RPC, update the fun2Api map

// 	return rpc
// }

// func (rpc *Api[i, o]) ReDataSource(DataSource string) (out *Api[i, o]) {
// 	//warn if DataSource not defined in the environment
// 	//however, the DataSource may be specified later in the environment, by dynamic loading from remove config file
// 	if _, err := config.GetRdsClientByName(DataSource); err != nil {
// 		log.Warn().Str("this data source specified in the api is not defined in the environment. Please check the configuration", DataSource).Send()
// 	}
// 	rpc.DataSource = DataSource

// 	return rpc
// }

// ApiOption is parameter to create an API, RPC, or CallAt
type ApiOption struct {
	Name       string
	DataSource string
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
	return o
}
