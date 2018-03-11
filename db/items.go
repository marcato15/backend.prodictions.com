package db

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

// Key A struct for storing a field and value for lookup
type Key struct {
	Field string
	Value string
}

func DecodeKey(encodedKey string) (Key, error) {
	key := Key{}
	decodedKey, err := base64.StdEncoding.DecodeString(encodedKey)
	if err != nil {
		return key, err
	}
	err = json.Unmarshal(decodedKey, &key)
	if err != nil {
		return key, err
	}
	return key, nil
}

func DecodeKeys(encodedKey string) ([]Key, error) {
	keys := []Key{}
	decodedKey, err := base64.StdEncoding.DecodeString(encodedKey)
	if err != nil {
		return keys, err
	}
	err = json.Unmarshal(decodedKey, &keys)
	if err != nil {
		return keys, err
	}
	return keys, nil
}

func EncodeKeys(keys []Key) (string, error) {
	bytes, err := json.Marshal(keys)
	if err != nil {
		return "", err
	}
	encodedKeys := base64.StdEncoding.EncodeToString(bytes)
	return encodedKeys, nil
}

func EncodeKey(key Key) (string, error) {
	bytes, err := json.Marshal(key)
	if err != nil {
		return "", err
	}
	encodedKey := base64.StdEncoding.EncodeToString(bytes)
	return encodedKey, nil
}

func CustomMarshalKeepEmptyLists(in interface{}, emptyListNames []string, av *dynamodb.AttributeValue) error {
	//To Make this easy, we are going to marshal to json then use the built in defult to marshal to dynamo. This is b/c if we ran it on the object it'd cause an infinite recursion.
	j, _ := json.Marshal(in)
	mapping := map[string]interface{}{}
	json.Unmarshal(j, &mapping)

	data, err := dynamodbattribute.MarshalMap(mapping)
	if err != nil {
		return err
	}

	//If list is null, set a custom empty List value, so things can be added
	for _, listName := range emptyListNames {
		if mapping[listName] == nil {
			emptyList := []*dynamodb.AttributeValue{}
			list := dynamodb.AttributeValue{
				L: emptyList,
			}
			data[listName] = &list
		}
	}
	av.M = data
	return nil
}

//GetItemByKeyFromTable Helper Function to Take a single key and prep for the array of keys needed
func (awsConnector AwsSetup) GetItemByKeyFromTable(key Key, tableName string, out interface{}) error {
	keys := []Key{}
	keys = append(keys, key)
	return awsConnector.GetItemByKeysFromTable(keys, tableName, out)

}

//GetItemByKeysFromTable Retrieves Items by Key(s)
func (awsConnector AwsSetup) GetItemByKeysFromTable(keys []Key, tableName string, out interface{}) error {

	table := awsConnector.TablePrefix + tableName

	keysAv, err := dynamodbattribute.MarshalMap(keys)
	if err != nil {
		fmt.Println("Error Marshalling keys")
		fmt.Println(err)
		return err
	}
	params := &dynamodb.GetItemInput{
		Key:       keysAv,
		TableName: aws.String(table),
	}
	result, err := awsConnector.Dyn.GetItem(params)

	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	marshalErr := dynamodbattribute.UnmarshalMap(result.Item, out)
	if err != nil {
		fmt.Println(marshalErr)
		return nil
	}
	return nil
}

func (awsConnector AwsSetup) GetItemsByKeyFromTable(out []interface{}, key Key, tableName string) error {

	table := awsConnector.TablePrefix + tableName

	input := &dynamodb.QueryInput{
		ExpressionAttributeNames: map[string]*string{
			"#field": aws.String(key.Field),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":value": {
				S: aws.String(key.Value),
			},
		},
		KeyConditionExpression: aws.String("#field = :value"),
		TableName:              aws.String(table),
	}

	result, err := awsConnector.Dyn.Query(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case dynamodb.ErrCodeProvisionedThroughputExceededException:
				fmt.Println(dynamodb.ErrCodeProvisionedThroughputExceededException, aerr.Error())
			case dynamodb.ErrCodeResourceNotFoundException:
				fmt.Println(dynamodb.ErrCodeResourceNotFoundException, aerr.Error())
			case dynamodb.ErrCodeInternalServerError:
				fmt.Println(dynamodb.ErrCodeInternalServerError, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
			return err
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
			return err
		}
	}

	marshalErr := dynamodbattribute.UnmarshalListOfMaps(result.Items, out)
	if err != nil {
		fmt.Println(marshalErr)
		return err
	}
	return nil

}

func MarshalKeys(keys []Key) (map[string]*dynamodb.AttributeValue, error) {
	keyOut := map[string]*dynamodb.AttributeValue{}
	for _, key := range keys {
		val, err := dynamodbattribute.Marshal(key.Value)
		keyOut[key.Field] = val
		if err != nil {
			return keyOut, err
		}
	}
	return keyOut, nil
}

func (awsConnector AwsSetup) AddItemToListByKeysInTable(item interface{}, listName string, keys []Key, tableName string) error {

	table := awsConnector.TablePrefix + tableName

	keysAv, err := MarshalKeys(keys)
	if err != nil {
		fmt.Println("Error Marshalling keys")
		fmt.Println(err)
		return err
	} else {
		fmt.Println("keysAv")
		fmt.Println(keysAv)
	}

	dynItem, err := dynamodbattribute.Marshal(item)
	if err != nil {
		fmt.Println("Error Marshalling Items")
		fmt.Println(err)
		return err
	}
	dynItems := []*dynamodb.AttributeValue{}
	dynItems = append(dynItems, dynItem)

	dynItemList := &dynamodb.AttributeValue{
		L: dynItems,
	}
	attributeValues := map[string]*dynamodb.AttributeValue{
		":values": dynItemList,
	}
	attributeNames := map[string]*string{
		"#field": aws.String(listName),
	}
	params := &dynamodb.UpdateItemInput{
		Key:                       keysAv,
		ReturnValues:              aws.String("ALL_NEW"),
		ExpressionAttributeNames:  attributeNames,
		ExpressionAttributeValues: attributeValues,
		UpdateExpression:          aws.String("SET #field = list_append(#field, :values)"),
		TableName:                 aws.String(table),
	}

	_, err = awsConnector.Dyn.UpdateItem(params)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case dynamodb.ErrCodeConditionalCheckFailedException:
				fmt.Println(dynamodb.ErrCodeConditionalCheckFailedException, aerr.Error())
			case dynamodb.ErrCodeProvisionedThroughputExceededException:
				fmt.Println(dynamodb.ErrCodeProvisionedThroughputExceededException, aerr.Error())
			case dynamodb.ErrCodeResourceNotFoundException:
				fmt.Println(dynamodb.ErrCodeResourceNotFoundException, aerr.Error())
			case dynamodb.ErrCodeItemCollectionSizeLimitExceededException:
				fmt.Println(dynamodb.ErrCodeItemCollectionSizeLimitExceededException, aerr.Error())
			case dynamodb.ErrCodeInternalServerError:
				fmt.Println(dynamodb.ErrCodeInternalServerError, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return err
	}

	return nil
}

func (awsConnector AwsSetup) PutItemInTable(item interface{}, tableName string) (*dynamodb.PutItemOutput, error) {

	table := aws.String(awsConnector.TablePrefix + tableName)

	dynItem, err := dynamodbattribute.MarshalMap(item)
	if err != nil {
		panic(fmt.Sprintf("failed to DynamoDB marshal Record, %v", err))
	}
	result, err := awsConnector.Dyn.PutItem(&dynamodb.PutItemInput{
		TableName: table,
		Item:      dynItem,
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case dynamodb.ErrCodeConditionalCheckFailedException:
				fmt.Println(dynamodb.ErrCodeConditionalCheckFailedException, aerr.Error())
				return nil, err
			case dynamodb.ErrCodeProvisionedThroughputExceededException:
				fmt.Println(dynamodb.ErrCodeProvisionedThroughputExceededException, aerr.Error())
				return nil, err
			case dynamodb.ErrCodeResourceNotFoundException:
				fmt.Println(dynamodb.ErrCodeResourceNotFoundException, aerr.Error())
				return nil, err
			case dynamodb.ErrCodeItemCollectionSizeLimitExceededException:
				fmt.Println(dynamodb.ErrCodeItemCollectionSizeLimitExceededException, aerr.Error())
				return nil, err
			case dynamodb.ErrCodeInternalServerError:
				fmt.Println(dynamodb.ErrCodeInternalServerError, aerr.Error())
				return nil, err
			default:
				fmt.Println(aerr.Error())
				return nil, err
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
			return nil, err
		}
	}
	return result, nil
}

func (awsConnector AwsSetup) ListItems(tableName string, out interface{}) error {
	table := aws.String(awsConnector.TablePrefix + tableName)
	input := &dynamodb.ScanInput{
		TableName: table,
	}
	result, err := awsConnector.Dyn.Scan(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			fmt.Println(aerr.Error())
		} else {
			fmt.Println(err.Error())
		}
		return err
	}
	marshalErr := dynamodbattribute.UnmarshalListOfMaps(result.Items, out)
	if err != nil {
		fmt.Println(marshalErr)
		return err
	}
	return nil
}
