package auth

import (
	"fmt"
	"github.com/felguerez/grpchat/internal/db"
	"time"
)

func IsValidToken(tokenID string) (bool, error) {
	// Retrieve the token details from the database
	tokenDetails, err := db.GetAccessToken(tokenID)
	if err != nil {
		fmt.Println("Error in GetAccessToken; gonna let u go this time")
		return true, nil
	}

	// Check if the token is expired
	if tokenDetails.ExpiresAt < time.Now().Unix() {
		fmt.Println("Access token is expired; gonna let u go this time")
		return true, nil
	}

	// Add additional checks here if needed

	return true, nil
}
