package httpserve

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/vmihailenco/msgpack/v5"
)

type DoptimeReqCtx struct {
	Ctx    context.Context
	Claims jwt.MapClaims
	// case get
	Cmd   string
	Key   string
	Field string
}

var ErrIncompleteRequest = errors.New("incomplete request")

func NewHttpContext(ctx context.Context, r *http.Request, w http.ResponseWriter) (httpCtx *DoptimeReqCtx, err error) {
	var (
		CmdKeyFields                           []string
		CmdKeyFieldsStr, pathStr, pathLastPart string
	)
	svcContext := &DoptimeReqCtx{Ctx: ctx}

	//i.g. https://url.com/rSvc/HGET-UserAvatar=fa4Y3oyQk2swURaJ?Queries=*&RspType=image/jpeg
	if pathStr = r.URL.Path; r.URL.RawPath != "" {
		pathStr = r.URL.RawPath
	}
	if pathLastPart = path.Base(pathStr); pathLastPart == "" {
		return nil, ErrIncompleteRequest
	}
	if CmdKeyFieldsStr, err = url.QueryUnescape(pathLastPart); err != nil {
		return nil, err
	}
	if CmdKeyFields = strings.SplitN(CmdKeyFieldsStr, "-", 2); len(CmdKeyFields) != 2 {
		return nil, ErrIncompleteRequest
	}
	// cmd and key and field, i.g. /HGET/UserAvatar?F=fa4Y3oyQk2swURaJ
	// both cmd and key are required
	if svcContext.Cmd, svcContext.Key = strings.ToUpper(CmdKeyFields[0]), CmdKeyFields[1]; svcContext.Cmd == "" || svcContext.Key == "" {
		return nil, ErrIncompleteRequest
	}
	//url decoded already
	svcContext.Field = r.FormValue("f")

	if err = svcContext.ParseJwtClaim(r); err != nil {
		return svcContext, err
	}

	return svcContext, nil
}
func (svc *DoptimeReqCtx) isValid() bool {
	return svc.Cmd != "" && svc.Key != ""
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

// Ensure the body is msgpack format
func (svc *DoptimeReqCtx) MsgpackBody(r *http.Request, checkContentType bool, validateMsgpackFormat bool) (MsgPack []byte, err error) {
	var (
		data interface{}
	)
	if checkContentType && r.Header.Get("Content-Type") != "application/octet-stream" {
		return nil, fmt.Errorf("invalid content type")
	}
	if MsgPack, err = io.ReadAll(r.Body); len(MsgPack) == 0 || err != nil {
		return nil, fmt.Errorf("empty msgpack body")
	}
	if validateMsgpackFormat {
		//should make sure the data is msgpack format
		if err = msgpack.Unmarshal(MsgPack, &data); err != nil {
			return nil, err
		}
		if MsgPack, err = msgpack.Marshal(data); err != nil {
			return nil, err
		}
	}
	//return remarshaled MsgPack, because golang MsgPack is better fullfill than javascript MsgPack
	return MsgPack, nil
}
