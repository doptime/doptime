package apipool

import (
	"reflect"

	"github.com/doptime/doptime/apiinfo"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/vmihailenco/msgpack/v5"
)

func APIShare[i any, o any](f func(InParameter i) (ret o, err error), apiName string, optSetter ...apiinfo.OptSetter) func(InParameter i) (ret o, err error) {
	iType := reflect.TypeOf((*i)(nil)).Elem()
	oType := reflect.TypeOf((*o)(nil)).Elem()
	websocketCallback := func(apiMessege *ApiContext) {
		if apiMessege.Req == nil || apiMessege.Req.ParamIn == nil {
			apiMessege.Resp.Error = "no input parameter"
			return
		}
		var req i
		var err error
		if iType.Kind() == reflect.Ptr {
			err = msgpack.Unmarshal(apiMessege.Req.ParamIn, req)
			if err != nil {
				apiMessege.Resp.Error = err.Error()
				return
			}
		} else {
			err = msgpack.Unmarshal(apiMessege.Req.ParamIn, &req)
			if err != nil {
				apiMessege.Resp.Error = err.Error()
				return
			}
		}
		o, err := f(req)
		if err != nil {
			apiMessege.Resp.Error = err.Error()
			return
		}
		apiMessege.Resp.Result, err = msgpack.Marshal(o)
		if err != nil {
			apiMessege.Resp.Error = err.Error()
			return
		}
	}
	var setting = &apiinfo.PublishSetting{ApiUrl: "https://api.doptime.com"}
	for _, setter := range optSetter {
		setter(setting)
	}
	apiinfo.RegisterApi(apiName, iType, oType, setting)
	ServerApis.Set(setting.ApiUrl+":"+apiName, websocketCallback)
	return f
}

// key the url of the WebSocket connection
// value the callback function for messages
var ServerApis = cmap.New[func(apiMessege *ApiContext)]()
