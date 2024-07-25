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
		case "sendRes":
			sendResponse(ctx, utils.Responses["success"])
		default:
			{
				if _, err := action.Resolve(ctx); err != nil {
					return err
				}
			}
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
