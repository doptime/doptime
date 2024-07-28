package apiwebsocket

import (
	"sync"

	"github.com/gorilla/websocket"
	cmap "github.com/orcaman/concurrent-map/v2"
)

// WebSocketPoolMap stores WebSocket pools for different URLs
var WebSocketPoolMap = cmap.New[*WebSocketPool]()

// WebSocketPool manages a pool of WebSocket connections
type WebSocketPool struct {
	url        string
	mus        []*sync.Mutex
	conns      []*websocket.Conn
	Fails      []int
	maxRetries int
	index      int
}

// NewWebSocketPool initializes a new WebSocket pool with a callback for received messages
func newWebSocketPool(url string, maxConns int, maxRetries int) (*WebSocketPool, error) {
	pool := &WebSocketPool{
		url:        url,
		mus:        make([]*sync.Mutex, maxConns),
		conns:      make([]*websocket.Conn, maxConns),
		Fails:      make([]int, maxConns),
		maxRetries: maxRetries,
		index:      0,
	}

	for i := 0; i < maxConns; i++ {
		pool.mus[i] = &sync.Mutex{}
	}

	// Initialize connections
	for i := 0; i < maxConns; i++ {
		if err := pool.connect(i, pool.mus[i]); err != nil {
			return pool, err
		}

	}
	go pool.pingLoop()

	return pool, nil
}
