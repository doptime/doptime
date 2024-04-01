package api

import (
	"net/http"
	"reflect"
)

type ApiInterface interface {
	GetName() string
	GetDataSource() string
	GetWithJwt() bool
	CallByMap(_map map[string]interface{}) (ret interface{}, err error)
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

func (a *Api[i, o]) CallByMap(_map map[string]interface{}) (ret interface{}, err error) {
	var (
		in  i
		pIn interface{}
		//datapack DataPacked
	)
	//load data from db
	for _, fun := range a.LoadParamFromDB {
		fun(_map)
	}

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
