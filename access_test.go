package godynamo

import (
	"github.com/stretchr/testify/suite"
	"testing"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/defaults"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/expression"
	"fmt"
)

type AccessSuite struct {
	suite.Suite

	svc    *dynamodb.DynamoDB
	access *dynamoAccess
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

	t.access.DropTables(&aaa{}, &ccc{}, &bbb{})

	t.access.CreateTables(&aaa{}, &ccc{}, &bbb{})

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
		tableName, err := t.access.reflect(candidate.item)
		t.Nil(err)
		t.Equal(candidate.expectedName, tableName)
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
	if err := t.access.QueryByAttribute(&item, "id", a.Id); err != nil {
		fmt.Println(err)
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

func (t *AccessSuite) TestScanByNestedId() {
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
				Ba: "bubu",
			},
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
	//item = aaa{}
	//if err := t.access.ScanCustom(&item, expression.Name("aac").Contains("bar10")); err != nil {
	//	t.Nil(err)
	//}
	//t.Equal(a, &item)

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

	t.Nil(t.access.Create(&b))

	item := bbb{}
	if err := t.access.ScanByAttribute(&item, "cId", a.Id); err != nil {
		t.Nil(err)
	}

	t.Equal(b, item)
}

func TestAccessSuite(t *testing.T) {
	suite.Run(t, &AccessSuite{})
}

//config, err := external.LoadDefaultAWSConfig()
//if err != nil {
//	panic(err)
//}
//
