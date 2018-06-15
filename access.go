package godynamo

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/expression"
	"github.com/satori/go.uuid"
	"reflect"
	"time"
	"strings"
)

type DynamoAccess struct {
	svc         *dynamodb.DynamoDB
	tablePrefix string
}

func NewDynamoAccess(config aws.Config, tablePrefix string) *DynamoAccess {
	return &DynamoAccess{svc: dynamodb.New(config), tablePrefix: tablePrefix}
}

var (
	ErrNotPointer       = errors.New("item must be pointer")
	ErrElemNil          = errors.New("elem is nil")
	ErrNotFound         = errors.New("item not found")
	ErrNotSupportedType = errors.New("not supported type")

	NoPaging = map[string]dynamodb.AttributeValue{}
)

func (a *DynamoAccess) typeToScalarType(Type string) (dynamodb.ScalarAttributeType, error) {
	if Type == "string" {
		return dynamodb.ScalarAttributeTypeS, nil
	}
	if Type == "int" ||
		Type == "int8" ||
		Type == "int16" ||
		Type == "int32" ||
		Type == "int64" ||
		Type == "uint" ||
		Type == "uint8" ||
		Type == "uint16" ||
		Type == "uint32" ||
		Type == "uint64" ||
		Type == "uintptr" {
		return dynamodb.ScalarAttributeTypeN, nil
	}

	return dynamodb.ScalarAttributeTypeS, ErrNotSupportedType
}

func (a *DynamoAccess) tableBuilder(item interface{}, table *dynamodb.CreateTableInput) (error) {
	var err error
	v := reflect.ValueOf(item)
	t := v.Type()

	if v.Elem().Type().String() == "reflect.rtype" {
		t = item.(reflect.Type)
	}

	for t.Kind() == reflect.Ptr || t.Kind() == reflect.Slice {
		t = t.Elem()
	}

	for i := 0; i < t.NumField(); i++ {
		if t.Field(i).Type.Kind() == reflect.Struct && t.Field(i).Name == "Model" {
			if err := a.tableBuilder(t.Field(i).Type, table); err != nil {
				return err
			}
		}

		dynamoTag, ok := t.Field(i).Tag.Lookup("godynamo")
		if !ok {
			continue
		}

		jsonTag, ok := t.Field(i).Tag.Lookup("json")
		if !ok {
			continue
		}

		dynamoFuncs := strings.Split(dynamoTag, ",")
		for _, dynamoFunc := range dynamoFuncs {
			var gsiB, lsiB bool
			index := 0

			attribute := dynamodb.AttributeDefinition{
				AttributeName: aws.String(jsonTag),
			}

			attribute.AttributeType, err = a.typeToScalarType(t.Field(i).Type.String())
			if err != nil {
				return err
			}

			table.AttributeDefinitions = append(table.AttributeDefinitions, attribute)
			keySchema := []dynamodb.KeySchemaElement{}

			if strings.HasPrefix(dynamoFunc, "global_secondary_index(") {
				gsiB = true
				dynamoFunc = strings.TrimSuffix(strings.TrimPrefix(dynamoFunc, "global_secondary_index("), ")")
				dynamoTags := strings.Split(dynamoFunc, ":")

				dynamoFunc = dynamoTags[1]

				for ; index < len(table.GlobalSecondaryIndexes) && *table.GlobalSecondaryIndexes[index].IndexName != dynamoTags[0]; index++ {
				}

				if len(table.GlobalSecondaryIndexes) == index {
					table.GlobalSecondaryIndexes = append(table.GlobalSecondaryIndexes, dynamodb.GlobalSecondaryIndex{
						IndexName: aws.String(dynamoTags[0]),
						ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
							ReadCapacityUnits:  aws.Int64(10),
							WriteCapacityUnits: aws.Int64(10),
						},
						Projection: &dynamodb.Projection{
							ProjectionType: dynamodb.ProjectionTypeAll,
						},
					})
				}
			}

			elem := dynamodb.KeySchemaElement{
				AttributeName: aws.String(jsonTag),
			}

			switch dynamoFunc {
			case "hash":
				elem.KeyType = dynamodb.KeyTypeHash
			case "range":
				elem.KeyType = dynamodb.KeyTypeRange
			}

			keySchema = append(keySchema, elem)

			if gsiB {
				if dynamoFunc == "hash" {
					table.GlobalSecondaryIndexes[index].KeySchema = append(keySchema, table.GlobalSecondaryIndexes[index].KeySchema...)
				} else {
					table.GlobalSecondaryIndexes[index].KeySchema = append(table.GlobalSecondaryIndexes[index].KeySchema, keySchema...)
				}
				continue
			} else if lsiB {
				//todo
				//table.LocalSecondaryIndexes
			}

			table.KeySchema = keySchema
		}
	}

	return nil
}

func (a *DynamoAccess) CreateTables(items ...interface{}) []error {
	var errors []error
	for _, item := range items {
		table := &dynamodb.CreateTableInput{}
		var err error

		table.TableName, err = a.tableName(item)
		if err != nil {
			errors = append(errors, err)
		}

		if err := a.tableBuilder(item, table); err != nil {
			errors = append(errors, err)
		}

		table.ProvisionedThroughput = &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(10),
			WriteCapacityUnits: aws.Int64(10),
		}

		// Send the request, and get the response or error back
		if _, err = a.svc.CreateTableRequest(table).Send();
			err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

func (a *DynamoAccess) DropTables(items ...interface{}) []error {
	var errors []error

	for _, item := range items {
		tableName, err := a.tableName(item)
		if err != nil {
			errors = append(errors, err)
		}

		if _, err := a.svc.DeleteTableRequest(&dynamodb.DeleteTableInput{
			TableName: tableName,
		}).Send(); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

// Create, given item si created in db, with new id
func (a *DynamoAccess) Create(item interface{}) error {
	tableName, err := a.tableName(item)
	if err != nil {
		return err
	}

	av, err := dynamodbattribute.MarshalMap(item)
	if err != nil {
		return err
	}

	// add uuid
	v4, err := uuid.NewV4()
	if err != nil {
		return err
	}
	if av["id"].NULL != nil && *av["id"].NULL {
		av["id"] = dynamodb.AttributeValue{
			S: aws.String(v4.String()),
		}
	}

	// add timestamps
	timeNow := fmt.Sprint(time.Now().Unix())
	av["created"] = dynamodb.AttributeValue{
		N: aws.String(timeNow),
	}
	av["updated"] = dynamodb.AttributeValue{
		N: aws.String(timeNow),
	}

	if _, err := a.svc.PutItemRequest(&dynamodb.PutItemInput{
		Item:      av,
		TableName: tableName,
	}).Send(); err != nil {
		return err
	}

	return dynamodbattribute.UnmarshalMap(av, item)
}

// Update, given item is updated
func (a *DynamoAccess) Update(item interface{}) error {
	tableName, err := a.tableName(item)
	if err != nil {
		return err
	}

	av, err := dynamodbattribute.MarshalMap(item)
	if err != nil {
		return err
	}

	// add timestamp
	av["updated"] = dynamodb.AttributeValue{
		N: aws.String(fmt.Sprint(time.Now().Unix())),
	}

	if _, err := a.svc.PutItemRequest(&dynamodb.PutItemInput{
		Item:      av,
		TableName: tableName,
	}).Send(); err != nil {
		return err
	}

	return dynamodbattribute.UnmarshalMap(av, item)
}

// Delete, given id of item is deleted
func (a *DynamoAccess) Delete(item interface{}, key, value string) error {
	if err := a.GetOneItem(item, key, value); err != nil {
		return err
	}

	tableName, err := a.tableName(item)
	if err != nil {
		return err
	}

	av, err := dynamodbattribute.MarshalMap(item)
	if err != nil {
		return err
	}

	// add timestamp
	av["deleted"] = dynamodb.AttributeValue{
		N: aws.String(fmt.Sprint(time.Now().Unix())),
	}

	if _, err := a.svc.PutItemRequest(&dynamodb.PutItemInput{
		Item:      av,
		TableName: tableName,
	}).Send(); err != nil {
		return err
	}

	return dynamodbattribute.UnmarshalMap(av, item)
}

// QueryByAttribute, find item by attribute
func (a *DynamoAccess) QueryByAttribute(item interface{}, key, value string) error {
	expr, err := expression.NewBuilder().
		WithKeyCondition(expression.Key(key).Equal(expression.Value(value))).
		WithFilter(expression.Name("deleted").Equal(expression.Value(0))).
		Build()
	if err != nil {
		return err
	}

	if err := a.QueryCustom(item, expr, "", 0, map[string]dynamodb.AttributeValue{}); err != nil {
		return err
	}

	return nil
}

//QueryByAttribute, find item by attribute
func (a *DynamoAccess) QueryCustom(item interface{}, expr expression.Expression, indexName string, limit int64, exclusiveStartKey map[string]dynamodb.AttributeValue) error {
	tableName, err := a.tableName(item)
	if err != nil {
		return err
	}

	queryInput := &dynamodb.QueryInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		TableName:                 tableName,
	}

	if expr.Filter() != nil && *expr.Filter() != "" {
		queryInput.FilterExpression = expr.Filter()
	}

	if limit != 0 {
		queryInput.Limit = aws.Int64(limit)
	}

	if indexName != "" {
		queryInput.IndexName = aws.String(indexName)
	}

	if len(exclusiveStartKey) != 0 {
		queryInput.ExclusiveStartKey = exclusiveStartKey
	}

	result, err := a.svc.QueryRequest(queryInput).Send()
	if err != nil {
		return err
	}

	t := reflect.TypeOf(item)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Slice {
		if len(result.Items) > 0 {
			if err := dynamodbattribute.UnmarshalMap(result.Items[0], item); err != nil {
				return err
			}
		}
	} else {
		if err := dynamodbattribute.UnmarshalListOfMaps(result.Items, item); err != nil {
			return err
		}
	}

	return nil
}

// GetItem, find item by attribute
func (a *DynamoAccess) GetOneItem(item interface{}, key, value string) error {
	tableName, err := a.tableName(item)
	if err != nil {
		return err
	}

	t := reflect.TypeOf(item)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() == reflect.Slice {
		return errors.New("slice is prohibited")
	}

	result, err := a.svc.GetItemRequest(&dynamodb.GetItemInput{
		TableName: tableName,
		Key: map[string]dynamodb.AttributeValue{
			key: {S: aws.String(value)},
		},
	}).Send()
	if err != nil {
		return err
	}

	if err := dynamodbattribute.UnmarshalMap(result.Item, item); err != nil {
		return err
	}

	if result.Item["id"].S == nil || (result.Item["deleted"].N != nil && *result.Item["deleted"].N != *aws.String("0")) {
		return ErrNotFound
	}

	return nil
}

// ScanByAttribute, find item by attribute
func (a *DynamoAccess) ScanByAttribute(item interface{}, key, value string) error {
	return a.ScanCustom(item, expression.Name(key).Equal(expression.Value(value)))
}

func (a *DynamoAccess) ScanCustom(item interface{}, filt expression.ConditionBuilder) error {
	tableName, err := a.tableName(item)
	if err != nil {
		return err
	}

	expr, err := expression.NewBuilder().
		WithFilter(filt.And(expression.Name("deleted").Equal(expression.Value(0)))).
		Build()
	if err != nil {
		return err
	}

	result, err := a.svc.ScanRequest(&dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		TableName:                 tableName,
	}).Send()
	if err != nil {
		return err
	}

	t := reflect.TypeOf(item)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Slice {
		if len(result.Items) > 0 {
			if err := dynamodbattribute.UnmarshalMap(result.Items[0], item); err != nil {
				return err
			}
		}
	} else {
		if err := dynamodbattribute.UnmarshalListOfMaps(result.Items, item); err != nil {
			return err
		}
	}

	return nil
}

func (a *DynamoAccess) tableName(item interface{}) (*string, error) {
	t := reflect.TypeOf(item)

	if t.Kind() != reflect.Ptr {
		return aws.String(""), ErrNotPointer
	}

	if t.Elem() == nil {
		return aws.String(""), ErrElemNil
	}

	for t.Kind() == reflect.Ptr || t.Kind() == reflect.Slice {
		t = t.Elem()
	}

	return aws.String(a.tablePrefix + t.Name()), nil
}
