package db

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

// AwsSetup The struct used for storing data about the AWS Call
type AwsSetup struct {
	Dyn         dynamodbiface.DynamoDBAPI
	TablePrefix string
}

// SetupAws The function to setup and return the AWS Object being used
func SetupAws() AwsSetup {
	sess := session.Must(session.NewSession())
	return AwsSetup{
		Dyn:         dynamodb.New(sess),
		TablePrefix: "prodictions.",
	}
}
