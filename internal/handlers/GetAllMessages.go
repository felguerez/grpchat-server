package handlers

import (
	"encoding/json"
	"github.com/felguerez/grpchat/internal/db"
	"net/http"
)

func HandleGetAllMessages(w http.ResponseWriter, r *http.Request) {
	messages, err := db.GetAllMessages()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}
