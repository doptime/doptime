package httpapi

import "net/http"

type ApiInterface interface {
	GetName() string
	CallByMap(_map map[string]interface{}) (ret interface{}, err error)
	GetDataSource() string
	MergeHeaderParam(req *http.Request, paramIn map[string]interface{})
}
