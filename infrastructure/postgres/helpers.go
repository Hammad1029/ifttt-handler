package infrastructure

import (
	"encoding/json"
	"fmt"
	"ifttt/handler/domain/api"
	"ifttt/handler/domain/orm_schema"
	requestvalidator "ifttt/handler/domain/request_validator.go"
	"ifttt/handler/domain/resolvable"

	"github.com/mitchellh/mapstructure"
)

func (pgRule *rules) toDomain() (*api.Rule, error) {
	domainRule := api.Rule{
		ID:          pgRule.ID,
		Name:        pgRule.Name,
		Description: pgRule.Description,
	}

	if err := json.Unmarshal(pgRule.Pre.Bytes, &domainRule.Pre); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(pgRule.Switch.Bytes, &domainRule.Switch); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(pgRule.Finally.Bytes, &domainRule.Finally); err != nil {
		return nil, err
	}

	return &domainRule, nil
}

func (t *trigger_flows) toDomain() (*api.TriggerFlow, error) {
	domanTFlow := api.TriggerFlow{
		ID:          t.ID,
		Name:        t.Name,
		Description: t.Description,
		StartState:  t.StartState,
		Rules:       map[uint]*api.Rule{},
		BranchFlows: map[uint]*api.BranchFlow{},
	}
	for _, r := range t.Rules {
		dRule, err := r.toDomain()
		if err != nil {
			return nil,
				fmt.Errorf("could not convert to domain rule: %s", err)
		}
		domanTFlow.Rules[r.ID] = dRule
	}

	if err := json.Unmarshal(t.BranchFlows.Bytes, &domanTFlow.BranchFlows); err != nil {
		return nil,
			fmt.Errorf("error in unmarshalling branchFlows: %s", err)
	}

	return &domanTFlow, nil
}

func (a *apis) toDomain() (*api.Api, error) {
	domainApi := api.Api{
		ID:          a.ID,
		Name:        a.Name,
		Path:        a.Path,
		Method:      a.Method,
		Description: a.Description,
		Request:     map[string]requestvalidator.RequestParameter{},
		PreConfig:   map[string]resolvable.Resolvable{},
		Triggers:    &[]api.TriggerCondition{},
	}

	if err := json.Unmarshal(a.Request.Bytes, &domainApi.Request); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(a.PreConfig.Bytes, &domainApi.PreConfig); err != nil {
		return nil, err
	}

	var tConditions []api_trigger_flow_json
	if err := json.Unmarshal(a.TriggerFlows.Bytes, &tConditions); err != nil {
		return nil, err
	}

	triggerFlowMap := make(map[uint]trigger_flows)
	for _, tFlow := range a.Triggers {
		triggerFlowMap[tFlow.ID] = tFlow
	}

	for _, tc := range tConditions {
		tcModel, ok := triggerFlowMap[tc.Trigger]
		if !ok {
			return nil,
				fmt.Errorf("trigger flow not found from conditions")
		}
		domainTFlow, err := tcModel.toDomain()
		if err != nil {
			return nil, err
		}
		*domainApi.Triggers = append(*domainApi.Triggers,
			api.TriggerCondition{If: tc.If, Trigger: *domainTFlow})
	}

	return &domainApi, nil
}

func (c *crons) toDomain() (*api.Cron, error) {
	dCron := api.Cron{
		ID:           c.ID,
		Name:         c.Name,
		Description:  c.Name,
		Cron:         c.Cron,
		TriggerFlows: &[]api.TriggerCondition{},
	}

	if err := json.Unmarshal(c.PreConfig.Bytes, &dCron.PreConfig); err != nil {
		return nil,
			fmt.Errorf("could not cast pgApi: %s", err)
	}

	var tConditions []api_trigger_flow_json
	if err := json.Unmarshal(c.TriggerFlows.Bytes, &tConditions); err != nil {
		return nil,
			fmt.Errorf("could not cast pgApi: %s", err)
	}

	triggerFlowMap := make(map[uint]trigger_flows)
	for _, tFlow := range c.TriggerFlowRef {
		triggerFlowMap[tFlow.ID] = tFlow
	}

	for _, tc := range tConditions {
		tcModel, ok := triggerFlowMap[tc.Trigger]
		if !ok {
			return nil,
				fmt.Errorf("trigger flow not found from conditions")
		}
		domainTFlow, err := tcModel.toDomain()
		if err != nil {
			return nil, err
		}
		*dCron.TriggerFlows = append(*dCron.TriggerFlows, api.TriggerCondition{If: tc.If, Trigger: *domainTFlow})
	}

	return &dCron, nil
}

func (o *orm_model) fromDomain(dModel *orm_schema.Model) error {
	return mapstructure.Decode(dModel, o)
}

// func (o *orm_projection) fromDomain(dProjection *orm_schema.Projection) error {
// 	return mapstructure.Decode(dProjection, o)
// }

func (o *orm_association) fromDomain(dAssociation *orm_schema.ModelAssociation) error {
	return mapstructure.Decode(dAssociation, o)
}

func (o *orm_model) toDomain() (*orm_schema.Model, error) {
	var domain orm_schema.Model
	if err := mapstructure.Decode(o, &domain); err != nil {
		return nil, err
	}
	return &domain, nil
}

// func (o *orm_projection) toDomain() (*orm_schema.Projection, error) {
// 	var domain orm_schema.Projection
// 	if err := mapstructure.Decode(o, &domain); err != nil {
// 		return nil, err
// 	}
// 	return &domain, nil
// }

func (o *orm_association) toDomain() (*orm_schema.ModelAssociation, error) {
	var domain orm_schema.ModelAssociation
	if err := mapstructure.Decode(o, &domain); err != nil {
		return nil, err
	}
	return &domain, nil
}
