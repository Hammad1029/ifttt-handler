package models

import (
	"encoding/json"
	"fmt"
	"handler/common"
)

type RequestDataModel struct {
	ReqBody  string `cql:"req_body" json:"reqBody" mapstructure:"reqBody"`
	Store    string `cql:"store" json:"store" mapstructure:"store"`
	Response string `cql:"response" json:"response" mapstructure:"response"`
	QueryRes string `cql:"query_res" json:"queryRes" mapstructure:"queryRes"`
	ApiRes   string `cql:"api_res" json:"apiRes" mapstructure:"apiRes"`
}

type RequestData struct {
	ReqBody  common.JsonObject              `json:"reqBody" mapstructure:"reqBody"`
	Store    common.JsonObject              `json:"store" mapstructure:"store"`
	Response common.JsonObject              `json:"response" mapstructure:"response"`
	QueryRes map[string][]common.JsonObject `json:"queryRes" mapstructure:"queryRes"`
	ApiRes   map[string]common.JsonObject   `json:"apiRes" mapstructure:"apiRes"`
}

func (r *RequestData) Initialize() {
	r.ReqBody = make(common.JsonObject)
	r.Store = make(common.JsonObject)
	r.Response = make(common.JsonObject)
	r.QueryRes = make(map[string][]common.JsonObject)
	r.ApiRes = make(map[string]common.JsonObject)
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
