package libapi

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/doptime/doptime/api"
	"github.com/doptime/doptime/rdsdb"
)

type ApiDocsIn struct {
}

var keyApiDataDocs = rdsdb.HashKey[string, *api.ApiDocs]()

var ApiApiDocs = api.Api(func(req *ApiDocsIn) (r string, err error) {
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
}).Func
