package database

import (
	"github.com/stretchr/testify/suite"
	"testing"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/defaults"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
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

	t.access.DropTables(&Aaa{}, &Ccc{}, &Bbb{})

	t.access.CreateTables(&Aaa{}, &Ccc{}, &Bbb{})

}

func (t *AccessSuite) TestReflect() {
	candidates := []struct {
		item         interface{}
		expectedName string
		expectedErr  error
	}{
		{
			item:         &[]Aaa{},
			expectedName: "Aaa",
			expectedErr:  nil,
		},

		{
			item:         &[]*Aaa{},
			expectedName: "Aaa",
			expectedErr:  nil,
		},
		{
			item:         &Aaa{},
			expectedName: "Aaa",
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

	a := &Aaa{
		Aa: "Aaa",
		Ab: "Aab",
		Ac: []Bbb{
			{
				Ba: "Bba",
				Bb: []Ddd{
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


func (t *AccessSuite) TestFindOneItem() {

	a := &Aaa{
		Aa: "Aa",
		Ab: "Ab",
		Ac: []Bbb{
			{
				Ba: "Ba",
				Bb: []Ddd{
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
	item := Aaa{}
	if err := t.access.FindByAttribute(&item, "id", a.Id); err != nil {
		t.Nil(err)
	}

	t.Equal(item.Id, a.Id)
	t.Equal(*a, item)
}

func (t *AccessSuite) TestUpdateItem() {

	a := &Aaa{
		Aa: "Aa",
		Ab: "Ab",
		Ac: []Bbb{
			{
				Ba: "Ba",
				Bb: []Ddd{
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
	item := Aaa{}
	if err := t.access.FindByAttribute(&item, "id", a.Id); err != nil {
		t.Nil(err)
	}

	t.Equal(item.Id, a.Id)
	t.Equal(*a, item)
}

func (t *AccessSuite) TestFindOneItemByIndex() {
	//create item
	c := &Ccc{
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
	items := []Ccc{}
	if err := t.access.FindByAttribute(&items, "dId", c.DId); err != nil {
		t.Nil(err)
	}

	t.Len(items, 2)
}

func (t *AccessSuite) TestCreateRelationship() {

	a := Aaa{
		Aa: "Aa",
		Ab: "Ab",
	}

	t.Nil(t.access.Create(&a))

	b := Bbb{
		Ba: "Ba",
		Bb: []Ddd{
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

	item := Bbb{}
	if err := t.access.FindByAttribute(&item, "cId", a.Id); err != nil {
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
