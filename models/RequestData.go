package models

type RequestDataModel struct {
	ReqBody  string `cql:"req_body" json:"reqBody" mapstructure:"reqBody"`
	Store    string `cql:"store" json:"store" mapstructure:"store"`
	Response string `cql:"response" json:"response" mapstructure:"response"`
	QueryRes string `cql:"query_res" json:"queryRes" mapstructure:"queryRes"`
}

type RequestData struct {
	ReqBody  map[string]interface{}              `json:"reqBody" mapstructure:"reqBody"`
	Store    map[string]interface{}              `json:"store" mapstructure:"store"`
	Response map[string]interface{}              `json:"response" mapstructure:"response"`
	QueryRes map[string][]map[string]interface{} `json:"queryRes" mapstructure:"queryRes"`
}

func (r *RequestData) Initialize(api *ApiModel) {
	r.ReqBody = make(map[string]interface{})
	r.Store = make(map[string]interface{})
	r.Response = make(map[string]interface{})
	r.QueryRes = make(map[string][]map[string]interface{})
}
