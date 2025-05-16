package httpdoc

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
	result, err := redisdb.KeyWebDataSchema.HGetAll()
	if err != nil {
		return "", err
	}
	var ret strings.Builder
	ret.WriteString("import { hashKey, stringKey, listKey, setKey, zsetKey, streamKey } from \"doptime-client\"\n")
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
		ret.WriteString(v.TSInterface + "\n")
		if v.KeyType == "hash" {
			ret.WriteString("export const key" + keyWithFirstCharUpper + " = new hashKey<" + v.ValueTypeName + ">(\"" + k + "\")")
		} else if v.KeyType == "string" {
			ret.WriteString("export const key" + keyWithFirstCharUpper + " = new stringKey<" + v.ValueTypeName + ">(\"" + k + "\")")
		} else if v.KeyType == "list" {
			ret.WriteString("export const key" + keyWithFirstCharUpper + " = new listKey<" + v.ValueTypeName + ">(\"" + k + "\")")
		} else if v.KeyType == "set" {
			ret.WriteString("export const key" + keyWithFirstCharUpper + " = new setKey<" + v.ValueTypeName + ">(\"" + k + "\")")
		} else if v.KeyType == "zset" {
			ret.WriteString("export const key" + keyWithFirstCharUpper + " = new zsetKey<" + v.ValueTypeName + ">(\"" + k + "\")")
		} else if v.KeyType == "stream" {
			ret.WriteString("export const key" + keyWithFirstCharUpper + " = new streamKey<" + v.ValueTypeName + ">(\"" + k + "\")")
		}
		ret.WriteString("\n\n")
	}
	// convert to toml string, do
	return ret.String(), nil
}
