package handlers

import (
	"encoding/json"
	"github.com/felguerez/grpchat/internal/db"
	"net/http"
)

func HandleSendMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var msg db.Message
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&msg)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	err = db.PutMessage(msg)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Message sent successfully"))
}
