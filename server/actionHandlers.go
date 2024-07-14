package server

import (
	"context"
	"fmt"
	"handler/models"
	"handler/utils"
	"sync"
)

func handleActions(actions []models.Resolvable, ctx context.Context) error {
	for _, action := range actions {
		switch action.ResolveType {
		case "rule":
			handleActionRule(ctx, action.ResolveData)
		case "res":
			sendResponse(ctx, utils.Responses["success"])
		case "db", "setRes", "setStore", "log":
			{
				if _, err := action.Resolve(ctx); err != nil {
					return err
				}
			}
		default:
			return fmt.Errorf("method handleActions: action type %s not found", action.ResolveType)
		}
	}
	return nil
}

func handleActionRule(ctx context.Context, actionData map[string]interface{}) {
	var wg sync.WaitGroup
	wg.Add(1)
	rules := ctx.Value("rules").(map[string]*models.RuleUDT)
	ruleId := fmt.Sprint(actionData["value"])
	go prepRule(rules[ruleId], &wg, ctx, ruleId)
}
