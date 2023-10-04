package main

import (
	"fmt"
	"log"

	"github.com/AlecAivazis/survey/v2"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func main() {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})

	if err != nil {
		log.Fatalf("Failed to connect to AWS: %s", err)
	}

	svc := dynamodb.New(sess)

	var tableName string
	prompt := &survey.Input{
		Message: "Enter table name:",
	}
	survey.AskOne(prompt, &tableName)

	// Check if table exists
	input := &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	}
	_, err = svc.DescribeTable(input)

	if err == nil {
		var shouldDelete bool
		prompt := &survey.Confirm{
			Message: "Table already exists. Do you want to delete it?",
		}
		survey.AskOne(prompt, &shouldDelete)

		if shouldDelete {
			// Delete table
			_, err := svc.DeleteTable(&dynamodb.DeleteTableInput{
				TableName: aws.String(tableName),
			})
			if err != nil {
				log.Fatalf("Failed to delete table: %s", err)
			}
			fmt.Println("Table deleted successfully.")
			return
		}
	}

	attributeDefinitions := []*dynamodb.AttributeDefinition{}
	keySchema := []*dynamodb.KeySchemaElement{}

	for {
		var attributeName, attributeType string

		prompt := &survey.Input{
			Message: "Enter name for attribute:",
		}
		survey.AskOne(prompt, &attributeName)

		options := []string{"S", "N", "B"}
		promptSelect := &survey.Select{
			Message: fmt.Sprintf("Choose a type for attribute %s:", attributeName),
			Options: options,
		}
		survey.AskOne(promptSelect, &attributeType)

		attributeDefinitions = append(attributeDefinitions, &dynamodb.AttributeDefinition{
			AttributeName: aws.String(attributeName),
			AttributeType: aws.String(attributeType),
		})

		if len(keySchema) == 0 {
			keySchema = append(keySchema, &dynamodb.KeySchemaElement{
				AttributeName: aws.String(attributeName),
				KeyType:       aws.String("HASH"),
			})
		}

		var shouldContinue bool
		confirmPrompt := &survey.Confirm{
			Message: "Do you want to add another attribute?",
		}
		survey.AskOne(confirmPrompt, &shouldContinue)

		if !shouldContinue {
			break
		}
	}

	// Create table
	createInput := &dynamodb.CreateTableInput{
		AttributeDefinitions:  attributeDefinitions,
		KeySchema:             keySchema,
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{ReadCapacityUnits: aws.Int64(5), WriteCapacityUnits: aws.Int64(5)},
		TableName:             aws.String(tableName),
	}

	_, err = svc.CreateTable(createInput)
	if err != nil {
		log.Fatalf("Failed to create table: %s", err)
	}

	fmt.Println("Table created successfully.")
}
