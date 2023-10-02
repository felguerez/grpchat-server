package db

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type AccessToken struct {
	ExpiresAt    int64  `json:"expiresAt,omitempty"`
	AccessToken  string `json:"accessToken,omitempty"`
	RefreshToken string `json:"refreshToken,omitempty"`
	TokenType    string `json:"tokenType,omitempty"`
	ID           string `json:"id,omitempty" dynamodbav:"ID"`
}

type Message struct {
	UserID         string `json:"user_id"`
	Content        string `json:"content"`
	ConversationID int    `json:"conversation_id"`
	Timestamp      int64  `json:"timestamp"`
}

var ChatMessagesTable *string
var SpotifyAccessTokensTable *string
var svc *dynamodb.DynamoDB

func GetChatMessagesTable() *string {
	if ChatMessagesTable == nil {
		ChatMessagesTable = aws.String(os.Getenv("CHAT_MESSAGES_TABLE"))
		if *ChatMessagesTable == "" {
			log.Fatalf("CHAT_MESSAGES_TABLE environment variable is not set")
		}
	}
	return ChatMessagesTable
}

func GetSpotifyAccessTokensTable() *string {
	if SpotifyAccessTokensTable == nil {
		SpotifyAccessTokensTable = aws.String(os.Getenv("SPOTIFY_ACCESS_TOKENS_TABLE"))
		if *SpotifyAccessTokensTable == "" {
			log.Fatalf("SPOTIFY_ACCESS_TOKENS_TABLE environment variable is not set")
		}
	}
	return SpotifyAccessTokensTable
}

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
		TableName: GetSpotifyAccessTokensTable(),
	}
	_, err = Client().PutItem(input)
	if err != nil {
		log.Fatalf("Got error calling PutAccessToken: %s", err)
	}
	fmt.Println("Successfully added item to db:", item.AccessToken)
}

func GetAccessToken(key string) (*AccessToken, error) {
	input := &dynamodb.GetItemInput{
		TableName: GetSpotifyAccessTokensTable(),
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

func Client() *dynamodb.DynamoDB {
	if svc != nil {
		return svc
	}
	config := aws.NewConfig().
		WithRegion("us-east-1").
		WithCredentials(credentials.NewStaticCredentials(os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"), ""))
	sess, err := session.NewSession(config)
	if err != nil {
		log.Fatalf("Failed to initialize AWS session: %s", err)
	}
	svc = dynamodb.New(sess)
	return svc
}

func PutMessage(msg Message) error {
	svc := Client()
	av, err := dynamodbattribute.MarshalMap(msg)
	if err != nil {
		return err
	}

	// Create the input for the PutItem operation
	fmt.Println("Table Name:", os.Getenv("CHAT_MESSAGES_TABLE"))
	input := &dynamodb.PutItemInput{
		TableName: GetChatMessagesTable(),
		Item:      av,
	}

	// Put the item into the DynamoDB table
	_, err = svc.PutItem(input)
	return err
}
