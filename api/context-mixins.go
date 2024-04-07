package api

import (
	"strings"
)

// this is to be used in MixinParamEnhancer, to quickly fix key & value
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
func (ctx *Context[i, o]) MixinParamEnhancer(paramEnhancer func(paramMap map[string]interface{}, param i) (out i, err error)) *Context[i, o] {
	ctx.ParamEnhancer = paramEnhancer
	return ctx
}

// MixinResultSaver is a mixin function to save result to db
// The resultSaver is a function used to save the result excuted by the service. & response the value to the web client. (hide the fields if you need)
func (ctx *Context[i, o]) MixinResultSaver(resultSaver func(param i, ret o, paramMap map[string]interface{}) (err error)) *Context[i, o] {
	ctx.ResultSaver = resultSaver
	return ctx
}

// MixinResponseModifier is a mixin function to modify the response  value to the web client.
func (ctx *Context[i, o]) MixinResponseModifier(ResponseModifier func(param i, ret o, paramMap map[string]interface{}) (valueToWebclient interface{}, err error)) *Context[i, o] {
	ctx.ResponseModifier = ResponseModifier
	return ctx
}
