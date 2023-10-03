package auth

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/felguerez/grpchat/internal/db"
	"time"
)

func IsValidSession(sessionID string) (bool, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(db.UserSessionsTable),
		Key: map[string]*dynamodb.AttributeValue{
			"SessionID": {
				S: aws.String(sessionID),
			},
		},
	}

	result, err := db.Client().GetItem(input)
	if err != nil {
		return false, err
	}

	if result.Item == nil {
		return false, nil
	}

	var session db.Session
	if err := dynamodbattribute.UnmarshalMap(result.Item, &session); err != nil {
		return false, err
	}

	if session.ExpiresAt < time.Now().Unix() {
		return false, nil
	}

	return true, nil
}
