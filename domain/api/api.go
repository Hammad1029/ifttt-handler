package api

import (
	"ifttt/handler/domain/resolvable"
)

type Api struct {
	ID           uint                             `json:"id" mapstructure:"id"`
	Name         string                           `json:"name" mapstructure:"name"`
	Path         string                           `json:"path" mapstructure:"path"`
	Method       string                           `json:"method" mapstructure:"method"`
	Description  string                           `json:"description" mapstructure:"description"`
	Request      map[string]any                   `json:"request" mapstructure:"request"`
	PreConfig    map[string]resolvable.Resolvable `json:"preConfig" mapstructure:"preConfig"`
	TriggerFlows *[]TriggerCondition              `json:"triggerFlows" mapstructure:"triggerFlows"`
}

type TriggerCondition struct {
	If      Condition   `json:"if" mapstructure:"if"`
	Trigger TriggerFlow `json:"trigger" mapstructure:"trigger"`
}

type Class struct {
	Name string `json:"name" mapstructure:"name"`
}

type TriggerFlow struct {
	ID          uint                   `json:"id" mapstructure:"id"`
	Name        string                 `json:"name" mapstructure:"name"`
	Description string                 `json:"description" mapstructure:"description"`
	Class       Class                  `json:"class" mapstructure:"class"`
	StartRules  []uint                 `json:"startRules" mapstructure:"startRules"`
	AllRules    map[uint]*Rule         `json:"allRules" mapstructure:"allRules"`
	BranchFlows map[uint]*[]BranchFlow `json:"branchFlows" mapstructure:"branchFlows"`
}

type BranchFlow struct {
	IfReturn resolvable.Resolvable `json:"ifReturn" mapstructure:"ifReturn"`
	Jump     uint                  `json:"jump" mapstructure:"jump"`
}

type Rule struct {
	ID          uint                    `json:"id" mapstructure:"id"`
	Name        string                  `json:"name" mapstructure:"name"`
	Description string                  `json:"description" mapstructure:"description"`
	Pre         []resolvable.Resolvable `json:"pre" mapstructure:"pre"`
	Switch      RuleSwitch              `json:"switch" mapstructure:"switch"`
}

type RuleSwitch struct {
	Cases   []RuleSwitchCase `json:"cases" mapstructure:"cases"`
	Default RuleSwitchCase   `json:"default" mapstructure:"default"`
}

type RuleSwitchCase struct {
	Condition Condition               `json:"condition" mapstructure:"condition"`
	Do        []resolvable.Resolvable `json:"do" mapstructure:"do"`
	Return    resolvable.Resolvable   `json:"return" mapstructure:"return"`
}
