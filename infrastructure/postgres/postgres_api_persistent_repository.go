package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"ifttt/handler/domain/api"

	"github.com/samber/lo"
)

type PostgresApiRepository struct {
	*PostgresBaseRepository
}

func NewPostgresApiRepository(base *PostgresBaseRepository) *PostgresApiRepository {
	return &PostgresApiRepository{PostgresBaseRepository: base}
}

func (p *PostgresApiRepository) GetAllApis(ctx context.Context) (*[]api.Api, error) {
	var domainApis []api.Api
	var postgresApis []apis
	if err := p.client.
		Preload("Triggerflows").Preload("Triggerflows.StartRules").Preload("Triggerflows.AllRules").
		Find(&postgresApis).Error; err != nil {
		return nil, fmt.Errorf("method *PostgresApiRepository.GetAllApis: could not get apis from postgres: %s", err)
	}

	// customDecoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
	// 	DecodeHook: mapstructure.ComposeDecodeHookFunc(
	// 		customApiDecoder,
	// 	),
	// 	Result: &domainApis,
	// })
	// if err != nil {
	// 	return nil, fmt.Errorf("method *PostgresApiRepository.GetAllApis: could not create custom decoder: %s", err)
	// }

	for _, currApi := range postgresApis {
		newApi := api.Api{
			Name:         currApi.Name,
			Path:         currApi.Path,
			Method:       currApi.Method,
			TriggerFlows: &[]api.TriggerFlow{},
		}
		for _, currTFlow := range currApi.Triggerflows {
			newTFlow := api.TriggerFlow{
				StartRules: []uint{},
				AllRules:   map[uint]*api.Rule{},
			}
			newTFlow.StartRules = lo.Map(currTFlow.StartRules, func(rule rules, _ int) uint {
				return rule.ID
			})
			for _, rule := range currTFlow.AllRules {
				newRule := api.Rule{}
				if err := json.Unmarshal(rule.Conditions.Bytes, &newRule.Conditions); err != nil {
					return nil, fmt.Errorf("method *PostgresApiRepository.GetAllApis: could not unmarshal conditions: %s", err)
				}
				if err := json.Unmarshal(rule.Then.Bytes, &newRule.Then); err != nil {
					return nil, fmt.Errorf("method *PostgresApiRepository.GetAllApis: could not unmarshal then: %s", err)
				}
				if err := json.Unmarshal(rule.Else.Bytes, &newRule.Else); err != nil {
					return nil, fmt.Errorf("method *PostgresApiRepository.GetAllApis: could not unmarshal else: %s", err)
				}
				newTFlow.AllRules[rule.ID] = &newRule
			}
			*newApi.TriggerFlows = append(*newApi.TriggerFlows, newTFlow)
		}
		if err := json.Unmarshal(currApi.Request.Bytes, &newApi.Request); err != nil {
			return nil, fmt.Errorf("method *PostgresApiRepository.GetAllApis: could not unmarshal request: %s", err)
		}
		if err := json.Unmarshal(currApi.PreConfig.Bytes, &newApi.PreConfig); err != nil {
			return nil, fmt.Errorf("method *PostgresApiRepository.GetAllApis: could not unmarshal preconfig: %s", err)
		}
		domainApis = append(domainApis, newApi)
	}

	return &domainApis, nil
}

// func customApiDecoder(
// 	from reflect.Type, to reflect.Type, data interface{}) (interface{}, error) {
// 	switch from {
// 	case reflect.TypeOf([]rules{}):
// 		{
// 			rulesList := data.([]rules)
// 			switch to {
// 			case reflect.TypeOf([]uint{}):
// 				{
// 					return lo.Map(rulesList, func(rule rules, _ int) uint {
// 						return rule.ID
// 					}), nil
// 				}
// 			case reflect.TypeOf(map[uint]api.Rule{}):
// 				{
// 					rulesMap := map[uint]api.Rule{}
// 					for _, rule := range rulesList {
// 						newRule := api.Rule{}
// 						customDecoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
// 							DecodeHook: mapstructure.ComposeDecodeHookFunc(
// 								customApiDecoder,
// 							),
// 							Result: &newRule,
// 						})
// 						if err != nil {
// 							return nil, fmt.Errorf("method *PostgresApiRepository.GetAllApis: could not create custom decoder: %s", err)
// 						}
// 						if err := customDecoder.Decode(rule); err != nil {
// 							return nil, fmt.Errorf("method *PostgresApiRepository.GetAllApis: could not decode rule: %s", err)
// 						}
// 						rulesMap[rule.ID] = newRule
// 					}
// 					return rulesMap, nil
// 				}
// 			}
// 		}
// 	case reflect.TypeOf(pgtype.JSONB{}):
// 		{
// 			jsonb := data.(pgtype.JSONB)
// 			output := reflect.ValueOf(to).Interface()
// 			if err := json.Unmarshal(jsonb.Bytes, &output); err != nil {
// 				return nil, fmt.Errorf("method *PostgresApiRepository.GetAllApis: could not decode json: %s", err)
// 			}
// 			return output, nil
// 		}
// 	}
// 	return data, nil
// }
