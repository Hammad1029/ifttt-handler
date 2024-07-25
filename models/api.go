package models

import (
	"encoding/json"
	"fmt"

	jsontocql "github.com/Hammad1029/json-to-cql"
	"github.com/scylladb/gocqlx/v2/table"
)

type ApiModelSerialized struct {
	ApiGroup       string   `cql:"api_group"`
	ApiName        string   `cql:"api_name"`
	ApiDescription string   `cql:"api_description"`
	ApiPath        string   `cql:"api_path"`
	ApiRequest     string   `cql:"api_request"`
	StartRules     []string `cql:"start_rules"`
	Rules          string   `cql:"rules"`
	Queries        string   `cql:"queries"`
}

type ApiModel struct {
	ApiGroup       string                                  `json:"apiGroup" mapstructure:"apiGroup"`
	ApiName        string                                  `json:"apiName" mapstructure:"apiName"`
	ApiDescription string                                  `json:"apiDescription" mapstructure:"apiDescription"`
	ApiPath        string                                  `json:"apiPath" mapstructure:"apiPath"`
	ApiRequest     map[string]interface{}                  `json:"apiRequest" mapstructure:"apiRequest"`
	StartRules     []string                                `json:"startRules" mapstructure:"startRules"`
	Rules          map[string]*RuleUDT                     `json:"rules" mapstructure:"rules"`
	Queries        map[string]jsontocql.ParameterizedQuery `json:"queries" mapstructure:"queries"`
}

var ApisMetadata = table.Metadata{
	Name:    "Apis",
	Columns: []string{"api_group", "api_name", "api_description", "api_path", "api_request", "start_rules", "rules", "queries"},
	PartKey: []string{"api_group"},
	SortKey: []string{"api_name", "api_description"},
}

func (a *ApiModelSerialized) Deserialize() (ApiModel, error) {
	unserializedApi := ApiModel{}
	unserializedApi.ApiDescription = a.ApiDescription
	unserializedApi.ApiGroup = a.ApiGroup
	unserializedApi.ApiName = a.ApiName
	unserializedApi.ApiPath = a.ApiPath
	unserializedApi.StartRules = a.StartRules

	err := json.Unmarshal([]byte(a.ApiRequest), &unserializedApi.ApiRequest)
	if err != nil {
		return unserializedApi, fmt.Errorf("method ReadApisToRedis: %s", err)
	}
	err = json.Unmarshal([]byte(a.Rules), &unserializedApi.Rules)
	if err != nil {
		return unserializedApi, fmt.Errorf("method ReadApisToRedis: %s", err)
	}
	err = json.Unmarshal([]byte(a.Queries), &unserializedApi.Queries)
	if err != nil {
		return unserializedApi, fmt.Errorf("method ReadApisToRedis: %s", err)
	}

	return unserializedApi, nil
}
