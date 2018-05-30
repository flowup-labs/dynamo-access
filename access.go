package godynamo

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/go-errors/errors"
	"github.com/satori/go.uuid"
	"time"
	"reflect"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/expression"
	"strings"
	"fmt"
)

type DynamoAccess struct {
	svc         *dynamodb.DynamoDB
	tablePrefix string
}

func NewDynamoAccess(config aws.Config, tablePrefix string) *DynamoAccess {
	return &DynamoAccess{svc: dynamodb.New(config), tablePrefix: tablePrefix}
}

var (
	ErrNotPointer = errors.New("item must be pointer")
	ErrElemNil    = errors.New("elem is nil")
)

func (a *DynamoAccess) tagOfFields(item interface{}, attributeDefinitions []dynamodb.AttributeDefinition, keySchemaElement []dynamodb.KeySchemaElement) ([]dynamodb.AttributeDefinition, []dynamodb.KeySchemaElement) {

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
			a, k := a.tagOfFields(t.Field(i).Type, attributeDefinitions, keySchemaElement)
			attributeDefinitions = append(attributeDefinitions, a...)
			keySchemaElement = append(keySchemaElement, k...)
		}

		dynamoTag, ok := t.Field(i).Tag.Lookup("godynamo")
		if !ok {
			continue
		}

		jsonTag, ok := t.Field(i).Tag.Lookup("json")
		if !ok {
			continue
		}

		dynamoTags := strings.Split(dynamoTag, ",")

		atribute := dynamodb.AttributeDefinition{
			AttributeName: aws.String(jsonTag),
		}

		if dynamoTags[0] == "S" {
			atribute.AttributeType = dynamodb.ScalarAttributeTypeS
		}
		if dynamoTags[0] == "N" {
			atribute.AttributeType = dynamodb.ScalarAttributeTypeN
		}

		attributeDefinitions = append(attributeDefinitions, atribute)

		if dynamoTags[1] == "hash" {
			keySchemaElement = append(keySchemaElement, dynamodb.KeySchemaElement{
				AttributeName: aws.String(jsonTag),
				KeyType:       dynamodb.KeyTypeHash,
			})
		}

		if dynamoTags[1] == "range" {
			keySchemaElement = append(keySchemaElement, dynamodb.KeySchemaElement{
				AttributeName: aws.String(jsonTag),
				KeyType:       dynamodb.KeyTypeRange,
			})
		}
	}

	return attributeDefinitions, keySchemaElement
}

func (a *DynamoAccess) CreateTables(items ...interface{}) error {

	for _, item := range items {
		attributeDefinitions, keySchemaElement := a.tagOfFields(item, []dynamodb.AttributeDefinition{}, []dynamodb.KeySchemaElement{})

		tableName, err := a.reflect(item)
		if err != nil {
			return err
		}

		// Send the request, and get the response or error back
		if _, err = a.svc.CreateTableRequest(&dynamodb.CreateTableInput{
			TableName:            aws.String(tableName),
			AttributeDefinitions: attributeDefinitions,
			KeySchema:            keySchemaElement,
			ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
				ReadCapacityUnits:  aws.Int64(10),
				WriteCapacityUnits: aws.Int64(10),
			},
		}).Send(); err != nil {
			return err
		}
	}

	return nil
}

func (a *DynamoAccess) DropTables(items ...interface{}) error {
	for _, item := range items {
		tableName, err := a.reflect(item)
		if err != nil {
			return err
		}

		if _, err := a.svc.DeleteTableRequest(&dynamodb.DeleteTableInput{
			TableName: aws.String(tableName),
		}).Send(); err != nil {
			return err
		}
	}

	return nil
}

// Create, given item si created in db, with new id
func (a *DynamoAccess) Create(item interface{}) (error) {
	tableName, err := a.reflect(item)
	if err != nil {
		return err
	}

	av, err := dynamodbattribute.MarshalMap(item)
	if err != nil {
		return err
	}

	// add uuid
	av["id"] = dynamodb.AttributeValue{
		S: aws.String(uuid.NewV4().String()),
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
		TableName: aws.String(tableName),
	}).Send(); err != nil {
		return err
	}

	return dynamodbattribute.UnmarshalMap(av, item)
}

// Update, given item is updated
func (a *DynamoAccess) Update(item interface{}) (error) {
	tableName, err := a.reflect(item)
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
		TableName: aws.String(tableName),
	}).Send(); err != nil {
		return err
	}

	return dynamodbattribute.UnmarshalMap(av, item)
}

func (a *DynamoAccess) nameOfFields(item interface{}, names []expression.NameBuilder) []expression.NameBuilder {

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
			names = append(names, a.nameOfFields(t.Field(i).Type, names)...)
		}

		if tag, ok := t.Field(i).Tag.Lookup("json"); ok {
			names = append(names, expression.Name(tag))
		}
	}

	return names
}

// QueryByAttribute, find item by attribute
func (a *DynamoAccess) QueryByCustom(item interface{}) error {
	tableName, err := a.reflect(item)
	if err != nil {
		return err
	}

	nameBuilder := []expression.NameBuilder{}
	nameBuilder = a.nameOfFields(item, nameBuilder)

	proj := expression.ProjectionBuilder{}

	for _, name := range nameBuilder {
		proj = proj.AddNames(name)
	}

	expr, err := expression.NewBuilder().
		WithKeyCondition(expression.Key("bbd").GreaterThanEqual(expression.Value(2))).
		WithProjection(proj).
		Build()
	if err != nil {
		return err
	}

	result, err := a.svc.QueryRequest(&dynamodb.QueryInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(tableName),
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

// QueryByAttribute, find item by attribute
func (a *DynamoAccess) QueryByAttribute(item interface{}, key, value string) error {
	tableName, err := a.reflect(item)
	if err != nil {
		return err
	}

	nameBuilder := []expression.NameBuilder{}
	nameBuilder = a.nameOfFields(item, nameBuilder)

	proj := expression.ProjectionBuilder{}

	for _, name := range nameBuilder {
		proj = proj.AddNames(name)
	}

	expr, err := expression.NewBuilder().
		WithKeyCondition(expression.Key(key).Equal(expression.Value(value))).
		WithProjection(proj).
		Build()
	if err != nil {
		return err
	}

	result, err := a.svc.QueryRequest(&dynamodb.QueryInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(tableName),
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

// GetItem, find item by attribute
func (a *DynamoAccess) GetOneItem(item interface{}, key, value string) error {
	tableName, err := a.reflect(item)
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

	nameBuilder := []expression.NameBuilder{}
	nameBuilder = a.nameOfFields(item, nameBuilder)

	proj := expression.ProjectionBuilder{}

	for _, name := range nameBuilder {
		proj = proj.AddNames(name)
	}

	expr, err := expression.NewBuilder().
		WithKeyCondition(expression.Key(key).Equal(expression.Value(value))).
		WithProjection(proj).
		Build()
	if err != nil {
		return err
	}

	result, err := a.svc.GetItemRequest(&dynamodb.GetItemInput{
		ExpressionAttributeNames: expr.Names(),
		ProjectionExpression:     expr.Projection(),
		TableName:                aws.String(tableName),
		Key: map[string]dynamodb.AttributeValue{
			key: dynamodb.AttributeValue{S: aws.String(value)},
		},
	}).Send()
	if err != nil {
		return err
	}

	if err := dynamodbattribute.UnmarshalMap(result.Item, item); err != nil {
		return err
	}

	return nil
}

// ScanByAttribute, find item by attribute
func (a *DynamoAccess) ScanByAttribute(item interface{}, key, value string) error {
	return a.ScanCustom(item, expression.Name(key).Equal(expression.Value(value)))
}

func (a *DynamoAccess) ScanCustom(item interface{}, filt expression.ConditionBuilder) error {
	tableName, err := a.reflect(item)
	if err != nil {
		return err
	}

	nameBuilder := []expression.NameBuilder{}
	nameBuilder = a.nameOfFields(item, nameBuilder)

	proj := expression.ProjectionBuilder{}

	for _, name := range nameBuilder {
		proj = proj.AddNames(name)
	}

	expr, err := expression.NewBuilder().
		WithFilter(filt).
		WithProjection(proj).
		Build()
	if err != nil {
		return err
	}

	result, err := a.svc.ScanRequest(&dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(tableName),
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

func (a *DynamoAccess) reflect(item interface{}) (string, error) {
	t := reflect.TypeOf(item)

	if t.Kind() != reflect.Ptr {
		return "", ErrNotPointer
	}

	if t.Elem() == nil {
		return "", ErrElemNil
	}

	for t.Kind() == reflect.Ptr || t.Kind() == reflect.Slice {
		t = t.Elem()
	}

	return a.tablePrefix + t.Name(), nil
}
