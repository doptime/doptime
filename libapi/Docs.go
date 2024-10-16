package libapi

import (
	"github.com/doptime/doptime/api"
	"github.com/doptime/doptime/apiinfo"
)

type APIDocs struct {
}

type DataDocs struct {
}

var ApiDocsAPI = api.Api(func(req *APIDocs) (r string, err error) {
	return apiinfo.GetApiDocs()
}).Func

var ApiDocs = api.Api(func(req *DataDocs) (r string, err error) {
	return apiinfo.GetDataDocs()
}).Func
