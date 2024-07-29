package server

import (
	"context"
	"fmt"
	"handler/common"
	"handler/models"
	"sync"
)

func handleActions(actions []models.Resolvable, ctx context.Context) error {
	for _, action := range actions {
		switch action.ResolveType {
		case "rule":
			handleActionRule(ctx, action.ResolveData)
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

func handleActionRule(ctx context.Context, actionData common.JsonObject) {
	var wg sync.WaitGroup
	wg.Add(1)
	rules := ctx.Value("rules").(map[string]*models.RuleUDT)
	ruleId := fmt.Sprint(actionData["value"])
	go prepRule(rules[ruleId], &wg, ctx, ruleId)
}
