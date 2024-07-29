package apipool

import "github.com/vmihailenco/msgpack/v5"

type ApiResponse struct {
	// to api response
	Result []byte
	Error  string
}
type ApiRequest struct {
	//belongs to api over websocket
	ParamIn []byte
}

type ApiContext struct {
	Name  string
	ReqID string
	Req   *ApiRequest
	Resp  *ApiResponse
}

func (msg *ApiContext) Bytes() (bs []byte, err error) {
	return msgpack.Marshal(msg)
}
