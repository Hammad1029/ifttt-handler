package api

import "ifttt/handler/domain/resolvable"

type triggerFlow struct {
	Name        string  `mapstructure:"name" json:"name"`
	Description string  `mapstructure:"description" json:"description"`
	Class       string  `mapstrucutre:"class" json:"class"`
	Rules       *[]Rule `mapstructure:"rules" json:"rules"`
}

type Rule struct {
	Id          string                  `json:"id" mapstructure:"id"`
	Description string                  `json:"description" mapstructure:"description"`
	Conditions  Condition               `json:"conditions" mapstructure:"conditions"`
	Then        []resolvable.Resolvable `json:"then" mapstructure:"then"`
	Else        []resolvable.Resolvable `json:"else" mapstructure:"else"`
}
