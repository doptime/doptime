package apiwebsocket

import (
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// connect establishes a connection for the given index
func (pool *WebSocketPool) connect(index int, mu *sync.Mutex) error {
	conn, _, err := websocket.DefaultDialer.Dial(pool.url, nil)
	if err != nil {
		return err
	}
	mu.Lock()
	pool.conns[index] = conn
	pool.Fails[index] = 0
	mu.Unlock()

	conn.SetPongHandler(func(appData string) error {
		mu.Lock()
		pool.Fails[index] = 0
		mu.Unlock()
		return nil
	})

	conn.SetPingHandler(func(appData string) error {
		return conn.WriteControl(websocket.PongMessage, []byte{}, time.Now().Add(time.Second*10))
	})

	go pool.listenForMessages(conn)

	return nil
}

// GetConnection gets the next available connection from the pool
func (pool *WebSocketPool) GetConnection() (*websocket.Conn, error) {
	for i := 0; i < len(pool.conns); i++ {
		if pool.conns[pool.index] != nil {
			conn := pool.conns[pool.index]
			pool.index = (pool.index + 1) % len(pool.conns)
			return conn, nil
		}
		pool.index = (pool.index + 1) % len(pool.conns)
	}

	return nil, fmt.Errorf("no available connections")
}
