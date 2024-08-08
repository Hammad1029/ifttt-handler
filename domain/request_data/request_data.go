package request_data

import (
	"encoding/json"
	"fmt"
	"handler/common"
	"time"
)

type RequestData struct {
	ReqBody  common.JsonObject            `json:"reqBody" mapstructure:"reqBody"`
	Store    common.JsonObject            `json:"store" mapstructure:"store"`
	Response common.JsonObject            `json:"response" mapstructure:"response"`
	QueryRes map[string][]QueryResult     `json:"queryRes" mapstructure:"queryRes"`
	ApiRes   map[string]common.JsonObject `json:"apiRes" mapstructure:"apiRes"`
}

type SerializedRequestData struct {
	ReqBody  string `json:"reqBody" mapstructure:"reqBody"`
	Store    string `json:"store" mapstructure:"store"`
	Response string `json:"response" mapstructure:"response"`
	QueryRes string `json:"queryRes" mapstructure:"queryRes"`
	ApiRes   string `json:"apiRes" mapstructure:"apiRes"`
}

type QueryResult struct {
	Start     time.Time            `json:"start" mapstructure:"start"`
	End       time.Time            `json:"end" mapstructure:"end"`
	TimeTaken int64                `json:"timeTaken" mapstructure:"timeTaken"`
	Results   *[]common.JsonObject `json:"results" mapstructure:"results"`
}

func (r *RequestData) Initialize() {
	r.ReqBody = make(common.JsonObject)
	r.Store = make(common.JsonObject)
	r.Response = make(common.JsonObject)
	r.QueryRes = make(map[string][]QueryResult)
	r.ApiRes = make(map[string]common.JsonObject)
}

func (r *RequestData) Serialize() (SerializedRequestData, error) {
	var serializedLog SerializedRequestData
	reqBodySerialized, err := json.Marshal(r.ReqBody)
	if err != nil {
		return serializedLog, fmt.Errorf("method RequestData.Serialize: could not serialize request body: %s", err)
	}
	storeSerialized, err := json.Marshal(r.Store)
	if err != nil {
		return serializedLog, fmt.Errorf("method RequestData.Serialize: could not serialize store: %s", err)
	}
	responseSerialized, err := json.Marshal(r.Response)
	if err != nil {
		return serializedLog, fmt.Errorf("method RequestData.Serialize: could not serialize response: %s", err)
	}
	queryResSerialized, err := json.Marshal(r.QueryRes)
	if err != nil {
		return serializedLog, fmt.Errorf("method RequestData.Serialize: could not serialize query results: %s", err)
	}
	apiResSerialized, err := json.Marshal(r.ApiRes)
	if err != nil {
		return serializedLog, fmt.Errorf("method RequestData.Serialize: could not serialize api results: %s", err)
	}

	serializedLog.ReqBody = string(reqBodySerialized)
	serializedLog.Store = string(storeSerialized)
	serializedLog.Response = string(responseSerialized)
	serializedLog.QueryRes = string(queryResSerialized)
	serializedLog.ApiRes = string(apiResSerialized)

	return serializedLog, nil
}
