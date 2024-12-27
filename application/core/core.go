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

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

type ServerCore struct {
	Cron                   *cron.Cron
	ConfigStore            *infraStore.ConfigStore
	DataStore              *infraStore.DataStore
	CacheStore             *infraStore.CacheStore
	AppCacheStore          *infraStore.AppCacheStore
	Configuration          *configuration.Configuration
	ResolvableDependencies map[common.IntIota]any
	Logger                 *logrus.Logger
}

func NewServerCore() (*ServerCore, error) {
	var serverCore ServerCore

	serverCore.Cron = cron.New()
	if configStore, err := infraStore.NewConfigStore(); err != nil {
		return nil, err
	} else {
		serverCore.ConfigStore = configStore
	}
	if dataStore, err := infraStore.NewDataStore(); err != nil {
		return nil, err
	} else {
		serverCore.DataStore = dataStore
	}
	if cacheStore, err := infraStore.NewCacheStore(); err != nil {
		return nil, err
	} else {
		serverCore.CacheStore = cacheStore
	}
	if appCacheStore, err := infraStore.NewAppCacheStore(); err != nil {
		return nil, err
	} else {
		serverCore.AppCacheStore = appCacheStore
	}
	logger := common.CreateLogrus()
	serverCore.Logger = logger
	serverCore.ResolvableDependencies = map[common.IntIota]any{
		common.DependencyRawQueryRepo: serverCore.DataStore.RawQueryRepo,
		common.DependencyAppCacheRepo: serverCore.AppCacheStore.AppCacheRepo,
		common.DependencyDbDumpRepo:   serverCore.DataStore.DumpRepo,
		common.DependencyOrmCacheRepo: serverCore.CacheStore.OrmRepo,
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
					preConfig.Store(key, val)
				}
			}
		}(key, r)
	}

	wg.Wait()
	return context.Cause(ctx)
}

func (c *ServerCore) InitExecution(triggerFlows *[]api.TriggerCondition, ctx context.Context) error {
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	common.LogWithTracer(common.LogSystem, "initiating triggers", nil, false, ctx)
	var flowWG sync.WaitGroup
	for _, flow := range *triggerFlows {
		flowWG.Add(1)
		go func(f *api.TriggerCondition) {
			defer flowWG.Done()
			select {
			case <-ctx.Done():
				return
			default:
				{
					common.LogWithTracer(common.LogSystem,
						fmt.Sprintf("initiating trigger %d | %s", f.Trigger.ID, f.Trigger.Name), nil, false, ctx)
					if ev, err := f.If.EvaluateGroup(ctx, c.ResolvableDependencies); err != nil {
						cancel(err)
					} else if ev {
						if err := c.execRule(
							f.Trigger.StartState, f.Trigger.BranchFlows, f.Trigger.Rules, f.Trigger.ID, ctx,
						); err != nil {
							cancel(err)
						}
					}
				}
			}
		}(&flow)
	}
	flowWG.Wait()
	common.LogWithTracer(common.LogSystem, "triggers finished executing", nil, false, ctx)

	if err := context.Cause(ctx); err != nil && err != context.Canceled {
		common.LogWithTracer(common.LogSystem, "error in triggers", err, true, ctx)
		return err
	}
	return nil
}

func (c *ServerCore) execRule(
	state uint, branchMap map[uint]*api.BranchFlow, rules map[uint]*api.Rule, triggerId uint, ctx context.Context,
) error {
	common.LogWithTracer(common.LogSystem,
		fmt.Sprintf("trigger %d: executing state %d", triggerId, state), nil, false, ctx)

	branch, ok := branchMap[state]
	if !ok {
		common.LogWithTracer(common.LogSystem,
			fmt.Sprintf("trigger %d: no branch found for state %d-> ending trigger", triggerId, state), nil, false, ctx)
		return nil
	}

	rule, ok := rules[branch.Rule]
	if !ok {
		return fmt.Errorf("trigger %d: rule %d for state %d not found", triggerId, branch.Rule, state)
	} else {
		common.LogWithTracer(common.LogSystem,
			fmt.Sprintf("trigger %d: executing rule %d | %s", triggerId, rule.ID, rule.Name), nil, false, ctx)
	}

	common.LogWithTracer(common.LogSystem,
		fmt.Sprintf("trigger %d rule %d | executing Pre", triggerId, rule.ID), nil, false, ctx)
	if err := c.resolveArray(rule.Pre, ctx); err != nil {
		return fmt.Errorf("could not resolve pre: %s", err)
	}

	common.LogWithTracer(common.LogSystem,
		fmt.Sprintf("trigger %d rule %d | evaluating cases", triggerId, rule.ID), nil, false, ctx)
	rVal, err := c.solveRuleSwitch(&rule.Switch, triggerId, rule.ID, ctx)
	if err != nil {
		return fmt.Errorf("could not solve switch: %s", err)
	}

	common.LogWithTracer(common.LogSystem,
		fmt.Sprintf("trigger %d rule %d | executing finally", triggerId, rule.ID), nil, false, ctx)
	if err := c.resolveArray(rule.Finally, ctx); err != nil {
		return fmt.Errorf("could not resolve finally: %s", err)
	}

	nextState, ok := branch.States[rVal.(uint)]
	if !ok {
		common.LogWithTracer(common.LogSystem,
			fmt.Sprintf("trigger %d: no next state for rVal %d -> ending trigger", triggerId, rVal), nil, false, ctx)
		return nil
	}

	common.LogWithTracer(common.LogSystem,
		fmt.Sprintf("trigger %d rule %d | moving to next state %d", triggerId, rule.ID, nextState), nil, false, ctx)
	return c.execRule(nextState, branchMap, rules, triggerId, ctx)
}

func (c *ServerCore) solveRuleSwitch(s *api.RuleSwitch, triggerId uint, ruleId uint, ctx context.Context) (any, error) {
	for idx, currCase := range s.Cases {
		if ev, err := currCase.Condition.EvaluateGroup(ctx, c.ResolvableDependencies); err != nil {
			return nil, fmt.Errorf("method solveRuleSwitch: error in solving case: %s", err)
		} else if ev {
			common.LogWithTracer(common.LogSystem,
				fmt.Sprintf("trigger %d rule %d | case %d matched", triggerId, ruleId, idx), nil, false, ctx)
			if rVal, err := c.doRuleCase(&currCase, ctx); err != nil {
				return nil, err
			} else {
				return rVal, nil
			}
		}
	}
	common.LogWithTracer(common.LogSystem,
		fmt.Sprintf("trigger %d rule %d | no case matched. performing default", triggerId, ruleId), nil, false, ctx)
	if rVal, err := c.doRuleCase(&s.Default, ctx); err != nil {
		return nil, err
	} else {
		return rVal, nil
	}
}

func (c *ServerCore) doRuleCase(s *api.RuleSwitchCase, ctx context.Context) (uint, error) {
	if err := c.resolveArray(s.Do, ctx); err != nil {
		return 0, err
	}
	return s.Return, nil
}

func (c *ServerCore) resolveArray(resolvables []resolvable.Resolvable, ctx context.Context) error {
	for _, r := range resolvables {
		common.LogWithTracer(common.LogSystem, fmt.Sprintf("resolving %s", r.ResolveType), r, false, ctx)
		if _, err := r.Resolve(ctx, c.ResolvableDependencies); err != nil {
			return err
		}
	}
	return nil
}
