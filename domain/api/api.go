package api

import (
	"ifttt/handler/domain/resolvable"
)

type Api struct {
	Name         string                           `json:"name" mapstructure:"name"`
	Path         string                           `json:"path" mapstructure:"path"`
	Method       string                           `json:"method" mapstructure:"method"`
	Request      map[string]any                   `json:"request" mapstructure:"request"`
	PreConfig    map[string]resolvable.Resolvable `json:"preConfig" mapstructure:"preConfig"`
	TriggerFlows *[]TriggerFlow                   `json:"triggerFlows" mapstructure:"triggerFlows"`
}

type TriggerFlow struct {
	StartRules []uint         `json:"startRules" mapstructure:"startRules"`
	AllRules   map[uint]*Rule `json:"allRules" mapstructure:"allRules"`
}

type Rule struct {
	Conditions Condition               `json:"conditions" mapstructure:"conditions"`
	Then       []resolvable.Resolvable `json:"then" mapstructure:"then"`
	Else       []resolvable.Resolvable `json:"else" mapstructure:"else"`
}
