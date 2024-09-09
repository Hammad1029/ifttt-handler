package request_data

import (
	"encoding/json"
	"fmt"
	"time"
)

type RequestData struct {
	PreConfig map[string]any            `json:"preConfig" mapstructure:"preConfig"`
	ReqBody   map[string]any            `json:"reqBody" mapstructure:"reqBody"`
	Store     map[string]any            `json:"store" mapstructure:"store"`
	Response  map[string]any            `json:"response" mapstructure:"response"`
	QueryRes  map[string][]QueryResult  `json:"queryRes" mapstructure:"queryRes"`
	ApiRes    map[string]map[string]any `json:"apiRes" mapstructure:"apiRes"`
}

type SerializedRequestData struct {
	ReqBody  string `json:"reqBody" mapstructure:"reqBody"`
	Store    string `json:"store" mapstructure:"store"`
	Response string `json:"response" mapstructure:"response"`
	QueryRes string `json:"queryRes" mapstructure:"queryRes"`
	ApiRes   string `json:"apiRes" mapstructure:"apiRes"`
}

type QueryResult struct {
	Start     time.Time         `json:"start" mapstructure:"start"`
	End       time.Time         `json:"end" mapstructure:"end"`
	TimeTaken int64             `json:"timeTaken" mapstructure:"timeTaken"`
	Results   *[]map[string]any `json:"results" mapstructure:"results"`
}

func (r *RequestData) Initialize() {
	r.PreConfig = make(map[string]any)
	r.ReqBody = make(map[string]any)
	r.Store = make(map[string]any)
	r.Response = make(map[string]any)
	r.QueryRes = make(map[string][]QueryResult)
	r.ApiRes = make(map[string]map[string]any)
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
