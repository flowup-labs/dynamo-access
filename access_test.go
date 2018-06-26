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
	t.access = NewDynamoAccess(config, "access_")
}

func (t *AccessSuite) SetupTest() {

	t.access.DropTables(&aaa{}, &ccc{}, &bbb{}, &ddd{}, &fff{}, &eee{}, &user{})

	t.access.CreateTables(&aaa{}, &ccc{}, &bbb{}, &ddd{}, &fff{}, &eee{}, &user{})
}

func (t *AccessSuite) TestReflect() {
	candidates := []struct {
		item         interface{}
		expectedName string
		expectedErr  error
	}{
		{
			item:         &[]aaa{},
			expectedName: t.access.tablePrefix + "aaa",
			expectedErr:  nil,
		},

		{
			item:         &[]*aaa{},
			expectedName: t.access.tablePrefix + "aaa",
			expectedErr:  nil,
		},
		{
			item:         &aaa{},
			expectedName: t.access.tablePrefix + "aaa",
			expectedErr:  nil,
		},
	}

	for _, candidate := range candidates {
		tableName, _, err := t.access.tableName(candidate.item)
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
	if err := t.access.GetItem(&item, "id", a.Id); err != nil {
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
	if err := t.access.ScanByFilter(&item, expression.Name("aad").Contains("var8")); err != nil {
		t.Nil(err)
	}

	t.Equal(a, &item)

	// find item
	item = aaa{}
	if err := t.access.ScanByFilter(&item, expression.Name("aae.bar10").AttributeExists()); err != nil {
		t.Nil(err)
	}
	t.Equal(a, &item)

	// find item
	item = aaa{}
	if err := t.access.ScanByFilter(&item, expression.Name("aaf.bar11").AttributeExists()); err != nil {
		t.Nil(err)
	}
	t.Equal(a, &item)

	////// find item
	//item = aaa{}
	//if err := t.access.Scan(&item, expression.Name("aac.bba").Contains("foo1")); err != nil {
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

	if err := t.access.ScanByFilter(&items, expression.Name("bbd").Between(expression.Value(5), expression.Value(16))); err != nil {
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

		err := t.access.GetItem(&b, "id", b.Id)
		t.Equal(err.Error(), ErrNotFound.Error())
	}

	items := []bbb{}

	if err := t.access.ScanByFilter(&items, expression.Name("bbd").Between(expression.Value(0), expression.Value(255))); err != nil {
		t.Nil(err)
	}

	t.Len(items, 0)
}

func (t *AccessSuite) TestGetNoItem() {

	item := bbb{}

	err := t.access.GetItem(&item, "id", "aaa")
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
	t.access.tableBuilder(&user{}, output)
	t.Equal(tableExpected, output)
}

func (t *AccessSuite) TestQuery() {

	candidates := []*user{
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

	us := &user{}

	if err := t.access.Query(us, RequestInput{
		Expr: expr,
	}); err != nil {
		t.Nil(err)
	}

	t.Equal(us, candidates[0])

	expr, err = expression.NewBuilder().
		WithKeyCondition(expression.Key("created_at").Equal(expression.Value(1)).And(expression.Key("first_name").Equal(expression.Value("John")))).
		Build()
	if err != nil {
		t.Nil(err)
	}

	users := []user{}

	if err := t.access.Query(&users, RequestInput{
		Expr:      expr,
		IndexName: "created_at_first_name_index",
		Limit:     2,
		ExclusiveStartKey: map[string]dynamodb.AttributeValue{
			"email":      dynamodb.AttributeValue{S: aws.String("john@gmail.com")},
			"created_at": dynamodb.AttributeValue{N: aws.String("1")},
			"first_name": dynamodb.AttributeValue{S: aws.String("John")},
		},
		ScanIndexForward: true,
	}); err != nil {
		t.Nil(err)
	}

	t.Len(users, 2)

	expected := []user{
		*candidates[1],
		*candidates[2],
	}

	t.Equal(expected, users)

}

func (t *AccessSuite) TestQueryDdd() {

	candidates := []*ddd{
		{
			Da: "John",
		},
		{
			Da: "John",
		},
		{
			Da: "John",
		},
		{
			Da: "James",
		},
	}

	for _, candidate := range candidates {
		if err := t.access.Create(candidate); err != nil {
			t.Nil(err)
		}
	}

	expr, err := expression.NewBuilder().
		WithKeyCondition(expression.Key("dda").Equal(expression.Value("John"))).
		Build()
	if err != nil {
		t.Nil(err)
	}

	ddds := []ddd{}

	if err := t.access.Query(&ddds, RequestInput{
		Expr:      expr,
		IndexName: "index",
	}); err != nil {
		t.Nil(err)
	}

	t.Len(ddds, 3)
}

//func (t *AccessSuite) TestOrder() {
//
//	candidates := []*fff{
//		{
//			Fa: "X",
//			Fb: "Y0",
//			Fc: 1,
//		},
//		{
//			Fa: "X",
//			Fb: "Y1",
//			Fc: 2,
//		},
//		{
//			Fa: "X",
//			Fb: "Y2",
//			Fc: 3,
//		},
//		{
//			Fa: "X",
//			Fb: "Y3",
//			Fc: 4,
//		},
//	}
//
//	for _, candidate := range candidates {
//		if err := t.access.Create(candidate); err != nil {
//			fmt.Println(err)
//			t.Nil(err)
//		}
//	}
//
//	expr, err := expression.NewBuilder().Build()
//	if err != nil {
//		t.Nil(err)
//	}
//
//	fffs := []fff{}
//
//	if err := t.access.Query(&fffs, expr, "index", 0, NoPaging); err != nil {
//		fmt.Println(err)
//		t.Nil(err)
//	}
//
//	for _, fff := range fffs {
//		fmt.Println(fff)
//	}
//
//	t.NotNil(nil)
//}

func (t *AccessSuite) TestOrder() {

	candidates := []*eee{
		{
			Ea: "X",
			Eb: 1,
			Ec: 10,
		},
		{
			Ea: "X",
			Eb: 5,
			Ec: 5,
		},
		{
			Ea: "X",
			Eb: 4,
			Ec: 3,
		},
		{
			Ea: "X",
			Eb: 2,
			Ec: 7,
		},
	}

	for _, candidate := range candidates {
		if err := t.access.Create(candidate); err != nil {
			fmt.Println(err)
			t.Nil(err)
		}
	}

	expr, err := expression.NewBuilder().WithKeyCondition(expression.Key("eea").Equal(expression.Value("X")).And(expression.Key("eeb").GreaterThan(expression.Value(0)))).Build()
	if err != nil {
		t.Nil(err)
	}

	eees := []eee{}

	if err := t.access.Query(&eees, RequestInput{
		Expr:             expr,
		IndexName:        "index",
		ScanIndexForward: true,
	}); err != nil {
		fmt.Println(err)
		t.Nil(err)
	}

	expected := []eee{
		*candidates[0],
		*candidates[3],
		*candidates[2],
		*candidates[1],
	}

	t.Equal(expected, eees)
}

func TestAccessSuite(t *testing.T) {
	suite.Run(t, &AccessSuite{})
}
