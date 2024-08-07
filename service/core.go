package service

import (
	"context"
	"fmt"
	"handler/repositories"
	"handler/rules"
	"log"
	"sync"
)

func initExec(startRules []string, ctx context.Context) {
	var wg sync.WaitGroup
	rules := ctx.Value("rules").(map[string]*rules.RuleStructure)
	for _, startId := range startRules {
		wg.Add(1)
		go prepRule(rules[startId], &wg, ctx, startId)
	}
	wg.Wait()
}

func prepRule(rule *rules.RuleStructure, wg *sync.WaitGroup, ctx context.Context, ruleId string) {
	defer wg.Done()
	if l, ok := ctx.Value("log").(*repositories.LogData); ok {
		l.ExecutionOrder = append(l.ExecutionOrder, ruleId)
	} else {
		log.Panic("method prepRule: could not type cast log model")
	}
	if err := execRule(rule, ctx); err != nil {
		addErrorToContext(err, ctx)
		return
	}
}

func execRule(rule *rules.RuleStructure, ctx context.Context) error {
	if ev, err := rule.Conditions.EvaluateGroup(ctx); err != nil {
		return err
	} else if ev {
		return handleActions(rule.Then, ctx)
	} else {
		return handleActions(rule.Else, ctx)
	}
}

func addErrorToContext(err error, ctx context.Context) error {
	if l, ok := ctx.Value("log").(*repositories.LogData); ok {
		l.AddExecLog("system", "error", err.Error())
	} else {
		return fmt.Errorf("method addErrorToContext: could not type cast log model")
	}
	return nil
}
