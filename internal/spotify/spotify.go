package spotify

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type SpotifyUser struct {
	ID string `json:"id"`
}

func GetSpotifyUserID(accessToken string) (string, error) {
	req, err := http.NewRequest("GET", "https://api.spotify.com/v1/me", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var user SpotifyUser
	if err := json.Unmarshal(body, &user); err != nil {
		return "", err
	}

	return user.ID, nil
}
