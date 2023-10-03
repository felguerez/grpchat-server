package auth

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

func GenerateSessionID() (string, error) {
	b := make([]byte, 16) // 128 bits
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func SetSessionCookie(w http.ResponseWriter, sessionID string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // set to true if you are using HTTPS
	})
}
