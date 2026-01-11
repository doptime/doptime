package httpserve

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/doptime/config/cfgredis"
	"github.com/doptime/doptime/lib"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"github.com/vmihailenco/msgpack/v5"
)

type DoptimeReqCtx struct {
	Ctx context.Context
	// case get
	Cmd    string
	Key    string
	Fields []string
	// Parmesters
	Queries   url.Values
	Params    map[string]interface{}
	JwtClaims jwt.MapClaims

	RedisDataSource string
	RdsClient       *redis.Client

	ReqID string
}

func (svc *DoptimeReqCtx) Field() string {
	return lib.Ternary(len(svc.Fields) > 0, svc.Fields[0], "")
}

func NewHttpContext(ctx context.Context, r *http.Request, w http.ResponseWriter) (svc *DoptimeReqCtx, err error, httpStatus int) {
	var (
		CmdKeyFields                           []string
		CmdKeyFieldsStr, pathStr, pathLastPart string
		ok                                     bool
	)
	svc = &DoptimeReqCtx{Ctx: ctx, JwtClaims: jwt.MapClaims{}, Params: map[string]interface{}{}}
	//case redis data access
	//i.g. https://url.com/rSvc/HGET-UserAvatar=fa4Y3oyQk2swURaJ?Queries=*&RspType=image/jpeg
	//case api command
	//i.g. https://url.com/ApiDocs
	if pathStr = r.URL.Path; r.URL.RawPath != "" {
		pathStr = r.URL.RawPath
	}
	if pathLastPart = path.Base(pathStr); pathLastPart == "" {
		return nil, errors.New("url missing api_name or data_command"), http.StatusBadRequest
	}
	if CmdKeyFieldsStr, err = url.QueryUnescape(pathLastPart); err != nil {
		return nil, err, http.StatusBadRequest
	}
	//we regard the unknow command or data operation as api command
	if CmdKeyFields = strings.SplitN(CmdKeyFieldsStr, "-", 2); len(CmdKeyFields) == 1 {
		CmdKeyFields = []string{"api", CmdKeyFieldsStr}
	} else if len(CmdKeyFields) == 0 {
		return nil, errors.New("url missing api_name or data_command"), http.StatusBadRequest
	}

	if svc.Cmd = strings.ToUpper(CmdKeyFields[0]); svc.Cmd == "" {
		return nil, errors.New("url missing api_name or data_command"), http.StatusBadRequest
	}

	// cmd and key and field, i.g. /HGET/UserAvatar?F=fa4Y3oyQk2swURaJ
	// both cmd and key are required
	if len(CmdKeyFields) > 1 {
		svc.Key = CmdKeyFields[1]
	}

	// ensure there's a key for certain cmds
	needed, ok := DataCmdRequireKey[svc.Cmd]
	if ok && needed && svc.Key == "" {
		return svc, errors.New("url  key required"), http.StatusBadRequest
	}

	// field is required for certain data cmds
	svc.Queries = r.URL.Query()
	svc.Fields = svc.Queries["f"]
	needed, ok = DataCmdRequireField[svc.Cmd]
	if ok && needed && svc.Field() == "" {
		return svc, errors.New("url  field required"), http.StatusBadRequest
	}

	//load redis datasource value from form
	svc.RedisDataSource = lib.Ternary(svc.Queries.Get("ds") == "", "default", svc.Queries.Get("ds"))
	if svc.RdsClient, ok = cfgredis.Servers.Get(svc.RedisDataSource); !ok {
		return svc, errors.New("redis datasource is unconfigured: " + svc.RedisDataSource), http.StatusBadRequest
	}

	if err = svc.ParseJwtClaim(r); err != nil {
		return svc, err, http.StatusUnauthorized
	}

	//@Tag in key or field should be replaced by value in Jwt
	if err = svc.ReplaceKeyFieldTagWithJwtClaims(); err != nil {
		return nil, err, http.StatusInternalServerError
	}

	svc.BuildParamFromHeaderQueryClaim(r)
	return svc, nil, http.StatusOK
}
func (svc *DoptimeReqCtx) BuildParamFromBody(r *http.Request) (msgpackNonstruct []byte, jsonpackNostruct []byte) {
	var interfaceIn interface{}
	paramIn, err := io.ReadAll(r.Body)

	//merge body param
	if contentType := r.Header.Get("Content-Type"); len(paramIn) > 0 && len(contentType) > 0 && err == nil {
		switch contentType {
		case "application/octet-stream":
			err = msgpack.Unmarshal(paramIn, &svc.Params)
			if err != nil {
				if err = msgpack.Unmarshal(paramIn, &interfaceIn); err == nil {
					msgpackNonstruct = paramIn
				}
			}
		case "application/json":
			err = json.Unmarshal(paramIn, &svc.Params)
			if err != nil {
				if err = json.Unmarshal(paramIn, &interfaceIn); err == nil {
					jsonpackNostruct = paramIn
				}
			}
		}
	}
	return msgpackNonstruct, jsonpackNostruct
}
func (svc *DoptimeReqCtx) removeSuspiciousAtParam(mapParam map[string]interface{}) {
	for k := range mapParam {
		if strings.HasPrefix(k, "@") {
			delete(mapParam, k)
		}
	}
}

func (svc *DoptimeReqCtx) BuildParamFromHeaderQueryClaim(r *http.Request) {

	svc.Params = lib.Ternary(svc.Params == nil, map[string]interface{}{}, svc.Params)

	//MergeQueryParam
	for key, value := range svc.Queries {
		if svc.Params[key] = value[0]; len(value) > 1 {
			svc.Params[key] = value // Assign the single value directly
		}
	}
	//MergeHeaderParam
	for key, value := range r.Header {
		if len(value) > 1 {
			svc.Params[key] = value
		} else {
			svc.Params[key] = value[0]
		}
	}

	//prevent forged jwt field: remove nay field that starts with "@"
	svc.removeSuspiciousAtParam(svc.Params)

	//add key and field to paramIn
	svc.Params["@key"] = svc.Key
	svc.Params["@field"] = svc.Field()
	// copy request info
	svc.Params["@remoteAddr"] = r.RemoteAddr
	svc.Params["@host"] = r.Host
	svc.Params["@method"] = r.Method
	svc.Params["@path"] = r.URL.Path
	svc.Params["@rawQuery"] = r.URL.RawQuery
	//add all Jwt fields to paramIn
	for k, v := range svc.JwtClaims {
		svc.Params["@"+k] = v
	}
}

// Ensure the body is msgpack format
func (svc *DoptimeReqCtx) MsgpackBody(r *http.Request, checkContentType bool) (MsgPack []byte) {
	var err error
	if checkContentType && r.Header.Get("Content-Type") != "application/octet-stream" {
		return nil
	}
	if MsgPack, err = io.ReadAll(r.Body); len(MsgPack) == 0 || err != nil {
		return nil
	}
	return MsgPack
}
