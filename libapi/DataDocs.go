package libapi

import (
	"encoding/json"
	"strings"

	"github.com/doptime/doptime/api"
	"github.com/doptime/doptime/rdsdb"
)

type DataDocsIn struct {
}

var ApiDataDocs = api.Api(func(req *DataDocsIn) (r string, err error) {
	result, err := rdsdb.KeyWebDataDocs.HGetAll()
	if err != nil {
		return "", err
	}
	var ret strings.Builder
	for k, v := range result {
		keyWithFirstCharUpper := strings.ToUpper(v.KeyName[0:1]) + v.KeyName[1:]
		keyWithFirstCharUpper = strings.Split(keyWithFirstCharUpper, ":")[0]
		ret.WriteString("\n")
		ret.WriteString("keyname: " + k + "\n")
		ret.WriteString("keyType: " + v.KeyType + "\n")
		jsBytes, _ := json.Marshal(v.Instance)
		if v.KeyType == "hash" {
			ret.WriteString("var key" + keyWithFirstCharUpper + " = new hashKey(\"" + k + "\"" + string(jsBytes) + ")")
		} else if v.KeyType == "string" {
			ret.WriteString("var key" + keyWithFirstCharUpper + " = new stringKey(\"" + k + "\"" + string(jsBytes) + ")")
		} else if v.KeyType == "list" {
			ret.WriteString("var key" + keyWithFirstCharUpper + " = new listKey(\"" + k + "\"" + string(jsBytes) + ")")
		} else if v.KeyType == "set" {
			ret.WriteString("var key" + keyWithFirstCharUpper + " = new setKey(\"" + k + "\"" + string(jsBytes) + ")")
		} else if v.KeyType == "zset" {
			ret.WriteString("var key" + keyWithFirstCharUpper + " = new zsetKey(\"" + k + "\"" + string(jsBytes) + ")")
		} else if v.KeyType == "stream" {
			ret.WriteString("var key" + keyWithFirstCharUpper + " = new streamKey(\"" + k + "\"" + string(jsBytes) + ")")
		}
	}
	// convert to toml string, do
	return ret.String(), nil
}).Func
