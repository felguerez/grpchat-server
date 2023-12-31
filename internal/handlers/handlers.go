package handlers

import (
	"fmt"
	"github.com/felguerez/grpchat/internal/auth"
	"github.com/felguerez/grpchat/internal/db"
	"github.com/felguerez/grpchat/internal/spotify"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"net/http"
	"net/url"
	"os"
	"time"
)

func HandleLogin(logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		spotifyOauthConfig := auth.GetSpotifyOauthConfig()
		authURL := spotifyOauthConfig.AuthCodeURL("your-state", oauth2.AccessTypeOffline)
		http.Redirect(w, r, authURL, http.StatusFound)
	}
}

func HandleCallback(logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		ctx := r.Context()
		spotifyOauthConfig := auth.GetSpotifyOauthConfig()
		token, err := spotifyOauthConfig.Exchange(ctx, code)
		if err != nil {
			logger.Error("Error exchanging Spotify code for token", zap.Error(err))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		userID, err := spotify.GetSpotifyUserID(token.AccessToken)
		if err != nil {
			logger.Error("Could not get spotify user userID")
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
			logger.Fatal("Could not put new session", zap.Error(err))
			return
		}

		auth.SetSessionCookie(w, sessionID)

		db.PutAccessToken(item)
		fmt.Sprintf("session ID is %s", sessionID)
		clientURL := os.Getenv("CLIENT_URL")
		if clientURL == "" {
			logger.Error("CLIENT_URL not set in environment")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		u, err := url.Parse(clientURL)
		if err != nil {
			logger.Error("Could not parse CLIENT_URL", zap.Error(err))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		u.Path = "/api/authenticated"
		q := u.Query()
		q.Set("sessionId", sessionID)
		u.RawQuery = q.Encode()

		logger.Info("We're going to redirect", zap.String("CLIENT_URL", u.String()))
		w.Header().Set("Location", u.String())
		http.Redirect(w, r, u.String(), http.StatusFound)
	}
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
