package httpserve

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
	Cmd   string
	Key   string
	Field string

	ResponseContentType string
}

var ErrIncompleteRequest = errors.New("incomplete request")

func NewHttpContext(ctx context.Context, r *http.Request, w http.ResponseWriter) (httpCtx *HttpContext, err error) {
	var (
		CmdKeyFields           []string
		param, CmdKeyFieldsStr string
	)
	svcContext := &HttpContext{Req: r, Rsb: w, Ctx: ctx}
	//i.g. https://url.com/rSvc/HGET=UserAvatar=fa4Y3oyQk2swURaJ?Queries=*&RspType=image/jpeg
	if CmdKeyFields = strings.Split(r.URL.RawPath, "/"); len(CmdKeyFields) < 1 {
		return nil, ErrIncompleteRequest
	}
	if CmdKeyFieldsStr, err = url.QueryUnescape(CmdKeyFields[len(CmdKeyFields)-1]); err != nil {
		return nil, err
	}
	if CmdKeyFields = strings.Split(CmdKeyFieldsStr, "-!"); len(CmdKeyFields) < 2 {
		return nil, ErrIncompleteRequest
	}
	// cmd and key and field, i.g. /HGET/UserAvatar?F=fa4Y3oyQk2swURaJ
	// both cmd and key are required
	if svcContext.Cmd, svcContext.Key = CmdKeyFields[0], CmdKeyFields[1]; svcContext.Cmd == "" || svcContext.Key == "" {
		return nil, ErrIncompleteRequest
	}
	//url decoded already
	svcContext.Field = r.FormValue("F")

	//default response content type: application/json
	svcContext.ResponseContentType = "application/json"
	svcContext.RedisDataSource = "default"
	for i, l := 2, len(CmdKeyFields); i < l; i++ {
		if param = CmdKeyFields[i]; len(param) <= 3 {
			continue
		}
		switch param[:3] {
		case "rt~": //response content type rt=application/json
			svcContext.ResponseContentType = param[3:]
		case "ds~": //redis db name RDB=redisDataSource
			svcContext.RedisDataSource = param[3:]
		}
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
