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

type HttpContext struct {
	Req             *http.Request
	Rsb             http.ResponseWriter
	jwtToken        *jwt.Token
	Ctx             context.Context
	RedisDataSource string
	// case get
	Cmd     string
	Key     string
	Field   string
	SUToken string

	ResponseContentType string
}

var ErrIncompleteRequest = errors.New("incomplete request")

func NewHttpContext(ctx context.Context, r *http.Request, w http.ResponseWriter) (httpCtx *HttpContext, err error) {
	var (
		CmdKeyFields                                  []string
		param, CmdKeyFieldsStr, pathStr, pathLastPart string
	)
	svcContext := &HttpContext{Req: r, Rsb: w, Ctx: ctx}

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

func (svc *HttpContext) MsgpackBodyBytes() (data []byte) {
	var (
		err error
	)
	if svc.Req.ContentLength == 0 {
		return nil
	}
	if !strings.HasPrefix(svc.Req.Header.Get("Content-Type"), "application/octet-stream") {
		return nil
	}
	if data, err = io.ReadAll(svc.Req.Body); err != nil {
		return nil
	}
	return data
}
func (svc *HttpContext) JsonBodyBytes() (data []byte) {
	var (
		err error
	)
	if svc.Req.ContentLength == 0 {
		return nil
	}
	if !strings.HasPrefix(svc.Req.Header.Get("Content-Type"), "application/json") {
		return nil
	}
	if data, err = io.ReadAll(svc.Req.Body); err != nil {
		return nil
	}
	return data
}

// Ensure the body is msgpack format
func (svc *HttpContext) MsgpackBody() (bytes []byte, err error) {
	var (
		data interface{}
	)
	if bytes = svc.MsgpackBodyBytes(); len(bytes) == 0 {
		return nil, fmt.Errorf("empty msgpack body")
	}
	//should make sure the data is msgpack format
	if err = msgpack.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}
	if bytes, err = msgpack.Marshal(data); err != nil {
		return nil, err
	}
	//return remarshaled bytes, because golang msgpack is better fullfill than javascript msgpack
	return bytes, nil
}
