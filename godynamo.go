package godynamo

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/aws"
	"errors"
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
	ErrSlice            = errors.New("slice is prohibited")

	NoPaging = map[string]dynamodb.AttributeValue{}
)
