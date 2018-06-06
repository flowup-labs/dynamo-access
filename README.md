# dynamo-access

Collection of basics helpers for operations on the aws-dynamoDB based on [aws-sdk-go-v2](https://github.com/aws/aws-sdk-go-v2)

### Installing

```
go get github/flowup-labs/dynamo-access
```

### Quickstart
dynamo-access take name of props according json tags
```
type person struct {
	godynamo.Model

	Firstname string  `json:"firstname"`
	Surname   string  `json:"surname"`
}

func main(){
    config, err := external.LoadDefaultAWSConfig()
    if err != nil {
    	panic(err)
    }

    access := godynamo.NewDynamoAccess(config, "prefix_")

    if err := access.CreateTables(&person{}); err != nil{
        panic(err)
    }

    me := person{
        Firstname: "Vladan",
        Surname:   "Rysavy",
    }

    if err := access.Create(&me); err != nil{
        panic(err)
    }

    newMe := person{}

    if err := access.GetOneItem(&newMe, "id", me.Id); err != nil{
        panic(err)
    }
}
```
for more examples visit [access_test.go](https://github.com/flowup-labs/dynamo-access/blob/master/access_test.go)

## Running the tests

tests are build in the docker on the image `dwmkerr/dynamodb`
and for run just write:

```
make test
```


## Authors

* **Vladan Rysavy** -  [rysavyvladan](https://github.com/rysavyvladan)

## Disclaimer

This is not an official FlowUp product.
