package api

import (
	"strings"
)

// this is to be used in HookParamEnhancer, to quickly fix key & value
func ReplaceTagsInKeyField(key string, field string, paramTable map[string]interface{}) (string, string) {
	for tag, value := range paramTable {
		if vstr, ok := value.(string); ok {
			if strings.Contains(key, "@"+tag) {
				key = strings.Replace(key, "@"+tag, vstr, 1)
			}
			if strings.Contains(field, "@"+tag) {
				field = strings.Replace(field, "@"+tag, vstr, 1)
			}
		}
	}
	return key, field
}

func (ctx *ApiCtx[i, o]) HookParamEnhancer(
	paramEnhancer func(param i) (out i, err error),
) *ApiCtx[i, o] {
	ctx.ParamEnhancer = paramEnhancer
	return ctx
}

// HookResultSaver is a hook function to save result to db
// The resultSaver is a function used to save the result excuted by the service. & response the value to the web client. (hide the fields if you need)
func (ctx *ApiCtx[i, o]) HookResultSaver(
	resultSaver func(param i, ret o) (err error),
) *ApiCtx[i, o] {
	ctx.ResultSaver = resultSaver
	return ctx
}

// HookResponseModifier is a hook function to modify the response  value to the web client.
func (ctx *ApiCtx[i, o]) HookResponseModifier(
	ResponseModifier func(param i, ret o) (valueToWebclient interface{}, err error),
) *ApiCtx[i, o] {
	ctx.ResponseModifier = ResponseModifier
	return ctx
}
