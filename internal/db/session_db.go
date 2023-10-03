package db

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type Session struct {
	SessionID string `json:"session_id" dynamodbav:"SessionID"`
	UserID    string `json:"userID" dynamodbav:"UserID"`
	ExpiresAt int64  `json:"expiresAt" dynamodbav:"ExpiresAt"`
}

var UserSessionsTable = "UserSessions"

func PutSession(session Session) error {
	av, err := dynamodbattribute.MarshalMap(session)
	if err != nil {
		return err
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(UserSessionsTable),
	}

	_, err = Client().PutItem(input)
	return err
}

func GetSession(sessionID string) (*Session, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(UserSessionsTable),
		Key: map[string]*dynamodb.AttributeValue{
			"SessionID": {
				S: aws.String(sessionID),
			},
		},
	}

	result, err := Client().GetItem(input)
	if err != nil {
		return nil, err
	}

	var session Session
	if err := dynamodbattribute.UnmarshalMap(result.Item, &session); err != nil {
		return nil, err
	}

	return &session, nil
}

func DeleteSession(sessionID string) error {
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String("Sessions"), // Replace with your table name
		Key: map[string]*dynamodb.AttributeValue{
			"SessionID": {
				S: aws.String(sessionID),
			},
		},
	}

	_, err := Client().DeleteItem(input)
	return err
}
