package server

import (
	"context"
	"errors"
	"handler/models"
	"handler/utils"
	"log"
	"sync"
)

func initExec(startRules []string, ctx context.Context) {
	var wg sync.WaitGroup
	rules := ctx.Value("rules").(map[string]*models.RuleUDT)
	for _, startId := range startRules {
		wg.Add(1)
		go prepRule(rules[startId], &wg, ctx, startId)
	}
	wg.Wait()
}

func prepRule(rule *models.RuleUDT, wg *sync.WaitGroup, ctx context.Context, ruleId string) {
	defer wg.Done()
	if l, ok := ctx.Value("log").(*models.LogModel); ok {
		l.ExecutionOrder = append(l.ExecutionOrder, ruleId)
	} else {
		log.Panic("method prepRule: could not type cast log model")
	}
	if err := execRule(rule, ctx); err != nil {
		addErrorToContext(err, ctx, true)
		return
	}
}

func execRule(rule *models.RuleUDT, ctx context.Context) error {
	if ev, err := rule.Conditions.EvaluateGroup(ctx); err != nil {
		return err
	} else if ev {
		return handleActions(rule.Then, ctx)
	} else {
		return handleActions(rule.Else, ctx)
	}
}

func addErrorToContext(err error, ctx context.Context, sendRes bool) {
	if l, ok := ctx.Value("log").(*models.LogModel); ok {
		l.AddExecLog("system", "error", err.Error())
	} else {
		log.Panic("method addErrorToContext: could not type cast log model")
	}

	log.Fatal(err.Error())

	if sendRes {
		sendResponse(ctx, utils.Responses["ServerError"])
	}
}

func sendResponse(ctx context.Context, res utils.Response) error {
	if responseChannel, ok := ctx.Value("resChan").(chan utils.Response); ok {
		responseChannel <- res
		return nil
	} else {
		err := errors.New("method sendResponse: send res type assertion failed")
		log.Panic(err)
		return err
	}
}
