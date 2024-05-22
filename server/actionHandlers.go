package server

import (
	"context"
	"fmt"
	"handler/models"
	"sync"
)

func handleActions(actions []models.ResolvableUDT, ctx context.Context) error {
	var wg sync.WaitGroup
	for _, action := range actions {
		switch action.Type {
		case "rule":
			handleActionRule(ctx, action.Data, &wg)
		case "res":
			sendResponse(ctx, "200")
		case "db", "setRes":
			{
				if _, err := action.Resolve(ctx); err != nil {
					return err
				}
			}
		default:
			return fmt.Errorf("action type %s not found", action.Type)
		}
	}
	return nil
}

func handleActionRule(ctx context.Context, actionData map[string]interface{}, wg *sync.WaitGroup) {
	wg.Add(1)
	rules := ctx.Value("rules").([]models.RuleUDT)
	ruleIdx := int(actionData["value"].(float64))
	log := ctx.Value("log").(models.LogModel)
	log.ExecutionOrder = append(log.ExecutionOrder, ruleIdx)
	go prepRule(rules[ruleIdx], wg, ctx)
	wg.Wait()
}
