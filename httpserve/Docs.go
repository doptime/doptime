package httpserve

import (
	"github.com/doptime/doptime/api"
	"github.com/doptime/doptime/httpserve/httpdoc"
)

type ApiDocs struct {
}

type DataDocs struct {
}
type Docs struct {
}

var ApiApiDocs = api.Api(func(req *ApiDocs) (r string, err error) {
	return httpdoc.GetApiDocs()
}).Func

var ApiDataDocs = api.Api(func(req *DataDocs) (r string, err error) {
	return httpdoc.GetDataDocs()
}).Func

var Api_Docs = api.Api(func(req *Docs) (r string, err error) {
	//create link to api docs or data docs
	linkToApiDocs := "<a href=\"/apidocs\">API Docs</a>"
	linkToDataDocs := "<a href=\"/datadocs\">Data Docs</a>"
	return "<html><body>" +
		"<h1>Welcome to Doptime</h1>" +
		"<p>Click here to see the " + linkToApiDocs + "</p>" +
		"<p>Click here to see the " + linkToDataDocs + "</p>" +
		"</body></html>", nil
}).Func
