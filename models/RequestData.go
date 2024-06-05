package models

type RequestData struct {
	ReqBody  map[string]interface{}              `cql:"req_body"`
	Store    map[string]interface{}              `cql:"store"`
	Response map[string]interface{}              `cql:"response"`
	QueryRes map[string][]map[string]interface{} `cql:"query_res"`
}

func (r *RequestData) Initialize(api *ApiModel) {
	r.ReqBody = make(map[string]interface{})
	r.Store = make(map[string]interface{})
	r.Response = make(map[string]interface{})
	r.QueryRes = make(map[string][]map[string]interface{})
}
