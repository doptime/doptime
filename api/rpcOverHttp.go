package api

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"reflect"
	"time"

	"github.com/doptime/doptime/config"
	"github.com/doptime/doptime/specification"
	"github.com/rs/zerolog/log"
	"github.com/vmihailenco/msgpack/v5"
)

func callViaHttp(url string, jwt string, InParam interface{}, retValueWithPointer interface{}) (err error) {
	var (
		b, revBytes []byte
		req         *http.Request
		resp        *http.Response
	)
	if b, err = specification.MarshalApiInput(InParam); err != nil {
		return err
	}

	//set timeout to 10s
	client := &http.Client{Timeout: 10 * time.Second}

	if req, err = http.NewRequest("POST", url, bytes.NewBuffer(b)); err != nil {
		return err
	}
	if len(jwt) > 0 {
		req.Header.Add("Authorization", "Bearer "+jwt)
	}
	req.Header.Add("Content-Type", "application/octet-stream")

	if resp, err = client.Do(req); err != nil {
		return err
	}
	defer resp.Body.Close()
	if revBytes, err = io.ReadAll(resp.Body); err != nil {
		return err
	}
	return msgpack.Unmarshal(revBytes, retValueWithPointer)
}

// this is designed to be used for point to point RPC. without dispatching parameter using redis
func RpcOverHttp[i any, o any](options ...*ApiOption) (rpc *Context[i, o]) {
	var option *ApiOption = mergeNewOptions(&ApiOption{ApiSourceHttp: "doptime", Name: specification.ApiNameByType((*i)(nil))}, options...)

	httpServer, err := config.GetHttpServerByName(option.ApiSourceHttp)
	if err != nil {
		log.Info().AnErr("DataSource not defined in enviroment", err).Send()
		return nil
	}
	rpc = &Context[i, o]{Name: option.Name, ApiSourceHttp: httpServer, Ctx: context.Background(),
		WithHeader: HeaderFieldsUsed(reflect.TypeOf(new(i)).Elem()),
		Validate:   needValidate(reflect.TypeOf(new(i)).Elem()),
	}
	rpc.Fun = func(InParam i) (ret o, err error) {
		oType := reflect.TypeOf((*o)(nil)).Elem()
		//if o type is a pointer, use reflect.New to create a new pointer
		if oType.Kind() == reflect.Ptr {
			ret = reflect.New(oType.Elem()).Interface().(o)
			return ret, callViaHttp(rpc.ApiSourceHttp.UrlBase+"/API-!"+rpc.Name+"-!rt~application%2Fmsgpack", rpc.ApiSourceHttp.Jwt, InParam, ret)
		}
		oValueWithPointer := reflect.New(oType).Interface().(*o)
		return *oValueWithPointer, callViaHttp(rpc.ApiSourceHttp.UrlBase+"/API-!"+rpc.Name+"-!rt~application%2Fmsgpack", rpc.ApiSourceHttp.Jwt, InParam, oValueWithPointer)

	}

	ApiServices.Set(rpc.Name, rpc)
	funcPtr := reflect.ValueOf(rpc.Fun).Pointer()
	fun2Api.Set(funcPtr, rpc)
	return rpc
}
