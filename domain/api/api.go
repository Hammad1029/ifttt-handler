package api

import (
	requestvalidator "ifttt/handler/domain/request_validator.go"
	"ifttt/handler/domain/resolvable"
)

type Cron struct {
	Name        string `json:"name" mapstructure:"name"`
	Description string `json:"description" mapstructure:"description"`
	CronExpr    string `json:"cronExpr" mapstructure:"cronExpr"`
	Api         Api    `json:"api" mapstructure:"api"`
}

type Api struct {
	ID          uint                                         `json:"id" mapstructure:"id"`
	Name        string                                       `json:"name" mapstructure:"name"`
	Path        string                                       `json:"path" mapstructure:"path"`
	Method      string                                       `json:"method" mapstructure:"method"`
	Description string                                       `json:"description" mapstructure:"description"`
	PreConfig   []resolvable.Resolvable                      `json:"preConfig" mapstructure:"preConfig"`
	Request     map[string]requestvalidator.RequestParameter `json:"request" mapstructure:"request"`
	Response    map[uint]resolvable.ResponseDefinition       `json:"response" mapstructure:"response"`
	Triggers    *[]TriggerCondition                          `json:"triggers" mapstructure:"triggers"`
}

type TriggerCondition struct {
	If      resolvable.Condition `json:"if" mapstructure:"if"`
	Trigger TriggerFlow          `json:"trigger" mapstructure:"trigger"`
}

type Class struct {
	Name string `json:"name" mapstructure:"name"`
}

type TriggerFlow struct {
	ID          uint                 `json:"id" mapstructure:"id"`
	Name        string               `json:"name" mapstructure:"name"`
	Description string               `json:"description" mapstructure:"description"`
	Class       Class                `json:"class" mapstructure:"class"`
	StartState  uint                 `json:"startState" mapstructure:"startState"`
	Rules       map[string]*Rule     `json:"rules" mapstructure:"rules"`
	BranchFlows map[uint]*BranchFlow `json:"branchFlows" mapstructure:"branchFlows"`
}

type BranchFlow struct {
	Rule   string        `json:"rule" mapstructure:"rule"`
	States map[uint]uint `json:"states" mapstructure:"states"`
}

type Rule struct {
	ID          uint                    `json:"id" mapstructure:"id"`
	Name        string                  `json:"name" mapstructure:"name"`
	Description string                  `json:"description" mapstructure:"description"`
	Pre         []resolvable.Resolvable `json:"pre" mapstructure:"pre"`
	Switch      RuleSwitch              `json:"switch" mapstructure:"switch"`
	Finally     []resolvable.Resolvable `json:"finally" mapstructure:"finally"`
}

type RuleSwitch struct {
	Cases   []RuleSwitchCase `json:"cases" mapstructure:"cases"`
	Default RuleSwitchCase   `json:"default" mapstructure:"default"`
}

type RuleSwitchCase struct {
	Condition resolvable.Condition    `json:"condition" mapstructure:"condition"`
	Do        []resolvable.Resolvable `json:"do" mapstructure:"do"`
	Return    uint                    `json:"return" mapstructure:"return"`
}
