package godynamo

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"encoding/json"
	"os"
	"io/ioutil"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/dynamodbattribute"
)

func (a *DynamoAccess) DumpTable(table interface{}) ([]byte, error) {

	tableName, _, err := a.tableName(table)
	if err != nil {
		return []byte{}, err
	}

	scanInput := &dynamodb.ScanInput{
		TableName: tableName,
	}
	result, err := a.svc.ScanRequest(scanInput).Send()
	if err != nil {
		return []byte{}, err
	}

	return json.Marshal(result.Items)
}

func (a *DynamoAccess) WriteStringToFile(data string, path string) (error) {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(string(data))
	if err != nil {
		return err
	}

	err = file.Sync()
	if err != nil {
		return err
	}
	return nil
}

func (a *DynamoAccess) OpenFile(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

func (a *DynamoAccess) Bind(item interface{}, data []byte) error {
	attributeValues := []map[string]dynamodb.AttributeValue{}

	if err := json.Unmarshal(data, &attributeValues); err != nil {
		return err
	}

	if err := dynamodbattribute.UnmarshalListOfMaps(attributeValues, item); err != nil {
		return err
	}

	return nil
}
