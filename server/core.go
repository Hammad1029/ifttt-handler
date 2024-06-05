package server

import (
	"context"
	"encoding/json"
	"errors"
	"handler/models"
	redisUtils "handler/rediisUtils"
	"log"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

func getApiFromRedis(c *fiber.Ctx) (*models.ApiModel, error) {
	var api models.ApiModel
	apiName := c.Params("apiName")
	apiJSON, err := redisUtils.GetRedis().HGet(c.Context(), "apis", apiName).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(apiJSON), &api)
	if err != nil {
		return nil, err
	}
	return &api, nil
}

func initExec(startRules []int, ctx context.Context) {
	var wg sync.WaitGroup
	rules := ctx.Value("rules").([]models.RuleUDT)
	for _, startIdx := range startRules {
		wg.Add(1)
		go prepRule(rules[startIdx], &wg, ctx, startIdx)
	}
	wg.Wait()
}

func prepRule(rule models.RuleUDT, wg *sync.WaitGroup, ctx context.Context, ruleIdx int) {
	defer wg.Done()
	log := ctx.Value("log").(models.LogModel)
	log.ExecutionOrder = append(log.ExecutionOrder, ruleIdx)
	if err := execRule(rule, ctx); err != nil {
		addErrorToContext(err, ctx, true)
		return
	}
}

func execRule(rule models.RuleUDT, ctx context.Context) error {
	if ev, err := evaluateCondition(rule, ctx); err != nil {
		return err
	} else if ev {
		return handleActions(rule.Then, ctx)
	} else {
		return handleActions(rule.Else, ctx)
	}
}

func evaluateCondition(rule models.RuleUDT, ctx context.Context) (bool, error) {
	evaluators := getEvaluators()
	if evalFunc, ok := evaluators[rule.Operand]; ok {
		op1Res, err := rule.Operator1.Resolve(ctx)
		if err != nil {
			return false, err
		}
		op2Res, err := rule.Operator2.Resolve(ctx)
		if err != nil {
			return false, err
		}

		return evalFunc(op1Res, op2Res), nil
	} else {
		return false, errors.New("operand not found")
	}
}

func getEvaluators() map[string]func(a, b string) bool {
	return map[string]func(a, b string) bool{
		"eq": func(a, b string) bool {
			return a == b
		},
	}
}

func addErrorToContext(err error, ctx context.Context, sendRes bool) {
	if l, ok := ctx.Value("log").(models.LogModel); ok {
		l.AddExecLog("system", "error", err.Error())
	} else {
		log.Panic("could not type cast request data")
	}

	if sendRes {
		sendResponse(ctx, "500")
	}
}

func sendResponse(ctx context.Context, resCode string) error {
	if responseChannel, ok := ctx.Value("resChan").(chan string); ok {
		responseChannel <- resCode
		return nil
	} else {
		return errors.New("send res type assertion failed")
	}
}
