package httpserve

import (
	"fmt"
	"log"
	"net/http"
	"time"

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

var Token2ClaimMap map[string]jwt.MapClaims = make(map[string]jwt.MapClaims)

func websocketAPI(w http.ResponseWriter, r *http.Request) {
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

	for {
		mt, message, err := ws.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		if mt == websocket.BinaryMessage {
			var doptimeReqCtx DoptimeReqCtx
			doptimeReqCtx.UpdateKeyFieldWithJwtClaims()
			if err = msgpack.Unmarshal(message, &doptimeReqCtx); err != nil {
				log.Println("msgpack.Unmarshal:", err)
				continue
			}
			doptimeReqCtx.Claims = claims

		}
	}
	w.WriteHeader(http.StatusOK)
}
