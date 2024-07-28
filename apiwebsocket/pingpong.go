package apiwebsocket

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

// pingLoop sends Ping messages periodically and checks for Pong responses
func (pool *WebSocketPool) pingLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		for index, conn := range pool.conns {
			mu := pool.mus[index]
			mu.Lock()
			err := conn.WriteMessage(websocket.PingMessage, nil)
			mu.Unlock()

			if err != nil {
				pool.Fails[index]++
			} else if pool.Fails[index] >= pool.maxRetries {
				log.Printf("Max retries reached for connection %d, reconnecting...", index)
				pool.conns[index].Close()
				pool.connect(index, mu)
			}
		}
	}

}
