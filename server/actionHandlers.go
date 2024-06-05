package server

import (
	"context"
	"fmt"
	"handler/models"
	"strconv"
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
		case "db", "setRes", "store":
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

func handleActionRule(ctx context.Context, actionData map[string]string, wg *sync.WaitGroup) {
	wg.Add(1)
	rules := ctx.Value("rules").([]models.RuleUDT)
	ruleIdx, err := strconv.Atoi(actionData["value"])
	if err != nil {
		addErrorToContext(err, ctx, false)
		wg.Done()
		return
	}
	go prepRule(rules[ruleIdx], wg, ctx, ruleIdx)
	wg.Wait()
}
