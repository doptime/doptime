package doc

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/doptime/redisdb"
)

func GetApiDocs() (string, error) {
	result, err := KeyApiDataDocs.HGetAll()
	if err != nil {
		return "", err
	}
	var ret strings.Builder
	var now = time.Now().Unix()
	for _, v := range result {
		// if not updated in latest 20min, ignore it
		if v.UpdateAt < now-20*60 {
			continue
		}
		apiName := strings.ReplaceAll(v.KeyName, "api:", "")
		if len(apiName) < 1 {
			continue
		}
		apiNameFirstCharUpper := strings.ToUpper(apiName[0:1]) + apiName[1:]
		jsParamIn, _ := json.Marshal(v.ParamIn)
		jsParamOut, _ := json.Marshal(v.ParamOut)
		//var apiGetProjectArchitectureInfo = newApi("getProjectArchitectureInfo", { "ProjectDir": "", "SkipFiles": [], "SkipDirs": [] },)
		ret.WriteString("var api" + apiNameFirstCharUpper + " = newApi(\"" + apiName + "\", { in: " + string(jsParamIn) + ", out: " + string(jsParamOut) + "})")
		ret.WriteString("\n\n")
	}
	// convert to toml string, do
	return ret.String(), nil
}

func GetDataDocs() (string, error) {
	result, err := redisdb.KeyWebDataDocs.HGetAll()
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
		if len(v.KeyName) < 1 {
			continue
		}
		keyWithFirstCharUpper := strings.ToUpper(v.KeyName[0:1]) + v.KeyName[1:]
		keyWithFirstCharUpper = strings.Split(keyWithFirstCharUpper, ":")[0]
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
}
