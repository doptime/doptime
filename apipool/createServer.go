package apipool

import (
	cmap "github.com/orcaman/concurrent-map/v2"
)

func CreateServer(url string, apiName string, callback func(bytes []byte) (ret []byte, err error)) {
	wsCallback := func(apiMessege *ApiContext) {
		ret, err := callback(apiMessege.Req.ParamIn)
		apiMessege.Resp.Result = ret
		if err != nil {
			apiMessege.Resp.Error = err.Error()
		}
	}
	ServerApis.Set(url+":"+apiName, wsCallback)
}

// key the url of the WebSocket connection
// value the callback function for messages
var ServerApis = cmap.New[func(apiMessege *ApiContext)]()
