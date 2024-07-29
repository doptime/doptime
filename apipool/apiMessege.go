package apipool

import (
	"log"

	"github.com/gorilla/websocket"
	"github.com/vmihailenco/msgpack/v5"
)

// listenForMessages listens for messages and handles them using the callback
func (pool *WebSocketPool) listenForMessages(conn *websocket.Conn) {
	defer conn.Close()
	for {
		mt, message, err := conn.ReadMessage()
		if mt != websocket.BinaryMessage {
			continue
		}
		if err != nil {
			log.Printf("Error reading message: %v", err)
			return
		}
		apiMessege := &ApiContext{}
		if err := msgpack.Unmarshal(message, apiMessege); err != nil {
			log.Printf("Error unmarshalling message: %v", err)
			return
		}
		if apiMessege.Resp != nil {
			//this is a callback message
			if _chan, ok := MsgToReceive.Get(apiMessege.ReqID); ok {
				_chan <- apiMessege.Resp
			}

		} else if callback, ok := ServerApis.Get(pool.url + ":" + apiMessege.Name); !ok {
			log.Printf("No callback for URL: %s", pool.url)
		} else {
			callback(apiMessege)
		}
	}
}
