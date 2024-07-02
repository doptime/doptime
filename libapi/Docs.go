package libapi

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/doptime/doptime/api"
	"github.com/doptime/doptime/rdsdb"
)

type DocsIn struct {
	// type of "api" or "data"
	t string
}

var ApiDocs = api.Api(func(req *DocsIn) (r string, err error) {
	if req.t == "api" {
		return GetApiDocs()
	} else if req.t == "data" {
		return GetDataDocs()
	}
	return "you should specify a type in your url '?t=api' or '?t=data'", nil
}).Func

var keyApiDataDocs = rdsdb.HashKey[string, *api.DocsOfApi]()

func GetApiDocs() (string, error) {
	result, err := keyApiDataDocs.HGetAll()
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
		apiName := strings.ReplaceAll(v.KeyName, "api:", "") + " "
		apiNameFirstCharUpper := strings.ToUpper(apiName) + apiName[1:]
		jsParamIn, _ := json.Marshal(v.ParamIn)
		jsParamOut, _ := json.Marshal(v.ParamOut)
		//var apiGetProjectArchitectureInfo = newApi("getProjectArchitectureInfo", { "ProjectDir": "", "SkipFiles": [], "SkipDirs": [] },)
		ret.WriteString("var api" + apiNameFirstCharUpper + " = newApi( paramIn=" + string(jsParamIn) + ", paramOut=" + string(jsParamOut) + ")")
		ret.WriteString("\n\n")
	}
	// convert to toml string, do
	return ret.String(), nil
}

func GetDataDocs() (string, error) {
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
