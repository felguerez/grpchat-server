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
	Id           string `json:"id,omitempty"`
}

type Message struct {
	UserID         string `json:"user_id"`
	Content        string `json:"content"`
	ConversationID int    `json:"conversation_id"`
	Timestamp      int64  `json:"timestamp"` // You can use Unix timestamp for ordering
}

var ChatMessagesTable *string
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

func PutAccessToken(item AccessToken) {
	av, err := dynamodbattribute.MarshalMap(item)
	if err != nil {
		log.Fatalf("Got error marshalling item: %s", err)
	}
	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: GetChatMessagesTable(),
	}
	_, err = Client().PutItem(input)
	if err != nil {
		log.Fatalf("Got error calling PutAccessToken: %s", err)
	}
	fmt.Println("Successfully added item to db:", item.AccessToken)
}

func GetAccessToken(key string) (*AccessToken, error) {
	input := &dynamodb.GetItemInput{
		TableName: GetChatMessagesTable(),
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
	// Marshal the Message struct into an AWS attribute value map
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
