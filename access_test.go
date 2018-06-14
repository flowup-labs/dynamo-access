package godynamo

import (
	"github.com/stretchr/testify/suite"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/defaults"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"fmt"
	"testing"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/expression"
)

type AccessSuite struct {
	suite.Suite

	svc    *dynamodb.DynamoDB
	access *DynamoAccess
}

func (t *AccessSuite) SetupSuite() {
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
	t.access = NewDynamoAccess(config, "")
}

func (t *AccessSuite) SetupTest() {

	t.access.DropTables(&aaa{}, &ccc{}, &bbb{}, &ddd{}, )

	if errs := t.access.DropTables(&User{}); len(errs) != 0 {
		fmt.Println(errs)
	}

	t.access.CreateTables(&aaa{}, &ccc{}, &bbb{}, &ddd{})

	if errs := t.access.CreateTables(&User{}); len(errs) != 0 {
		fmt.Println(errs)
	}

}

func (t *AccessSuite) TestReflect() {
	candidates := []struct {
		item         interface{}
		expectedName string
		expectedErr  error
	}{
		{
			item:         &[]aaa{},
			expectedName: "aaa",
			expectedErr:  nil,
		},

		{
			item:         &[]*aaa{},
			expectedName: "aaa",
			expectedErr:  nil,
		},
		{
			item:         &aaa{},
			expectedName: "aaa",
			expectedErr:  nil,
		},
	}

	for _, candidate := range candidates {
		tableName, err := t.access.tableName(candidate.item)
		t.Nil(err)
		t.Equal(candidate.expectedName, *tableName)
	}
}

func (t *AccessSuite) TestCreateItem() {

	a := &aaa{
		Aa: "aaa",
		Ab: "Aab",
		Ac: []bbb{
			{
				Ba: "Bba",
				Bb: []ddd{
					{
						Da: "Dda",
					},
					{
						Da: "Dda",
					},
				},
			},
		},
	}

	if err := t.access.Create(a); err != nil {
		t.Nil(err)
	}
}

func (t *AccessSuite) TestQueryOneItem() {

	a := &aaa{
		Aa: "Aa",
		Ab: "Ab",
		Ac: []bbb{
			{
				Ba: "Ba",
				Bb: []ddd{
					{
						Da: "Da",
					},
					{
						Da: "Da",
					},
				},
			},
		},
	}

	if err := t.access.Create(a); err != nil {
		t.Nil(err)
	}

	// find item
	item := aaa{}
	if err := t.access.GetOneItem(&item, "id", a.Id); err != nil {
		t.Nil(err)
	}

	t.Equal(item.Id, a.Id)
	t.Equal(*a, item)
}

func (t *AccessSuite) TestScanOneItem() {

	a := &aaa{
		Aa: "Aa",
		Ab: "Ab",
		Ac: []bbb{
			{
				Ba: "Ba",
				Bb: []ddd{
					{
						Da: "Da",
					},
					{
						Da: "Da",
					},
				},
			},
		},
	}

	if err := t.access.Create(a); err != nil {
		t.Nil(err)
	}

	// find item
	item := aaa{}
	if err := t.access.ScanByAttribute(&item, "id", a.Id); err != nil {
		t.Nil(err)
	}

	t.Equal(item.Id, a.Id)
	t.Equal(*a, item)
}

func (t *AccessSuite) TestUpdateItem() {

	a := &aaa{
		Aa: "Aa",
		Ab: "Ab",
		Ac: []bbb{
			{
				Ba: "Ba",
				Bb: []ddd{
					{
						Da: "Da",
					},
					{
						Da: "Da",
					},
				},
			},
		},
	}
	if err := t.access.Create(a); err != nil {
		t.Nil(err)
	}

	a.Aa = "AAA"

	if err := t.access.Update(a); err != nil {
		t.Nil(err)
	}

	// find item
	item := aaa{}
	if err := t.access.ScanByAttribute(&item, "id", a.Id); err != nil {
		t.Nil(err)
	}

	t.Equal(item.Id, a.Id)
	t.Equal(*a, item)
}

func (t *AccessSuite) TestScanOneItemByIndex() {
	//create item
	c := &ccc{
		Ca:  "Ca",
		DId: "",
	}

	if err := t.access.Create(c); err != nil {
		t.Nil(err)
	}

	c.DId = "Did"
	if err := t.access.Create(c); err != nil {
		t.Nil(err)
	}

	c.Id = ""
	if err := t.access.Create(c); err != nil {
		t.Nil(err)
	}

	// find item
	items := []ccc{}
	if err := t.access.ScanByAttribute(&items, "dId", c.DId); err != nil {
		t.Nil(err)
	}

	t.Len(items, 2)
}

func (t *AccessSuite) TestScan() {
	//create item
	a := &aaa{
		Aa: "Aa",
		Ac: []bbb{
			{
				Ba: "foo1",
				Bc: []string{
					"bar2",
					"var3",
					"foo4",
				},
			},
			{
				Ba: "bar5",
			},
			{
				Ba: "var6",
			},
		},
		Ad: []string{
			"bar7",
			"var8",
			"foo9",
		},
		Ae: map[string]bbb{
			"bar10": bbb{
				Ba: "foo",
			},
		},
		Af: map[string]string{
			"bar11": "boo",
		},
	}
	if err := t.access.Create(a); err != nil {
		t.Nil(err)
	}

	// find item
	item := aaa{}
	if err := t.access.ScanByAttribute(&item, "aac[0].bba", "foo1"); err != nil {
		t.Nil(err)
	}

	t.Equal(a, &item)

	// find item
	item = aaa{}
	if err := t.access.ScanCustom(&item, expression.Name("aad").Contains("var8")); err != nil {
		t.Nil(err)
	}

	t.Equal(a, &item)

	// find item
	item = aaa{}
	if err := t.access.ScanCustom(&item, expression.Name("aae.bar10").AttributeExists()); err != nil {
		t.Nil(err)
	}
	t.Equal(a, &item)

	// find item
	item = aaa{}
	if err := t.access.ScanCustom(&item, expression.Name("aaf.bar11").AttributeExists()); err != nil {
		t.Nil(err)
	}
	t.Equal(a, &item)

	////// find item
	//item = aaa{}
	//if err := t.access.ScanCustom(&item, expression.Name("aac.bba").Contains("foo1")); err != nil {
	//	t.Nil(err)
	//}
	//t.Equal(a, &item)
}

func (t *AccessSuite) TestCreateRelationship() {

	a := aaa{
		Aa: "Aa",
		Ab: "Ab",
	}

	t.Nil(t.access.Create(&a))

	b := bbb{
		Ba: "Ba",
		Bb: []ddd{
			{
				Da: "Da",
			},
			{
				Da: "Da",
			},
		},
		CId: a.Id,
	}

	err := t.access.Create(&b)
	t.Nil(err)

	item := bbb{}
	if err := t.access.ScanByAttribute(&item, "cId", a.Id); err != nil {
		t.Nil(err)
	}

	t.Equal(b, item)
}

func (t *AccessSuite) TestScanRange() {

	bs := []bbb{
		{
			Ba: "Ba",
			Bd: 1,
		},
		{
			Ba: "Ba",
			Bd: 5,
		},
		{
			Ba: "Ba",
			Bd: 10,
		},
		{
			Ba: "Ba",
			Bd: 15,
		},
		{
			Ba: "Ba",
			Bd: 20,
		},
	}

	for _, b := range bs {
		t.Nil(t.access.Create(&b))
	}

	items := []bbb{}

	if err := t.access.ScanCustom(&items, expression.Name("bbd").Between(expression.Value(5), expression.Value(16))); err != nil {
		t.Nil(err)
	}

	t.Len(items, 3)
}

func (t *AccessSuite) TestDeleteItem() {

	bs := []bbb{
		{
			Ba: "Ba",
			Bd: 1,
		},
		{
			Ba: "Ba",
			Bd: 5,
		},
		{
			Ba: "Ba",
			Bd: 10,
		},
		{
			Ba: "Ba",
			Bd: 15,
		},
		{
			Ba: "Ba",
			Bd: 20,
		},
	}

	for _, b := range bs {
		t.Nil(t.access.Create(&b))

		if err := t.access.Delete(&b, "id", b.Id); err != nil {
			t.Nil(err)
		}

		err := t.access.GetOneItem(&b, "id", b.Id)
		t.Equal(err.Error(), ErrNotFound.Error())
	}

	items := []bbb{}

	if err := t.access.ScanCustom(&items, expression.Name("bbd").Between(expression.Value(0), expression.Value(255))); err != nil {
		t.Nil(err)
	}

	t.Len(items, 0)
}

func (t *AccessSuite) TestGetNoItem() {

	item := bbb{}

	err := t.access.GetOneItem(&item, "id", "aaa")
	t.Equal(err.Error(), ErrNotFound.Error())
}

func (t *AccessSuite) TestTableBuilder() {
	tableExpected := &dynamodb.CreateTableInput{
		AttributeDefinitions: []dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("first_name"),
				AttributeType: dynamodb.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("email"),
				AttributeType: dynamodb.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("created_at"),
				AttributeType: dynamodb.ScalarAttributeTypeN,
			},
		},
		KeySchema: []dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("email"),
				KeyType:       dynamodb.KeyTypeHash,
			},
		},
		GlobalSecondaryIndexes: []dynamodb.GlobalSecondaryIndex{
			{
				IndexName: aws.String("created_at_first_name_index"),
				KeySchema: []dynamodb.KeySchemaElement{
					{
						AttributeName: aws.String("created_at"),
						KeyType:       dynamodb.KeyTypeHash,
					},
					{
						AttributeName: aws.String("first_name"),
						KeyType:       dynamodb.KeyTypeRange,
					},
				},
				ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
					ReadCapacityUnits:  aws.Int64(10),
					WriteCapacityUnits: aws.Int64(10),
				},
				Projection: &dynamodb.Projection{
					ProjectionType: dynamodb.ProjectionTypeAll,
				},
			},
		},
	}

	output := &dynamodb.CreateTableInput{}
	t.access.tableBuilder(&User{}, output)
	t.Equal(tableExpected, output)
}

func (t *AccessSuite) TestQueryCustom() {

	candidates := []*User{
		{
			FirstName: "John",
			Email:     "john@gmail.com",
			CreatedAt: 1,
		},
		{
			FirstName: "John",
			Email:     "john1@gmail.com",
			CreatedAt: 1,
		},
		{
			FirstName: "John",
			Email:     "john2@gmail.com",
			CreatedAt: 1,
		},
	}

	for _, candidate := range candidates {
		if err := t.access.Create(candidate); err != nil {
			t.Nil(err)
		}
	}

	expr, err := expression.NewBuilder().
		WithKeyCondition(expression.Key("email").Equal(expression.Value("john@gmail.com"))).
		Build()
	if err != nil {
		t.Nil(err)
	}

	user := &User{}

	if err := t.access.QueryCustom(user, expr, "", 0, map[string]dynamodb.AttributeValue{}); err != nil {
		t.Nil(err)
	}

	t.Equal(user, candidates[0])

	expr, err = expression.NewBuilder().
		WithKeyCondition(expression.Key("created_at").Equal(expression.Value(1)).And(expression.Key("first_name").Equal(expression.Value("John")))).
		Build()
	if err != nil {
		t.Nil(err)
	}

	users := []User{}

	if err := t.access.QueryCustom(&users, expr, "created_at_first_name_index", 2, map[string]dynamodb.AttributeValue{
		"email":      dynamodb.AttributeValue{S: aws.String("john@gmail.com")},
		"created_at": dynamodb.AttributeValue{N: aws.String("1")},
		"first_name": dynamodb.AttributeValue{S: aws.String("John")},
	}); err != nil {
		t.Nil(err)
	}

	t.Len(users, 2)

}

func (t *AccessSuite) TestQueryByAttribute() {

	candidates := []*bbb{
		{
			Ba: "Ba",
			Bd: 1,
		},
	}

	for _, candidate := range candidates {
		if err := t.access.Create(candidate); err != nil {
			t.Nil(err)
		}
	}

	b := &bbb{}

	if err := t.access.QueryByAttribute(b, "id", candidates[0].Id); err != nil {
		t.Nil(err)
	}

	t.Equal(b, candidates[0])
}

func TestAccessSuite(t *testing.T) {
	suite.Run(t, &AccessSuite{})
}
