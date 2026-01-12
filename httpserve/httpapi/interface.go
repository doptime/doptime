package httpapi

import "context"

type ApiInterface interface {
	GetName() string
	CallByMap(ctx context.Context, _map map[string]interface{}, msgpackNonstruct []byte, jsonpackNostruct []byte) (ret interface{}, err error)
	GetDataSource() string
}
