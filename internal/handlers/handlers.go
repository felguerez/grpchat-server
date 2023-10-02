package handlers

import (
	"github.com/felguerez/grpchat/internal/auth"
	"github.com/felguerez/grpchat/internal/db"
	"golang.org/x/oauth2"
	"net/http"
)

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	spotifyOauthConfig := auth.GetSpotifyOauthConfig()
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
	item := db.AccessToken{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		ExpiresAt:    token.Expiry.Unix(),
		Id:           "felguerez", // @TODO: need username?
	}

	// Save the token to DynamoDB
	db.PutAccessToken(item)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Token saved successfully"))
}
