package handlers

import (
	"fmt"
	"github.com/felguerez/grpchat/internal/auth"
	"github.com/felguerez/grpchat/internal/db"
	"github.com/felguerez/grpchat/internal/spotify"
	"golang.org/x/oauth2"
	"log"
	"net/http"
	"time"
)

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	spotifyOauthConfig := auth.GetSpotifyOauthConfig()
	fmt.Println("GET /login")
	authURL := spotifyOauthConfig.AuthCodeURL("your-state", oauth2.AccessTypeOffline)
	http.Redirect(w, r, authURL, http.StatusFound)
}

func HandleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	ctx := r.Context()
	spotifyOauthConfig := auth.GetSpotifyOauthConfig()
	token, err := spotifyOauthConfig.Exchange(ctx, code)
	if err != nil {
		// Handle error
	}
	userID, err := spotify.GetSpotifyUserID(token.AccessToken)
	if err != nil {
		log.Fatalf("Could not get spotify user userID %s", err)
	}
	item := db.AccessToken{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		ExpiresAt:    token.Expiry.Unix(),
		ID:           userID,
	}

	sessionID, err := auth.GenerateSessionID()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	session := db.Session{
		SessionID: sessionID,
		UserID:    userID,
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(), // 24 hours from now
	}

	if err := db.PutSession(session); err != nil {
		log.Fatalf("Could not put new session: %s", err)
		return
	}

	auth.SetSessionCookie(w, sessionID)

	db.PutAccessToken(item)
	http.Redirect(w, r, "http://localhost:3001/home", http.StatusSeeOther)
}

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		// @TODO: combine RequireAuthorizationToken and RequireSessionCookie
		next.ServeHTTP(writer, request)
	})
}

func RequireAuthorizationToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		isValid, err := auth.IsValidToken(token)
		if err != nil || !isValid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func RequireSessionCookie(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		sessionID := cookie.Value
		isValid, err := auth.IsValidSession(sessionID)
		if err != nil || !isValid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	})
}
