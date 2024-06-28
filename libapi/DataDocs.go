package libapi

import (
	"encoding/json"
	"strings"
	"time"

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
	var now = time.Now().Unix()
	for k, v := range result {
		// if not updated in latest 20min, ignore it
		if v.UpdateAt < now-20*60 {
			continue
		}
		keyWithFirstCharUpper := strings.ToUpper(v.KeyName[0:1]) + v.KeyName[1:]
		keyWithFirstCharUpper = strings.Split(keyWithFirstCharUpper, ":")[0]
		ret.WriteString("key: " + v.KeyType + ", " + k + "\n")
		jsBytes, _ := json.Marshal(v.Instance)
		if v.KeyType == "hash" {
			ret.WriteString("var key" + keyWithFirstCharUpper + " = new hashKey(\"" + k + "\", " + string(jsBytes) + ")")
		} else if v.KeyType == "string" {
			ret.WriteString("var key" + keyWithFirstCharUpper + " = new stringKey(\"" + k + "\", " + string(jsBytes) + ")")
		} else if v.KeyType == "list" {
			ret.WriteString("var key" + keyWithFirstCharUpper + " = new listKey(\"" + k + "\", " + string(jsBytes) + ")")
		} else if v.KeyType == "set" {
			ret.WriteString("var key" + keyWithFirstCharUpper + " = new setKey(\"" + k + "\", " + string(jsBytes) + ")")
		} else if v.KeyType == "zset" {
			ret.WriteString("var key" + keyWithFirstCharUpper + " = new zsetKey(\"" + k + "\", " + string(jsBytes) + ")")
		} else if v.KeyType == "stream" {
			ret.WriteString("var key" + keyWithFirstCharUpper + " = new streamKey(\"" + k + "\", " + string(jsBytes) + ")")
		}
		ret.WriteString("\n\n")
	}
	// convert to toml string, do
	return ret.String(), nil
}).Func
