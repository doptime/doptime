package httpserve

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type DoptimeReqCtx struct {
	Ctx    context.Context
	Claims jwt.MapClaims
	// case get
	Cmd   string
	Key   string
	Field string

	//belongs to api over websocket
	ParamIn []byte
	ReqID   string
}

func NewHttpContext(ctx context.Context, r *http.Request, w http.ResponseWriter) (httpCtx *DoptimeReqCtx, err error) {
	var (
		CmdKeyFields                           []string
		CmdKeyFieldsStr, pathStr, pathLastPart string
	)
	svcContext := &DoptimeReqCtx{Ctx: ctx}
	//case redis data access
	//i.g. https://url.com/rSvc/HGET-UserAvatar=fa4Y3oyQk2swURaJ?Queries=*&RspType=image/jpeg
	//case api command
	//i.g. https://url.com/ApiDocs
	if pathStr = r.URL.Path; r.URL.RawPath != "" {
		pathStr = r.URL.RawPath
	}
	if pathLastPart = path.Base(pathStr); pathLastPart == "" {
		return nil, errors.New("url missing api_name or data_command")
	}
	if CmdKeyFieldsStr, err = url.QueryUnescape(pathLastPart); err != nil {
		return nil, err
	}
	//we regard the unknow command or data operation as api command
	if CmdKeyFields = strings.SplitN(CmdKeyFieldsStr, "-", 2); len(CmdKeyFields) == 1 {
		CmdKeyFields = []string{"api", CmdKeyFieldsStr}
	} else if len(CmdKeyFields) == 0 {
		return nil, errors.New("url missing api_name or data_command")
	}

	if svcContext.Cmd = strings.ToUpper(CmdKeyFields[0]); svcContext.Cmd == "" {
		return nil, errors.New("url missing api_name or data_command")
	}

	// cmd and key and field, i.g. /HGET/UserAvatar?F=fa4Y3oyQk2swURaJ
	// both cmd and key are required
	if len(CmdKeyFields) > 1 {
		svcContext.Key = CmdKeyFields[1]
	}
	// ensure there's a field for certain cmds
	svcContext.Field = r.FormValue("f")

	if err = svcContext.ParseJwtClaim(r); err != nil {
		return svcContext, err
	}

	return svcContext, nil
}
func (svc *DoptimeReqCtx) MergeJwtParam(paramIn map[string]interface{}) {
	for k := range paramIn {
		if strings.HasPrefix(k, "Jwt") {
			delete(paramIn, k)
		}
	}
	//add all Jwt fields to paramIn
	for k, v := range svc.Claims {
		//convert first letter of k to upper case
		k = strings.ToUpper(k[:1]) + k[1:]
		paramIn["Jwt"+k] = v
	}

}
func (svc *DoptimeReqCtx) MergeFormParam(Form url.Values, paramIn map[string]interface{}) {
	for key, value := range Form {
		if paramIn[key] = value[0]; len(value) > 1 {
			paramIn[key] = value // Assign the single value directly
		}
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
