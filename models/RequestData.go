package models

import (
	"encoding/json"
	"fmt"
)

type RequestDataModel struct {
	ReqBody  string `cql:"req_body" json:"reqBody" mapstructure:"reqBody"`
	Store    string `cql:"store" json:"store" mapstructure:"store"`
	Response string `cql:"response" json:"response" mapstructure:"response"`
	QueryRes string `cql:"query_res" json:"queryRes" mapstructure:"queryRes"`
	ApiRes   string `cql:"api_res" json:"apiRes" mapstructure:"apiRes"`
}

type RequestData struct {
	ReqBody  map[string]interface{}              `json:"reqBody" mapstructure:"reqBody"`
	Store    map[string]interface{}              `json:"store" mapstructure:"store"`
	Response map[string]interface{}              `json:"response" mapstructure:"response"`
	QueryRes map[string][]map[string]interface{} `json:"queryRes" mapstructure:"queryRes"`
	ApiRes   map[string]map[string]interface{}   `json:"apiRes" mapstructure:"apiRes"`
}

func (r *RequestData) Initialize() {
	r.ReqBody = make(map[string]interface{})
	r.Store = make(map[string]interface{})
	r.Response = make(map[string]interface{})
	r.QueryRes = make(map[string][]map[string]interface{})
	r.ApiRes = make(map[string]map[string]interface{})
}

func (r *RequestData) serialize() (RequestDataModel, error) {
	var serializedLog RequestDataModel
	reqBodySerialized, err := json.Marshal(r.ReqBody)
	if err != nil {
		return serializedLog, fmt.Errorf("method serialize: could not serialize request body: %s", err)
	}
	storeSerialized, err := json.Marshal(r.Store)
	if err != nil {
		return serializedLog, fmt.Errorf("method serialize: could not serialize store: %s", err)
	}
	responseSerialized, err := json.Marshal(r.Response)
	if err != nil {
		return serializedLog, fmt.Errorf("method serialize: could not serialize response: %s", err)
	}
	queryResSerialized, err := json.Marshal(r.QueryRes)
	if err != nil {
		return serializedLog, fmt.Errorf("method serialize: could not serialize query results: %s", err)
	}
	apiResSerialized, err := json.Marshal(r.ApiRes)
	if err != nil {
		return serializedLog, fmt.Errorf("method serialize: could not serialize api results: %s", err)
	}

	serializedLog.ReqBody = string(reqBodySerialized)
	serializedLog.Store = string(storeSerialized)
	serializedLog.Response = string(responseSerialized)
	serializedLog.QueryRes = string(queryResSerialized)
	serializedLog.ApiRes = string(apiResSerialized)

	return serializedLog, nil
}
