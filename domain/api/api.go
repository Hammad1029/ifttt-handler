package api

import (
	"handler/common"

	jsontocql "github.com/Hammad1029/json-to-cql"
)

type Api struct {
	ApiGroup       string                                  `json:"apiGroup" mapstructure:"apiGroup"`
	ApiName        string                                  `json:"apiName" mapstructure:"apiName"`
	ApiDescription string                                  `json:"apiDescription" mapstructure:"apiDescription"`
	ApiPath        string                                  `json:"apiPath" mapstructure:"apiPath"`
	ApiRequest     common.JsonObject                       `json:"apiRequest" mapstructure:"apiRequest"`
	StartRules     []string                                `json:"startRules" mapstructure:"startRules"`
	Rules          map[string]*Rule                        `json:"rules" mapstructure:"rules"`
	Queries        map[string]jsontocql.ParameterizedQuery `json:"queries" mapstructure:"queries"`
}
