package libapi

import (
	"github.com/doptime/doptime/api"
	"github.com/doptime/doptime/apiinfo"
)

type DocsIn struct {
	// type of "api" or "data"
	T string `msgpack:"t"`
}

var ApiDocs = api.Api(func(req *DocsIn) (r string, err error) {
	if req.T == "api" {
		return apiinfo.GetApiDocs()
	} else if req.T == "data" {
		return apiinfo.GetDataDocs()
	}
	return "you should specify a type in your url '?t=api' or '?t=data'", nil
}).Func
