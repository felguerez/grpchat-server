package wsutil

import (
	"github.com/gorilla/websocket"
	"sync"
)

type WebSocketConnection struct {
	*websocket.Conn
	closed bool
	mutex  sync.Mutex
}

func NewWebSocketConnection(wsConn *websocket.Conn) *WebSocketConnection {
	return &WebSocketConnection{
		Conn:   wsConn,
		closed: false,
	}
}

// SetClosed Safely set the connection as closed
func (wsc *WebSocketConnection) SetClosed() {
	wsc.mutex.Lock()
	defer wsc.mutex.Unlock()
	wsc.closed = true
}

// IsOpen Check if the connection is open
func (wsc *WebSocketConnection) IsOpen() bool {
	wsc.mutex.Lock()
	defer wsc.mutex.Unlock()
	return !wsc.closed
}

// SafeClose Safely close the connection
func (wsc *WebSocketConnection) SafeClose() {
	wsc.SetClosed()
	wsc.Close()
}
