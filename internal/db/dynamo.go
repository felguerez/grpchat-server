package db

import (
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type AccessToken struct {
	ExpiresAt    int64  `json:"expiresAt,omitempty"`
	AccessToken  string `json:"accessToken,omitempty"`
	RefreshToken string `json:"refreshToken,omitempty"`
	TokenType    string `json:"tokenType,omitempty"`
	ID           string `json:"id,omitempty" dynamodbav:"ID"`
}

var svc *dynamodb.DynamoDB

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
