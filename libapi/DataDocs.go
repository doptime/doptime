package libapi

import (
	"encoding/json"
	"strings"

	"github.com/doptime/doptime/api"
	"github.com/doptime/doptime/rdsdb"
)

type DataDocsIn struct {
}

var ApiDataDocs = api.Api(func(req *DataDocsIn) (ret map[string]string, err error) {
	result, err := rdsdb.KeyWebDataDocs.HGetAll()
	if err != nil {
		return nil, err
	}
	ret = make(map[string]string)
	for k, v := range result {
		keyWithFirstCharUpper := strings.ToUpper(v.KeyName[0:1]) + v.KeyName[1:]
		keyWithFirstCharUpper = strings.Split(keyWithFirstCharUpper, ":")[0]
		jsBytes, _ := json.Marshal(v)
		if v.KeyType == "hash" {
			ret[k] = "var key" + keyWithFirstCharUpper + " = new hashKey(" + string(jsBytes) + ")"
		} else if v.KeyType == "string" {
			ret[k] = "var key" + keyWithFirstCharUpper + " = new stringKey(" + string(jsBytes) + ")"
		} else if v.KeyType == "list" {
			ret[k] = "var key" + keyWithFirstCharUpper + " = new listKey(" + string(jsBytes) + ")"
		} else if v.KeyType == "set" {
			ret[k] = "var key" + keyWithFirstCharUpper + " = new setKey(" + string(jsBytes) + ")"
		} else if v.KeyType == "zset" {
			ret[k] = "var key" + keyWithFirstCharUpper + " = new zsetKey(" + string(jsBytes) + ")"
		} else if v.KeyType == "stream" {
			ret[k] = "var key" + keyWithFirstCharUpper + " = new streamKey(" + string(jsBytes) + ")"
		}
	}
	return ret, nil
}).Func
