package models

type RequestData struct {
	ReqBody  map[string]interface{}              `cql:"req_body"`
	Store    map[string]interface{}              `cql:"store"`
	Response map[string]interface{}              `cql:"response"`
	Errors   []string                            `cql:"errors"`
	QueryRes map[string][]map[string]interface{} `cql:"query_res"`
	Queries  map[string]QueryUDT                 `cql:"queries"`
	Comments []string                            `cql:"comments"`
}

func (r *RequestData) Initialize() {
	r.ReqBody = make(map[string]interface{})
	r.Store = make(map[string]interface{})
	r.Response = make(map[string]interface{})
	r.Errors = []string{}
	r.QueryRes = make(map[string][]map[string]interface{})
	r.Queries = make(map[string]QueryUDT)
	r.Comments = []string{}
}

func (r *RequestData) AddError(e error) {
	r.Errors = append(r.Errors, e.Error())
}

func (r *RequestData) AddComment(c string) {
	r.Comments = append(r.Comments, c)
}
