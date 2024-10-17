package core

import (
	"context"
	"fmt"
	"ifttt/handler/common"
	"ifttt/handler/domain/api"
	"ifttt/handler/domain/configuration"
	"ifttt/handler/domain/resolvable"
	infraStore "ifttt/handler/infrastructure/store"
	"sync"

	"github.com/gofiber/fiber/v2"
)

type ServerCore struct {
	ConfigStore            *infraStore.ConfigStore
	DataStore              *infraStore.DataStore
	CacheStore             *infraStore.CacheStore
	Configuration          *configuration.Configuration
	ResolvableDependencies map[string]any
}

func NewServerCore() (*ServerCore, error) {
	var serverCore ServerCore

	if configStore, err := infraStore.NewConfigStore(); err != nil {
		return nil, fmt.Errorf("method newCore: could not create config store: %s", err)
	} else {
		serverCore.ConfigStore = configStore
	}
	if dataStore, err := infraStore.NewDataStore(); err != nil {
		return nil, fmt.Errorf("method newCore: could not create data store: %s", err)
	} else {
		serverCore.DataStore = dataStore
	}
	if cacheStore, err := infraStore.NewCacheStore(); err != nil {
		return nil, fmt.Errorf("method newCore: could not create cache store: %s", err)
	} else {
		serverCore.CacheStore = cacheStore
	}

	serverCore.ResolvableDependencies = map[string]any{
		common.DependencyRawQueryRepo: serverCore.DataStore.RawQueryRepo,
	}

	return &serverCore, nil
}

func (c *ServerCore) PreparePreConfig(config map[string]resolvable.Resolvable, ctx context.Context) error {
	preConfig := resolvable.GetRequestData(ctx).PreConfig
	var wg sync.WaitGroup

	cancelCtx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	for key, r := range config {
		wg.Add(1)
		go func(key string, r resolvable.Resolvable) {
			defer wg.Done()
			select {
			case <-cancelCtx.Done():
				return
			default:
				if val, err := r.Resolve(cancelCtx, c.ResolvableDependencies); err != nil {
					cancel(err)
					return
				} else {
					preConfig[key] = val
				}
			}
		}(key, r)
	}

	wg.Wait()
	return context.Cause(ctx)
}

func (c *ServerCore) InitExec(triggerFlows *[]api.TriggerCondition, ctx context.Context, fiberCtx *fiber.Ctx) {
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	var flowWG sync.WaitGroup

	for _, flow := range *triggerFlows {
		flowWG.Add(1)
		go func(f *api.TriggerCondition) {
			defer flowWG.Done()
			select {
			case <-ctx.Done():
				return
			default:
				if err := c.initTriggerFlow(f, ctx); err != nil {
					cancel(err)
				}
			}
		}(&flow)
	}

	flowWG.Wait()

	var res resolvable.ResponseResolvable
	if err := context.Cause(ctx); err != nil {
		res = resolvable.ResponseResolvable{
			ResponseCode:        "500",
			ResponseDescription: "Server Errror",
		}
	}
	if _, err := res.Resolve(ctx, c.ResolvableDependencies); err != nil {
		fiberCtx.Send([]byte("something very bad happened"))
	}
}

func (c *ServerCore) initTriggerFlow(tFlow *api.TriggerCondition, ctx context.Context) error {
	ctx, cancel := context.WithCancelCause(ctx)

	if ev, err := tFlow.If.EvaluateGroup(ctx, c.ResolvableDependencies); err != nil {
		return fmt.Errorf("method initTriggerFlow: error in solving tFlow if: %s", err)
	} else if !ev {
		return nil
	}

	startRules := []*api.Rule{}
	for _, rId := range tFlow.Trigger.StartRules {
		rIdUint := uint(rId)
		currRule, ok := tFlow.Trigger.AllRules[rIdUint]
		if !ok {
			return fmt.Errorf("method *core.prepRule: rule %d not found", rId)
		}
		startRules = append(startRules, currRule)
	}

	var ruleWG sync.WaitGroup
	for _, rule := range startRules {
		ruleWG.Add(1)
		go func(r *api.Rule) {
			defer ruleWG.Done()
			select {
			case <-ctx.Done():
				return
			default:
				if err := c.execRule(r, tFlow.Trigger.BranchFlows, tFlow.Trigger.AllRules, ctx); err != nil {
					cancel(err)
				}
			}
		}(rule)
	}

	ruleWG.Wait()
	return context.Cause(ctx)
}

func (c *ServerCore) execRule(
	rule *api.Rule, branchMap map[uint]*[]api.BranchFlow, allRules map[uint]*api.Rule, ctx context.Context,
) error {
	if err := c.resolveArray(rule.Pre, ctx); err != nil {
		return fmt.Errorf("method *core.execRule: could not resolve pre: %s", err)
	}

	rVal, err := c.solveRuleSwitch(&rule.Switch, ctx)
	if err != nil {
		return fmt.Errorf("method *core.execRule: could not solve switch: %s", err)
	}

	branchFlow, ok := branchMap[rule.Id]
	if !ok {
		return fmt.Errorf("method *core.execRule: branch flow for rule %d not found", rule.Id)
	}

	for _, bF := range *branchFlow {
		if requiredRVal, err := bF.IfReturn.Resolve(ctx, c.ResolvableDependencies); err != nil {
			return fmt.Errorf("method *core.execRule: could not resolve IfReturn for branch: %s", err)
		} else if rVal == requiredRVal {
			if nextRule, ok := allRules[bF.Jump]; !ok {
				return fmt.Errorf("method *core.execRule: rule %d not found", bF.Jump)
			} else if err := c.execRule(nextRule, branchMap, allRules, ctx); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *ServerCore) solveRuleSwitch(s *api.RuleSwitch, ctx context.Context) (any, error) {
	for _, currCase := range s.Cases {
		if ev, err := currCase.Condition.EvaluateGroup(ctx, c.ResolvableDependencies); err != nil {
			return nil, fmt.Errorf("method solveRuleSwitch: error in solving case: %s", err)
		} else if ev {
			if rVal, err := c.doRuleCase(&currCase, ctx); err != nil {
				return nil, fmt.Errorf("method solveRuleSwitch: error in solving case: %s", err)
			} else {
				return rVal, nil
			}
		}
	}
	if rVal, err := c.doRuleCase(&s.Default, ctx); err != nil {
		return nil, fmt.Errorf("method solveRuleSwitch: error in solving default: %s", err)
	} else {
		return rVal, nil
	}
}

func (c *ServerCore) doRuleCase(s *api.RuleSwitchCase, ctx context.Context) (any, error) {
	if err := c.resolveArray(s.Do, ctx); err != nil {
		return nil, fmt.Errorf("method solveRuleSwitch: error in resolving do: %s", err)
	}
	if rVal, err := s.Return.Resolve(ctx, c.ResolvableDependencies); err != nil {
		return nil, fmt.Errorf("method solveRuleSwitch: error in resolving return: %s", err)
	} else {
		return rVal, nil
	}
}

func (c *ServerCore) resolveArray(resolvables []resolvable.Resolvable, ctx context.Context) error {
	for _, r := range resolvables {
		if _, err := r.Resolve(ctx, c.ResolvableDependencies); err != nil {
			return fmt.Errorf("method *core.resolveArray: error in resolving: %s", err)
		}
	}
	return nil
}
