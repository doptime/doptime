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
	Ctx             context.Context
	Claims          jwt.MapClaims
	RedisDataSource string
	// case get
	Cmd     string
	Key     string
	Field   string
	SUToken string

	ResponseContentType string
}

var ErrIncompleteRequest = errors.New("incomplete request")

func NewHttpContext(ctx context.Context, r *http.Request, w http.ResponseWriter) (httpCtx *DoptimeReqCtx, err error) {
	var (
		CmdKeyFields                                  []string
		param, CmdKeyFieldsStr, pathStr, pathLastPart string
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

	//default response content type: application/json
	if svcContext.ResponseContentType, param = "application/json", r.FormValue("rt"); param != "" {
		svcContext.ResponseContentType = param
	}
	if svcContext.RedisDataSource, param = "default", r.FormValue("ds"); param != "" {
		svcContext.RedisDataSource = param
	}
	if svcContext.SUToken, param = "", r.FormValue("su"); param != "" {
		svcContext.SUToken = param
	}
	return svcContext, nil
}

// Ensure the body is msgpack format
func (svc *DoptimeReqCtx) MsgpackBody(r *http.Request, checkContentType bool, validateMsgpackFormat bool) (MsgPack []byte, err error) {
	var (
		data interface{}
	)
	if MsgPack, err = io.ReadAll(r.Body); len(MsgPack) == 0 || err != nil {
		return nil, fmt.Errorf("empty msgpack body")
	}
	//should make sure the data is msgpack format
	if err = msgpack.Unmarshal(MsgPack, &data); err != nil {
		return nil, err
	}
	if MsgPack, err = msgpack.Marshal(data); err != nil {
		return nil, err
	}
	//return remarshaled MsgPack, because golang MsgPack is better fullfill than javascript MsgPack
	return MsgPack, nil
}
