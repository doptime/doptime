package httpserve

import (
	"encoding/json"
	"fmt"

	"github.com/doptime/doptime/api"
	"github.com/vmihailenco/msgpack/v5"
)

// this is used to support http based  api
func (svcCtx *HttpContext) APiHandler() (ret interface{}, err error) {

	var (
		paramIn     map[string]interface{} = map[string]interface{}{}
		ServiceName string                 = svcCtx.Key
		_api        api.ApiInterface
		ok          bool
	)
	//convert query fields to JsonPack. but ignore K field(api name )
	if svcCtx.Req.ParseForm(); len(svcCtx.Req.Form) > 0 {
		for key, value := range svcCtx.Req.Form {
			if paramIn[key] = value[0]; len(value) > 1 {
				paramIn[key] = value // Assign the single value directly
			}
		}
	}
	if msgPack := svcCtx.MsgpackBodyBytes(); len(msgPack) > 0 {
		if err = msgpack.Unmarshal(msgPack, &paramIn); err != nil {
			return nil, fmt.Errorf("msgpack.Unmarshal msgPack error %s", err)
		}
		//paramIn["MsgpackBody"] = msgPack
	} else if jsonBody := svcCtx.JsonBodyBytes(); len(jsonBody) > 0 {
		//convert to msgpack, so that fields can be renamed in ProcessOneJob
		if err = json.Unmarshal(jsonBody, &paramIn); err != nil {
			return nil, fmt.Errorf("msgpack.Unmarshal JsonBody error %s", err)
		}
	}

	if _api, ok = api.GetApiByName(ServiceName); !ok {
		return nil, fmt.Errorf("err no such api")
	}
	_api.MergeHeader(svcCtx.Req, paramIn)
	svcCtx.MergeJwtField(paramIn)
	return _api.CallByMap(paramIn)

}
