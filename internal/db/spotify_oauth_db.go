package db

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"log"
)

var SpotifyAccessTokensTable = "SpotifyAccessTokens"

func PutAccessToken(item AccessToken) {
	if item.ID == "" {
		log.Fatalf("ID field must not be empty")
		return
	}
	av, err := dynamodbattribute.MarshalMap(item)
	if err != nil {
		log.Fatalf("Got error marshalling item: %s", err)
	}
	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: &SpotifyAccessTokensTable,
	}
	_, err = Client().PutItem(input)
	if err != nil {
		log.Fatalf("Got error calling PutAccessToken: %s", err)
	}
}

func GetAccessToken(key string) (*AccessToken, error) {
	input := &dynamodb.GetItemInput{
		TableName: &SpotifyAccessTokensTable,
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(key),
			},
		},
	}
	result, err := Client().GetItem(input)
	if err != nil {
		return nil, err
	}
	var item AccessToken
	if err := dynamodbattribute.UnmarshalMap(result.Item, &item); err != nil {
		return nil, err
	}
	return &item, nil
}

//func RefreshAndSaveToken(refreshToken string, userID string) error {
//	newToken, err := auth.RefreshAccessToken(refreshToken)
//	if err != nil {
//		return err
//	}
//	item := AccessToken{
//		AccessToken:  newToken.AccessToken,
//		RefreshToken: newToken.RefreshToken,
//		TokenType:    newToken.TokenType,
//		ExpiresAt:    newToken.Expiry.Unix(),
//		ID:           userID,
//	}
//	PutAccessToken(item)
//	return nil
//}
