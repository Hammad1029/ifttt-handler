package api

import (
	"encoding/json"
	"fmt"
)

const (
	RulesApiType   = "rules"
	DumpingApiType = "dumping"
)

type ApiSerialized struct {
	Group       string   `json:"group" mapstructure:"group"`
	Name        string   `json:"name" mapstructure:"name"`
	Method      string   `json:"method" mapstructure:"method"`
	Type        string   `json:"type" mapstructure:"type"`
	Path        string   `json:"path" mapstructure:"path"`
	Description string   `json:"description" mapstructure:"description"`
	Request     string   `json:"request" mapstructure:"request"`
	Dumping     string   `json:"dumping" mapstructure:"dumping"`
	StartRules  []string `json:"rules" mapstructure:"rules"`
	Rules       string   `json:"startRules" mapstructure:"startRules"`
}

type Api struct {
	Group       string           `json:"group" mapstructure:"group"`
	Name        string           `json:"name" mapstructure:"name"`
	Method      string           `json:"method" mapstructure:"method"`
	Type        string           `json:"type" mapstructure:"type"`
	Path        string           `json:"path" mapstructure:"path"`
	Description string           `json:"description" mapstructure:"description"`
	Request     map[string]any   `json:"request" mapstructure:"request"`
	Dumping     Dumping          `json:"dumping" mapstructure:"dumping"`
	StartRules  []string         `json:"startRules" mapstructure:"startRules"`
	Rules       map[string]*Rule `json:"rules" mapstructure:"rules"`
}

func UnserializeApis(serializedApis *[]ApiSerialized) (*[]Api, error) {
	if serializedApis == nil {
		return nil, nil
	}
	unserializedApis := &([]Api{})
	for _, v := range *serializedApis {
		unserialized, err := v.Unserialize()
		if err != nil {
			return nil, fmt.Errorf("method *ScyllaApiPersistentRepository.UnserializeApis: failed to unserialize apis: %s", err)
		}
		*unserializedApis = append(*unserializedApis, *unserialized)
	}
	return unserializedApis, nil
}

func (a *ApiSerialized) Unserialize() (*Api, error) {
	unserializedApi := Api{}
	unserializedApi.Group = a.Group
	unserializedApi.Name = a.Name
	unserializedApi.Method = a.Method
	unserializedApi.Type = a.Type
	unserializedApi.Path = a.Path
	unserializedApi.Description = a.Description
	unserializedApi.StartRules = a.StartRules

	err := json.Unmarshal([]byte(a.Request), &unserializedApi.Request)
	if err != nil {
		return nil, fmt.Errorf("method ScyllaApi.unserialize: could not unserialize api request: %s", err)
	}
	err = json.Unmarshal([]byte(a.Rules), &unserializedApi.Rules)
	if err != nil {
		return nil, fmt.Errorf("method ScyllaApi.unserialize: could not unserialize rules: %s", err)
	}
	err = json.Unmarshal([]byte(a.Dumping), &unserializedApi.Dumping)
	if err != nil {
		return nil, fmt.Errorf("method ScyllaApi.unserialize: could not unserialize dumping: %s", err)
	}
	return &unserializedApi, nil
}
