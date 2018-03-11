package db

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

type mockDynamoDBClient struct {
	dynamodbiface.DynamoDBAPI
}

type item struct {
	Title string `json:"title"`
}

var awsConfig = AwsSetup{}

func init() {
	awsConfig = AwsSetup{
		Dyn:         &mockDynamoDBClient{},
		TablePrefix: "prefix.",
	}

}

func (m *mockDynamoDBClient) GetItem(input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {

	itemOutput := dynamodb.GetItemOutput{
		Item: map[string]*dynamodb.AttributeValue{
			"Title": &dynamodb.AttributeValue{
				S: aws.String("Test"),
			},
		},
	}
	return &itemOutput, nil
}

func TestGetItem(t *testing.T) {
	i := item{}

	err := awsConfig.GetItemByIdFromTable("123", "table", &i)
	if err != nil {
		t.Error("error in listing item")
	}
	if i.Title != "Test" {
		t.Error("title not set correctly")
	}
}
