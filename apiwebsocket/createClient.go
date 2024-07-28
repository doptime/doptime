package apiwebsocket

import (
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/gorilla/websocket"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/vmihailenco/msgpack/v5"
)

var MsgToReceive = cmap.New[chan *ApiResponse]()

// CreateClient sends a message using the pool's connections
func CreateClient[i any, o any](url, jwt string) (f func(InParameter i) (ret o, err error)) {

	//support jwt should be enabled here
	//msg * ApiContext
	f = func(InParameter i) (ret o, err error) {
		pool, ok := WebSocketPoolMap.Get(url)
		if !ok {
			pool, err = newWebSocketPool(url, 1, 3)
			if err != nil {
				return ret, err
			}
			WebSocketPoolMap.Set(url, pool)
		}
		paramInBytes, err := msgpack.Marshal(InParameter)
		if err != nil {
			return ret, err
		}

		apiCtx := &ApiContext{
			Req: &ApiRequest{
				ParamIn: paramInBytes,
			},
			ReqID: big.NewInt(int64(xxhash.Sum64(paramInBytes)) ^ rand.Int63()).String(),
		}
		paramBytes, err := apiCtx.Bytes()
		if err != nil {
			log.Printf("Failed to marshal message: %v", err)
			return ret, err
		}

		for attempt := 0; attempt < pool.maxRetries; attempt++ {
			conn, connErr := pool.GetConnection()
			if connErr != nil {
				continue
			}

			err = conn.WriteMessage(websocket.BinaryMessage, paramBytes)
			if err != nil {
				return ret, err
			}
			messageChan := make(chan *ApiResponse, 1)
			MsgToReceive.Set(apiCtx.ReqID, messageChan)

			timeoutChan := time.After(20 * time.Second)
			select {
			case resp := <-messageChan:
				apiCtx.Resp = resp
				if resp.Error != "" {
					err = fmt.Errorf(resp.Error)
					return ret, err
				}
				err = msgpack.Unmarshal(apiCtx.Resp.Result, &ret)
				return ret, err
			case <-timeoutChan:
				MsgToReceive.Remove(apiCtx.ReqID)
				return ret, fmt.Errorf("timeout waiting for response")
			}
		}
		return ret, fmt.Errorf("max retries reached")
	}
	return f
}
