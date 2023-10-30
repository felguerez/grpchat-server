package wschat

import (
	"fmt"
	"github.com/felguerez/grpchat/internal/db"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strings"
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

// InitializeWebSocket initializes the WebSocket server and returns the http.HandlerFunc
func InitializeWebSocket() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// ... (your existing code)

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
		conversationId := parts[3]
		connections[conversationId] = append(connections[conversationId], ws)

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

func BroadcastMessageToWebSockets(message db.Message) {
	fmt.Printf("ConversationID: %s\n", message.ConversationID)
	fmt.Printf("connection count: %d\n", len(connections))
	fmt.Printf("connections to conversationId: %d\n", len(connections[message.ConversationID]))
	for _, conn := range connections[message.ConversationID] {
		if err := conn.WriteJSON(message); err != nil {
			log.Printf("Error sending message: %v", err)
			conn.Close()
			removeConnection(message.ConversationID, conn)
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
