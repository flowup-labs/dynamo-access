package database

type Model struct {
	Id      string `json:"id"`
	Created int64  `json:"created"`
	Updated int64  `json:"updated"`
	Deleted int64  `json:"deleted"`
}

type Aaa struct {
	Model

	Aa string `json:"aaa"`
	Ab string `json:"aab"`
	Ac []Bbb  `json:"aac"`
}

type Bbb struct {
	Model

	Ba  string `json:"bba"`
	Bb  []Ddd  `json:"bbb"`
	CId string `json:"cId"`
}

type Ccc struct {
	Model

	Ca  string `json:"cca"`
	DId string `json:"dId"`
}

type Ddd struct {
	Model

	Da string `json:"dda"`
	Db string `json:"ddb"`
}
