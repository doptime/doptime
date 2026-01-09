package httpserve

import (
	"github.com/doptime/doptime/utils/mapper"
	"github.com/doptime/redisdb"
)

func (req *DoptimeReqCtx) ToValue(key redisdb.IHttpKey, msgpack []byte) (value interface{}, err error) {
	value, err = key.DeserializeValue(msgpack)
	mapper.Decode(req.Params, &value)
	return value, err
}
func (req *DoptimeReqCtx) ToValues(key redisdb.IHttpKey, msgpack []string) (value []interface{}, err error) {
	value, err = key.DeserializeValues(msgpack)
	for i, v := range value {
		//这个地方有错误，fields 的值会不断变化
		req.Params["@field"] = req.Fields[i]
		err = mapper.Decode(req.Params, &v)
		if err != nil {
			return value, err
		}
	}

	return value, err
}
