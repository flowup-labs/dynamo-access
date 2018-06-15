package godynamo

type Model struct {
	Id      string `json:"id" godynamo:"hash"`
	Created int64  `json:"created"`
	Updated int64  `json:"updated"`
	Deleted int64  `json:"deleted"`
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

type User struct {
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
