package db

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"sort"
)

type Conversation struct {
	ID        string   `json:"id" dynamodbav:"ID"`
	Name      string   `json:"name" dynamodbav:"Name"`
	CreatedAt int64    `json:"createdAt" dynamodbav:"CreatedAt"`
	UpdatedAt int64    `json:"updatedAt" dynamodbav:"UpdatedAt"`
	CreatedBy string   `json:"createdBy" dynamodbav:"CreatedBy"`
	Members   []string `json:"members" dynamodbav:"Members"`
}

var ConversationsTable = "Conversations"

func GetConversations(userID string, limit int64, sortBy string) ([]Conversation, error) {
	svc := Client()
	input := &dynamodb.ScanInput{
		TableName: aws.String("Conversations"),
		Limit:     aws.Int64(limit),
	}

	if sortBy != "" {
		// DynamoDB-specific code to sort by the given attribute
		// This often requires setting up a secondary index
	}

	result, err := svc.Scan(input)
	if err != nil {
		return nil, err
	}

	var conversations []Conversation
	for _, i := range result.Items {
		conversation := Conversation{}
		if err := dynamodbattribute.UnmarshalMap(i, &conversation); err != nil {
			return nil, err
		}
		// Assuming Members is a slice of string user IDs
		for _, member := range conversation.Members {
			if member == userID {
				conversations = append(conversations, conversation)
				break
			}
		}
	}

	// If you couldn't sort in the database query, sort here in Go code
	if sortBy == "CreatedAt" {
		sort.Slice(conversations, func(i, j int) bool {
			return conversations[i].CreatedAt < conversations[j].CreatedAt
		})
	} else if sortBy == "UpdatedAt" {
		sort.Slice(conversations, func(i, j int) bool {
			return conversations[i].UpdatedAt < conversations[j].UpdatedAt
		})
	}

	return conversations, nil
}

func PutConversation(conversation Conversation) error {
	svc := Client()
	av, err := dynamodbattribute.MarshalMap(conversation)
	if err != nil {
		return err
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String("Conversations"),
		Item:      av,
	}

	_, err = svc.PutItem(input)
	if err != nil {
		return err
	}
	return nil
}
