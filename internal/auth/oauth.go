package auth

import (
	"context"
	"golang.org/x/oauth2"
)

var spotifyOauthConfig *oauth2.Config

func InitializeSpotifyOauthConfig(clientID, clientSecret, redirectURL string, scopes []string) {
	spotifyOauthConfig = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.spotify.com/authorize",
			TokenURL: "https://accounts.spotify.com/api/token",
		},
	}
}

func RefreshAccessToken(refreshToken string) (*oauth2.Token, error) {
	ctx := context.Background()
	token := &oauth2.Token{RefreshToken: refreshToken}
	ts := spotifyOauthConfig.TokenSource(ctx, token)
	newToken, err := ts.Token()
	if err != nil {
		return nil, err
	}
	return newToken, nil
}

func GetSpotifyOauthConfig() *oauth2.Config {
	return spotifyOauthConfig
}
