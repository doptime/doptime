package httpserve

import (
	"reflect"

	"github.com/doptime/doptime/utils/mapper"
	"github.com/doptime/redisdb"
)

func (req *DoptimeReqCtx) ToValue(key redisdb.IHttpKey, msgpack []byte) (value interface{}, err error) {
	value, err = key.DeserializeValue(msgpack)
	if err != nil {
		return nil, err
	}

	// Avoid double reference if value is already a pointer
	if isPtr(value) {
		err = mapper.Decode(req.Params, value)
	} else {
		err = mapper.Decode(req.Params, &value)
	}

	return value, err
}

func (req *DoptimeReqCtx) ToValues(key redisdb.IHttpKey, msgpack []string) (value []interface{}, err error) {
	value, err = key.DeserializeValues(msgpack)
	if err != nil {
		return nil, err
	}

	for i := range value {
		if i < len(req.Fields) {
			req.Params["@field"] = req.Fields[i]
		}

		// Access via index to modify the actual element, not the loop copy
		if isPtr(value[i]) {
			err = mapper.Decode(req.Params, value[i])
		} else {
			err = mapper.Decode(req.Params, &value[i])
		}

		if err != nil {
			return value, err
		}
	}

	return value, err
}

func isPtr(v interface{}) bool {
	if v == nil {
		return false
	}
	return reflect.ValueOf(v).Kind() == reflect.Ptr
}
