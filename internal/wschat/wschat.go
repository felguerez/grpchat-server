package wschat

import (
	"bytes"
	"github.com/felguerez/grpchat/internal/db"
	"github.com/felguerez/grpchat/internal/wsutil" // Ensure this path is correct
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

type Client struct {
	hub  *Hub
	conn *wsutil.WebSocketConnection
	send chan []byte
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.Conn.SetReadLimit(maxMessageSize)
	c.conn.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.Conn.SetPongHandler(func(string) error { c.conn.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.hub.logger.Error("Error reading message", zap.Error(err))
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		c.hub.broadcast <- message
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				c.hub.logger.Error("Error writing message", zap.Any("message", message), zap.String("method", "writePump"), zap.Time("timestamp", time.Now()), zap.Any("message", message))
				return
			}
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)
		// Add queued messages to the current ws message
		// @TODO: handle errors below
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.hub.logger.Error("Error writing message", zap.Error(err), zap.String("method", "writePump"), zap.Time("timestamp", time.Now()))
				return
			}
		}
	}
}

func ServeWebSocketConnection(hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			// @TODO: handle error
			return
		}
		hub.logger.Info("WebSocket upgraded")
		client := &Client{hub: hub, conn: wsutil.NewWebSocketConnection(conn), send: make(chan []byte, 256)}
		client.hub.register <- client
		go client.writePump()
		go client.readPump()
	}
}

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// @TODO: secure this function
		return true
	},
}

var connections = make(map[string][]*wsutil.WebSocketConnection)
var connectionsMutex sync.Mutex

func InitializeWebSocket(logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logger.Info("Upgrade error: ", zap.Error(err))
			return
		}
		logger.Info("WebSocket upgraded")

		parts := strings.Split(r.URL.Path, "/")
		if len(parts) < 4 {
			http.Error(w, "Malformed URL", http.StatusBadRequest)
			return
		}
		conversationId := parts[3]
		wrappedConnection := wsutil.NewWebSocketConnection(ws) // Wrap the connection

		connectionsMutex.Lock()
		connections[conversationId] = append(connections[conversationId], wrappedConnection)
		connectionsMutex.Unlock()

		go alsoHandleWebSocketConnection(wrappedConnection, conversationId, logger)
	}
}

func alsoHandleWebSocketConnection(conn *wsutil.WebSocketConnection, conversationId string, logger *zap.Logger) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Error("Error reading message", zap.Error(err))
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		broadcastMessage(db.Message{Content: string(message), ConversationID: conversationId}, logger)
	}
	removeConnection(conversationId, conn) // Remove connection when done
}

func handleWebSocketConnection(conn *wsutil.WebSocketConnection, conversationId string, logger *zap.Logger) {
	for {
		var message db.Message
		logger.Info("Message received", zap.Time("timestamp", time.Now()), zap.Any("message", message), zap.String("method", "handleWebSocketConnection"))
		err := conn.Conn.ReadJSON(&message) // Use conn.Conn to access the underlying WebSocket
		if err != nil {
			logger.Error("Error reading JSON", zap.Error(err), zap.String("method", "handleWebSocketConnection"), zap.Time("timestamp", time.Now()), zap.Any("message", message))
			break
		}
		broadcastMessage(message, logger)
	}
	removeConnection(conversationId, conn) // Remove connection when done
}

func broadcastMessage(message db.Message, logger *zap.Logger) {
	connectionsMutex.Lock()
	connectionsCopy := make([]*wsutil.WebSocketConnection, len(connections[message.ConversationID]))
	copy(connectionsCopy, connections[message.ConversationID])
	connectionsMutex.Unlock()

	for _, conn := range connectionsCopy {
		if conn.IsOpen() { // Check if the connection is ready
			if err := conn.Conn.WriteJSON(message); err != nil {
				logger.Error("Error sending message: %v", zap.Error(err))
				conn.Close()
			}
		}
	}
}

func BroadcastMessageToWebSockets(message db.Message, logger *zap.Logger) {
	connectionsMutex.Lock()
	defer connectionsMutex.Unlock()
	logger.Info("Preparing to broadcast message", zap.Time("timestamp", time.Now()), zap.Any("message", message), zap.String("method", "BroadcastMessageToWebSockets"))
	connectionsCopy := make([]*wsutil.WebSocketConnection, len(connections[message.ConversationID]))
	copy(connectionsCopy, connections[message.ConversationID])

	for _, conn := range connectionsCopy {
		logger.Info("Checking connection status before sending message", zap.String("conversationID", message.ConversationID))
		if conn != nil && conn.IsOpen() {
			logger.Info("Connection is open, attempting to send message", zap.String("conversationID", message.ConversationID))
		} else {
			logger.Info("Connection is not open, skipping", zap.String("conversationID", message.ConversationID))
			continue
		}

		if err := conn.WriteJSON(message); err != nil {
			logger.Error("Error sending message", zap.String("conversationID", message.ConversationID), zap.Time("timestamp", time.Now()),
				zap.Error(err), zap.String("method", "BroadcastMessageToWebSockets"))
			conn.Close() // Close the connection on error.
			removeConnection(message.ConversationID, conn)
			logger.Info("Connection closed and removed due to error", zap.Time("timestamp", time.Now()),
				zap.String("conversationID", message.ConversationID), zap.String("method", "BroadcastMessageToWebSockets"))
		} else {
			logger.Info("Message sent successfully", zap.String("conversationID", message.ConversationID), zap.Time("timestamp", time.Now()), zap.String("method", "BroadcastMessageToWebSockets"))
		}
	}
}

func removeConnection(id string, ws *wsutil.WebSocketConnection) {
	existingConnections, ok := connections[id]
	if !ok {
		return
	}

	for i, conn := range existingConnections {
		if conn == ws {
			existingConnections = append(existingConnections[:i], existingConnections[i+1:]...)
			connections[id] = existingConnections
			break
		}
	}
}
