package db

import (
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type Message struct {
	UserID         string `json:"user_id" dynamodbav:"UserID"`
	Content        string `json:"content" dynamodbav:"Content"`
	ConversationID int32  `json:"conversation_id" dynamodbav:"ConversationID"`
	Timestamp      int64  `json:"timestamp" dynamodbav:"Timestamp"`
}

var ChatMessagesTable = "ChatMessages"

func PutMessage(msg Message) error {
	svc := Client()
	av, err := dynamodbattribute.MarshalMap(msg)
	if err != nil {
		return err
	}

	input := &dynamodb.PutItemInput{
		TableName: &ChatMessagesTable,
		Item:      av,
	}

	_, err = svc.PutItem(input)
	return err
}

func GetAllMessages() ([]Message, error) {
	var messages []Message

	scanInput := &dynamodb.ScanInput{
		TableName: &ChatMessagesTable,
	}

	scanOutput, err := Client().Scan(scanInput)
	if err != nil {
		return nil, err
	}

	for _, item := range scanOutput.Items {
		var message Message
		err := dynamodbattribute.UnmarshalMap(item, &message)
		if err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}

	return messages, nil
}
