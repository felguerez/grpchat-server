package handlers

import (
	"fmt"
	"github.com/felguerez/grpchat/internal/auth"
	"github.com/felguerez/grpchat/internal/db"
	"github.com/felguerez/grpchat/internal/spotify"
	"golang.org/x/oauth2"
	"log"
	"net/http"
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
	id, err := spotify.GetSpotifyUserID(token.AccessToken)
	if err != nil {
		log.Fatalf("Could not get spotify user id %s", err)
	}
	item := db.AccessToken{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		ExpiresAt:    token.Expiry.Unix(),
		ID:           id,
	}

	// Save the token to DynamoDB
	db.PutAccessToken(item)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Token saved successfully"))
}
