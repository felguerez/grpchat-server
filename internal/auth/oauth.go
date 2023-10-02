package auth

import "golang.org/x/oauth2"

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

func GetSpotifyOauthConfig() *oauth2.Config {
	return spotifyOauthConfig
}
