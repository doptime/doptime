package httpserve

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/doptime/doptime/api"
	"github.com/doptime/doptime/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/vmihailenco/msgpack/v5"
)

// WebSocket 升级器
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WsParam struct {
	Api   string
	Param []byte
}
type DoptimeRespCtx struct {
	Data  []byte
	ReqID string
}

func (ctx *DoptimeRespCtx) Response(ws *websocket.Conn, mu *sync.Mutex, result interface{}) (err error) {
	ctx.Data, err = msgpack.Marshal(result)
	if err != nil {
		return
	}
	mu.Lock()
	defer mu.Unlock()
	ws.WriteMessage(websocket.BinaryMessage, ctx.Data)
	return
}

var Token2ClaimMap map[string]jwt.MapClaims = make(map[string]jwt.MapClaims)

func websocketAPICallback(w http.ResponseWriter, r *http.Request) {
	var claims jwt.MapClaims
	var ok bool
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer ws.Close()
	if len(config.Cfg.Http.JwtSecret) > 0 {
		jwtStr := w.Header().Get("Authorization")
		if jwtStr == "" {
			ws.WriteMessage(websocket.TextMessage, []byte("Missing Authorization in Header,Unauthorized!"))
			return
		}
		if claims, ok = Token2ClaimMap[jwtStr]; !ok {
			if claims, err = Jwt2Claim(jwtStr, config.Cfg.Http.JwtSecret); err != nil {
				ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Invalid JWT token: %v", err)))
				return
			}
			Token2ClaimMap[jwtStr] = claims
		}
	}
	//enable auto ping
	ws.SetReadLimit(512)
	pongWait := 60 * time.Second
	ws.SetReadDeadline(time.Now().Add(pongWait))
	ws.SetPongHandler(func(string) error { ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	var mu sync.Mutex
	for {
		mt, message, err := ws.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		if mt == websocket.BinaryMessage {
			var reqCtx DoptimeReqCtx
			var _api api.ApiInterface
			var result interface{}
			reqCtx.UpdateKeyFieldWithJwtClaims()
			if err = msgpack.Unmarshal(message, &reqCtx); err != nil {
				log.Println("msgpack.Unmarshal:", err)
				continue
			}
			reqCtx.Claims = claims
			var paramIn map[string]interface{} = make(map[string]interface{})
			reqCtx.MergeJwtParam(paramIn)
			rsp := DoptimeRespCtx{ReqID: reqCtx.ReqID}
			if !reqCtx.isValid() {
				rsp.Response(ws, &mu, "err invalid request")
				continue
			}

			err = msgpack.Unmarshal(reqCtx.Data, &paramIn)
			if err != nil {
				rsp.Response(ws, &mu, err)
				continue
			}
			//always response with msgpack format

			if _api, ok = api.GetApiByName(reqCtx.Key); !ok {
				rsp.Response(ws, &mu, "err no such api")
			}
			_api.MergeHeaderParam(r, paramIn)

			result, err = _api.CallByMap(paramIn)
			if err != nil {
				rsp.Response(ws, &mu, err)
				continue
			}
			rsp.Response(ws, &mu, result)

		}
	}
	w.WriteHeader(http.StatusOK)
}
