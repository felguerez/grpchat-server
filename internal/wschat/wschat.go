package wschat

import (
	"fmt"
	"github.com/felguerez/grpchat/internal/db"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"log"
	"net/http"
	"strings"
	"sync"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// put logic here to allow or disallow origins, for example:
		fmt.Printf("origin: %s\n", r.Header.Get("Origin"))
		return true
	},
}

var connections = make(map[string][]*websocket.Conn)
var connectionsMutex sync.Mutex

// InitializeWebSocket initializes the WebSocket server and returns the http.HandlerFunc
func InitializeWebSocket() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Upgrade error: %v", err)
			return
		}
		log.Println("WebSocket upgraded")

		log.Printf("WebSocket connection established")
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) < 4 {
			http.Error(w, "Malformed URL", http.StatusBadRequest)
			return
		}
		log.Printf("parts looks like %v", parts)
		conversationId := parts[3]
		log.Printf("connecting to conversation %s", conversationId)
		connectionsMutex.Lock()
		connections[conversationId] = append(connections[conversationId], ws)
		connectionsMutex.Unlock()

		go func() {
			for {
				var message db.Message
				fmt.Printf("message is %v", message)
				err := ws.ReadJSON(&message)
				if err != nil {
					log.Printf("Error reading JSON: %v", err)
					break
				}
				broadcastMessage(message)
			}
		}()
	}
}

func broadcastMessage(message db.Message) {
	for _, conn := range connections[message.ConversationID] {
		if err := conn.WriteJSON(message); err != nil {
			log.Printf("Error sending message: %v", err)
			conn.Close()
			removeConnection(message.ConversationID, conn)
		}
	}
}

func BroadcastMessageToWebSockets(message db.Message, logger *zap.Logger) {
	connectionsMutex.Lock()
	defer connectionsMutex.Unlock()

	logger.Info("Broadcasting to conversation ID: %s\n", zap.String("conversationID", message.ConversationID))
	logger.Info("Total connections: %d\n", zap.Int("connections length", len(connections)))
	logger.Info("Connections to conversation ID %s: %d\n", zap.String("ConversationID", message.ConversationID), zap.Int("connections length", len(connections[message.ConversationID])))
	logger.Info("My message is", zap.Any("message", message))

	// Iterate over a copy of the slice to prevent issues if removeConnection modifies the original slice.
	connectionsCopy := make([]*websocket.Conn, len(connections[message.ConversationID]))
	copy(connectionsCopy, connections[message.ConversationID])

	for _, conn := range connectionsCopy {
		if err := conn.WriteJSON(message); err != nil {
			log.Printf("Error sending message to conversation ID %s: %v", message.ConversationID, err)
			conn.Close()                                   // Close the connection on error.
			removeConnection(message.ConversationID, conn) // Safely remove the connection.
		}
	}
}

func removeConnection(id string, ws *websocket.Conn) {
	conns, ok := connections[id]
	if !ok {
		return
	}

	for i, conn := range conns {
		if conn == ws {
			// Remove the connection from the slice
			conns = append(conns[:i], conns[i+1:]...)
			connections[id] = conns
			break
		}
	}
}
