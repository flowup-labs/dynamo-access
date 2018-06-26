package godynamo

import (
	"github.com/stretchr/testify/suite"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/defaults"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"testing"
)

type MigrationSuite struct {
	suite.Suite

	svc    *dynamodb.DynamoDB
	access *DynamoAccess
}

func (t *MigrationSuite) SetupSuite() {
	config := defaults.Config()
	config.Region = "mock-region"
	config.EndpointResolver = aws.ResolveWithEndpointURL("http://dynamodb:8000")
	config.Credentials = aws.StaticCredentialsProvider{
		Value: aws.Credentials{
			AccessKeyID: "AKID", SecretAccessKey: "SECRET", SessionToken: "SESSION",
			Source:      "unit test credentials",
		},
	}

	t.svc = dynamodb.New(config)
	t.access = NewDynamoAccess(config, "migration_")
}

func (t *MigrationSuite) SetupTest() {

	t.access.DropTables(&aaa{}, &ccc{}, &bbb{}, &ddd{}, &fff{}, &eee{}, &user{})

	t.access.CreateTables(&aaa{}, &ccc{}, &bbb{}, &ddd{}, &fff{}, &eee{}, &user{})
}

func (t *MigrationSuite) TestMigration() {
	candidates := []*aaa{
		{
			Aa: "1",
		},
		{
			Aa: "2",
		},
		{
			Aa: "3",
		},
	}

	for _, candidate := range candidates {
		t.access.Create(candidate)
	}

	data, err := t.access.DumpTable(&aaa{})
	t.Nil(err)

	//if err = t.access.WriteStringToFile(string(data), "./backup"); err != nil {
	//	t.Nil(err)
	//}
	//
	//data, err = t.access.OpenFile("./backup")
	//if err != nil {
	//	t.Nil(err)
	//}

	model := []aaa{}

	if err := t.access.Bind(&model, data); err != nil {
		t.Nil(err)
	}

	t.Len(model, 3)

}

func TestMigrationSuite(t *testing.T) {
	suite.Run(t, &MigrationSuite{})
}
