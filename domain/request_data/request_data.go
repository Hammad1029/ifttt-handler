package request_data

import (
	"sync"
)

type RequestData struct {
	ReqBody   map[string]any              `json:"reqBody" mapstructure:"reqBody"`
	PreConfig *sync.Map                   `json:"preConfig" mapstructure:"preConfig"`
	Store     *sync.Map                   `json:"store" mapstructure:"store"`
	Response  *sync.Map                   `json:"response" mapstructure:"response"`
	QueryRes  map[string][]map[string]any `json:"queryRes" mapstructure:"queryRes"`
	ApiRes    map[string]map[string]any   `json:"apiRes" mapstructure:"apiRes"`
}

func (r *RequestData) Initialize() {
	r.ReqBody = make(map[string]any)
	r.PreConfig = &sync.Map{}
	r.Store = &sync.Map{}
	r.Response = &sync.Map{}
	r.QueryRes = make(map[string][]map[string]any)
	r.ApiRes = make(map[string]map[string]any)
}
