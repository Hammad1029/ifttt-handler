package service

import (
	"context"
	"fmt"
	"handler/common"
	"handler/rules"
	"sync"
)

func handleActions(actions []rules.Resolvable, ctx context.Context) error {
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
	rules := ctx.Value("rules").(map[string]*rules.RuleStructure)
	ruleId := fmt.Sprint(actionData["value"])
	go prepRule(rules[ruleId], &wg, ctx, ruleId)
}
