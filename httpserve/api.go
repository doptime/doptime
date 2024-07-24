package httpserve

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/doptime/doptime/api"
	"github.com/vmihailenco/msgpack/v5"
)

// this is used to support http based  api
func (svcCtx *DoptimeReqCtx) APiHandler(r *http.Request) (ret interface{}, err error) {

	var (
		paramIn     map[string]interface{} = map[string]interface{}{}
		ServiceName string                 = svcCtx.Key
		_api        api.ApiInterface
		ok          bool
	)
	//convert query fields to JsonPack. but ignore K field(api name )
	if r.ParseForm(); len(r.Form) > 0 {
		for key, value := range r.Form {
			if paramIn[key] = value[0]; len(value) > 1 {
				paramIn[key] = value // Assign the single value directly
			}
		}
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if len(bodyBytes) == 0 || err != nil {
		return nil, fmt.Errorf("empty msgpack body")
	}

	contentType := r.Header.Get("Content-Type")
	if contentType == "application/octet-stream" {
		if err = msgpack.Unmarshal(bodyBytes, &paramIn); err != nil {
			return nil, fmt.Errorf("msgpack.Unmarshal msgPack error %s", err)
		}
		//paramIn["MsgpackBody"] = msgPack
	} else if contentType == "application/json" {
		//convert to msgpack, so that fields can be renamed in ProcessOneJob
		if err = json.Unmarshal(bodyBytes, &paramIn); err != nil {
			return nil, fmt.Errorf("msgpack.Unmarshal JsonBody error %s", err)
		}
	}

	if _api, ok = api.GetApiByName(ServiceName); !ok {
		return nil, fmt.Errorf("err no such api")
	}
	_api.MergeHeader(r, paramIn)
	svcCtx.MergeJwtField(r, paramIn)
	return _api.CallByMap(paramIn)

}
