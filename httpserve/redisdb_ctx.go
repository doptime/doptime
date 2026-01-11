package httpserve

import (
	"reflect"

	"github.com/doptime/doptime/utils/mapper"
	"github.com/doptime/redisdb"
	"github.com/vmihailenco/msgpack/v5"
)

func (req *DoptimeReqCtx) ToValue(key redisdb.IHttpKey, _msgpack []byte) (interface{}, error) {
	val := key.GetValue()

	var dest interface{} = val
	if !isPtr(val) {
		dest = &val
	}

	if len(_msgpack) == 0 {
		err := mapper.Decode(req.Params, dest)
		return val, err
	}

	tempMap := make(map[string]interface{})
	if err := msgpack.Unmarshal(_msgpack, &tempMap); err == nil {
		req.removeSuspiciousAtParam(tempMap)
		for k, v := range req.Params {
			tempMap[k] = v
		}
		err = mapper.Decode(tempMap, dest)
		return val, err
	}

	err := msgpack.Unmarshal(_msgpack, dest)
	return val, err
}

func (req *DoptimeReqCtx) ToValues(key redisdb.IHttpKey, dataList []string) ([]interface{}, error) {
	values := make([]interface{}, 0, len(dataList))

	for i, data := range dataList {
		if i < len(req.Fields) {
			req.Params["@field"] = req.Fields[i]
		}

		val, err := req.ToValue(key, []byte(data))
		if err != nil {
			return nil, err
		}
		values = append(values, val)
	}

	return values, nil
}

func isPtr(v interface{}) bool {
	if v == nil {
		return false
	}
	return reflect.ValueOf(v).Kind() == reflect.Ptr
}
