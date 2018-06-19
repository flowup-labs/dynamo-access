package godynamo

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// Represents basic model data as id, and
// times of created, updated and deleted saved
// in unix format
type Model struct {
	Id      string `json:"id" godynamo:"hash"`
	Created int64  `json:"created"`
	Updated int64  `json:"updated"`
	Deleted int64  `json:"deleted"`
}

// Represents the input of a Query operation.
type RequestInput struct {
	// Expression represents a collection of DynamoDB Expressions. The getter
	// methods of the Expression struct retrieves the formatted DynamoDB
	// Expressions, ExpressionAttributeNames, and ExpressionAttributeValues.
	//
	// Example:
	//
	//     // keyCond represents the Key Condition Expression
	//     keyCond := expression.Key("someKey").Equal(expression.Value("someValue"))
	//     // proj represents the Projection Expression
	//     proj := expression.NamesList(expression.Name("aName"), expression.Name("anotherName"), expression.Name("oneOtherName"))
	//
	//     // Add keyCond and proj to builder as a Key Condition and Projection
	//     // respectively
	//     builder := expression.NewBuilder().WithKeyCondition(keyCond).WithProjection(proj)
	//     expression := builder.Build()
	//
	//     queryInput := dynamodb.QueryInput{
	//       KeyConditionExpression:    expression.KeyCondition(),
	//       ProjectionExpression:      expression.Projection(),
	//       ExpressionAttributeNames:  expression.Names(),
	//       ExpressionAttributeValues: expression.Values(),
	//       TableName: aws.String("SomeTable"),
	//     }
	Expr expression.Expression

	// The name of an index to query. This index can be any local secondary index
	// or global secondary index on the table. Note that if you use the IndexName
	// parameter, you must also provide TableName.
	IndexName string

	// The maximum number of items to evaluate (not necessarily the number of matching
	// items). If DynamoDB processes the number of items up to the limit while processing
	// the results, it stops the operation and returns the matching values up to
	// that point, and a key in LastEvaluatedKey to apply in a subsequent operation,
	// so that you can pick up where you left off. Also, if the processed data set
	// size exceeds 1 MB before DynamoDB reaches this limit, it stops the operation
	// and returns the matching values up to the limit, and a key in LastEvaluatedKey
	// to apply in a subsequent operation to continue the operation. For more information,
	// see Query and Scan (http://docs.aws.amazon.com/amazondynamodb/latest/developerguide/QueryAndScan.html)
	// in the Amazon DynamoDB Developer Guide.
	Limit int64

	// The primary key of the first item that this operation will evaluate. Use
	// the value that was returned for LastEvaluatedKey in the previous operation.
	//
	// The data type for ExclusiveStartKey must be String, Number or Binary. No
	// set data types are allowed.
	ExclusiveStartKey map[string]dynamodb.AttributeValue

	// Specifies the order for index traversal: If true (default), the traversal
	// is performed in ascending order; if false, the traversal is performed in
	// descending order.
	//
	// Items with the same partition key value are stored in sorted order by sort
	// key. If the sort key data type is Number, the results are stored in numeric
	// order. For type String, the results are stored in order of UTF-8 bytes. For
	// type Binary, DynamoDB treats each byte of the binary data as unsigned.
	//
	// If ScanIndexForward is true, DynamoDB returns the results in the order in
	// which they are stored (by sort key value). This is the default behavior.
	// If ScanIndexForward is false, DynamoDB reads the results in reverse order
	// by sort key value, and then returns the results to the client.
	ScanIndexForward bool
}

type aaa struct {
	Model

	Aa string            `json:"aaa"`
	Ab string            `json:"aab"`
	Ac []bbb             `json:"aac"`
	Ad []string          `json:"aad"`
	Ae map[string]bbb    `json:"aae"`
	Af map[string]string `json:"aaf"`
}

type bbb struct {
	Model

	Ba string   `json:"bba"`
	Bb []ddd    `json:"bbb"`
	Bc []string `json:"bbc"`

	CId string `json:"cId"`

	Bd int64 `json:"bbd"`
}

type ccc struct {
	Model

	Ca string `json:"cca"`

	DId string `json:"dId"`
}

type ddd struct {
	Model

	Da string `json:"dda" godynamo:"global_secondary_index(index:hash)"`
	Db string `json:"ddb"`
}

type fff struct {
	Model

	Fa string `json:"ffa"`
	Fb string `json:"ffb" godynamo:"range"`
	Fc int64  `json:"ffc" godynamo:"local_secondary_index(index:range)"`
}

type eee struct {
	Model

	Ea string `json:"eea" godynamo:"global_secondary_index(index:hash)"`
	Eb int64  `json:"eeb" godynamo:"global_secondary_index(index:range),global_secondary_index(index2:range)"`
	Ec int64  `json:"eec"  godynamo:"global_secondary_index(index2:hash)"`
}

type user struct {
	FirstName string `json:"first_name" godynamo:"global_secondary_index(created_at_first_name_index:range)"`
	LastName  string `json:"last_name"`
	Email     string `json:"email" godynamo:"hash"`
	CreatedAt int64  `json:"created_at" godynamo:"global_secondary_index(created_at_first_name_index:hash)"`
	UpdatedAt int64  `json:"updated_at"`
	AuthenticationTokens []struct {
		Token      string `json:"token"`
		LastUsedAt string `json:"last_used_at"`
	} `json:"authentication_tokens"`
	Addresses []struct {
		City    string `json:"city"`
		State   string `json:"state"`
		Country string `json:"country"`
	} `json:"addresses"`
}
