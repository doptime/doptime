package rpc

import (
	"net/http"
	"reflect"

	"github.com/doptime/doptime/utils"
)

func (a *Context[i, o]) GetName() string {
	return a.Name
}
func (a *Context[i, o]) MergeHeaderParam(req *http.Request, paramIn map[string]interface{}) {
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
func (a *Context[i, o]) GetDataSource() string {
	return a.ApiSourceRds
}

func (a *Context[i, o]) CallByMap(_map map[string]interface{}) (ret interface{}, err error) {
	var (
		in          i
		pIn         interface{}
		isTypeInPtr bool = false
		//datapack DataPacked
	)
	// case double pointer decoding
	if vType := reflect.TypeOf((*i)(nil)).Elem(); vType.Kind() == reflect.Ptr {
		pIn = reflect.New(vType.Elem()).Interface()
		in = pIn.(i)
		isTypeInPtr = true
	} else {
		pIn = reflect.New(vType).Interface()
		in = *pIn.(*i)
	}

	if decoder, errMapTostruct := utils.MapToStructDecoder(pIn); errMapTostruct != nil {
		return nil, errMapTostruct
	} else if err = decoder.Decode(_map); err != nil {
		return nil, err
	}
	//load fill the left fields from db
	if a.ParamEnhancer != nil {
		if out, err := a.ParamEnhancer(in); err != nil {
		} else if isTypeInPtr {
			pIn = out
		} else {
			*pIn.(*i) = out
		}
	}

	//validate the input if it is struct and has tag "validate"
	if err = a.Validate(pIn); err != nil {
		return nil, err
	}
	//post save the result to db
	ret, err = a.Func(in)
	if a.ResultSaver != nil && err == nil {
		_ = a.ResultSaver(in, ret.(o))
	}
	//modify the result value to the web client.
	if a.ResponseModifier != nil {
		ret, err = a.ResponseModifier(in, ret.(o))
	}
	return ret, err
}
