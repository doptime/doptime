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
	Ctx       context.Context
	Params    map[string]interface{}
	JwtClaims jwt.MapClaims
	// case get
	Cmd   string
	Key   string
	Field string

	RedisDataSource string
	RdsClient       *redis.Client

	ReqID string
}

func NewHttpContext(ctx context.Context, r *http.Request, w http.ResponseWriter) (svc *DoptimeReqCtx, err error, httpStatus int) {
	var (
		CmdKeyFields                           []string
		CmdKeyFieldsStr, pathStr, pathLastPart string
		ok                                     bool
	)
	svc = &DoptimeReqCtx{Ctx: ctx, JwtClaims: jwt.MapClaims{}}
	// field is required for certain data cmds
	svc.Field = r.FormValue("f")
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
	needed, ok = DataCmdRequireField[svc.Cmd]
	if ok && needed && svc.Field == "" {
		return svc, errors.New("url  field required"), http.StatusBadRequest
	}

	//load redis datasource value from form
	svc.RedisDataSource = lib.Ternary(r.FormValue("ds") == "", "default", r.FormValue("ds"))
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

	svc.BuildParamIn(r)
	return svc, nil, http.StatusOK
}
func (svc *DoptimeReqCtx) BuildParamIn(r *http.Request) {

	svc.Params = lib.Ternary(svc.Params == nil, map[string]interface{}{}, svc.Params)

	paramIn, err := io.ReadAll(r.Body)

	//merge body param
	if contentType := r.Header.Get("Content-Type"); len(paramIn) > 0 && len(contentType) > 0 && err == nil {
		switch contentType {
		case "application/octet-stream":
			err = msgpack.Unmarshal(paramIn, &svc.Params)
			if err != nil {
				var interfaceIn interface{}
				if err = msgpack.Unmarshal(paramIn, &interfaceIn); err == nil {
					svc.Params["_msgpack-nonstruct"] = paramIn
				}
			}
		case "application/json":
			err = json.Unmarshal(paramIn, &svc.Params)
			if err != nil {
				var interfaceIn interface{}
				if err = json.Unmarshal(paramIn, &interfaceIn); err == nil {
					svc.Params["_jsonpack-nonstruct"] = paramIn
				}
			}
		}
	}
	//MergeFormParam
	r.ParseForm()
	for key, value := range r.Form {
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
	for k := range svc.Params {
		if strings.HasPrefix(k, "@") {
			delete(svc.Params, k)
		}
	}

	// copy request info
	svc.Params["@RemoteAddr"] = r.RemoteAddr
	svc.Params["@Host"] = r.Host
	svc.Params["@Method"] = r.Method
	svc.Params["@Path"] = r.URL.Path
	svc.Params["@RawQuery"] = r.URL.RawQuery

	//add all Jwt fields to paramIn
	for k, v := range svc.JwtClaims {
		//convert first letter of k to upper case
		k = strings.ToUpper(k[:1]) + k[1:]
		svc.Params["@"+k] = v
	}
	//add key and field to paramIn
	svc.Params["@Key"] = svc.Key
	svc.Params["@Field"] = svc.Field
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
