package httpserve

import (
	"github.com/doptime/doptime/api"
	"github.com/doptime/doptime/httpserve/httpdoc"
)

type ApiDocs struct {
}

type DataDocs struct {
}

var ApiApiDocs = api.Api(func(req *ApiDocs) (r string, err error) {
	return httpdoc.GetApiDocs()
}).Func

var ApiDataDocs = api.Api(func(req *DataDocs) (r string, err error) {
	return httpdoc.GetDataDocs()
}).Func
