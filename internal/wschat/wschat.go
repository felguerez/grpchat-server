package wschat

import (
	"fmt"
	"github.com/felguerez/grpchat/internal/db"
	"github.com/felguerez/grpchat/internal/wsutil" // Ensure this path is correct
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"sync"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		fmt.Printf("origin: %s\n", r.Header.Get("Origin"))
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

		go handleWebSocketConnection(wrappedConnection, conversationId, logger)
	}
}

func handleWebSocketConnection(conn *wsutil.WebSocketConnection, conversationId string, logger *zap.Logger) {
	for {
		var message db.Message
		err := conn.Conn.ReadJSON(&message) // Use conn.Conn to access the underlying WebSocket
		if err != nil {
			logger.Error("Error reading JSON", zap.Error(err))
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

	connectionsCopy := make([]*wsutil.WebSocketConnection, len(connections[message.ConversationID]))
	copy(connectionsCopy, connections[message.ConversationID])

	for _, conn := range connectionsCopy {
		if conn.IsOpen() { // Check if the connection is ready
			if err := conn.Conn.WriteJSON(message); err != nil {
				logger.Error("Error sending message", zap.String("message.ConversationID", message.ConversationID), zap.Error(err))
				conn.Close()
				removeConnection(message.ConversationID, conn)
			}
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
