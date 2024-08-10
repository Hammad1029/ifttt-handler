package api

type Api struct {
	ApiGroup       string           `json:"apiGroup" mapstructure:"apiGroup"`
	ApiName        string           `json:"apiName" mapstructure:"apiName"`
	ApiDescription string           `json:"apiDescription" mapstructure:"apiDescription"`
	ApiPath        string           `json:"apiPath" mapstructure:"apiPath"`
	ApiRequest     map[string]any   `json:"apiRequest" mapstructure:"apiRequest"`
	StartRules     []string         `json:"startRules" mapstructure:"startRules"`
	Rules          map[string]*Rule `json:"rules" mapstructure:"rules"`
}
