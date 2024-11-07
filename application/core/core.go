package core

import (
	"context"
	"fmt"
	"ifttt/handler/common"
	"ifttt/handler/domain/api"
	"ifttt/handler/domain/audit_log"
	"ifttt/handler/domain/configuration"
	"ifttt/handler/domain/resolvable"
	infraStore "ifttt/handler/infrastructure/store"
	"sync"

	"github.com/robfig/cron/v3"
)

type ServerCore struct {
	Cron                   *cron.Cron
	ConfigStore            *infraStore.ConfigStore
	DataStore              *infraStore.DataStore
	CacheStore             *infraStore.CacheStore
	AppCacheStore          *infraStore.AppCacheStore
	Configuration          *configuration.Configuration
	ResolvableDependencies map[common.IntIota]any
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
	serverCore.ResolvableDependencies = map[common.IntIota]any{
		common.DependencyRawQueryRepo: serverCore.DataStore.RawQueryRepo,
		common.DependencyAppCacheRepo: serverCore.AppCacheStore.AppCacheRepo,
		common.DependencyDbDumpRepo:   serverCore.DataStore.DumpRepo,
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

func (c *ServerCore) InitMiddleWare(triggerFlows *[]api.TriggerFlow, ctx context.Context) error {
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	audit_log.AddExecLog(common.LogSystem, common.LogInfo, "initiating pre/post-ware", ctx)
	var flowWG sync.WaitGroup
	for _, flow := range *triggerFlows {
		flowWG.Add(1)
		go func(f *api.TriggerFlow) {
			defer flowWG.Done()
			select {
			case <-ctx.Done():
				return
			default:
				{
					audit_log.AddExecLog(common.LogSystem, common.LogInfo, fmt.Sprintf("initiating trigger %d | %s", f.ID, f.Name), ctx)
					if err := c.execRule(
						f.StartState, f.BranchFlows, f.Rules, f.ID, ctx,
					); err != nil {
						cancel(err)
					}
				}
			}
		}(&flow)
	}
	flowWG.Wait()
	audit_log.AddExecLog(common.LogSystem, common.LogInfo, "pre/post-ware finished executing", ctx)

	if err := context.Cause(ctx); err != nil {
		audit_log.AddExecLog(common.LogSystem, common.LogError, err, ctx)
		return err
	}
	return nil
}

func (c *ServerCore) InitMainWare(triggerFlows *[]api.TriggerCondition, ctx context.Context) error {
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	audit_log.AddExecLog(common.LogSystem, common.LogInfo, "initiating mainware", ctx)
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
					if log := audit_log.GetAuditLogFromContext(ctx); log != nil {
						(*log).InitExecOrder(f.Trigger.ID)
					}
					audit_log.AddExecLog(common.LogSystem, common.LogInfo, fmt.Sprintf("initiating trigger %d | %s", f.Trigger.ID, f.Trigger.Name), ctx)
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
	audit_log.AddExecLog(common.LogSystem, common.LogInfo, "mainware finished executing", ctx)

	if err := context.Cause(ctx); err != nil {
		audit_log.AddExecLog(common.LogSystem, common.LogError, err, ctx)
		return err
	}
	return nil
}

func (c *ServerCore) execRule(
	state uint, branchMap map[uint]*api.BranchFlow, rules map[uint]*api.Rule, flowId uint, ctx context.Context,
) error {
	audit_log.AddExecLog(common.LogSystem, common.LogInfo, fmt.Sprintf("executing state %d", state), ctx)

	execState := audit_log.ExecState{State: state}

	branch, ok := branchMap[state]
	if ok {
		execState.Rule = branch.Rule
	} else {
		audit_log.AddExecLog(common.LogSystem, common.LogInfo, fmt.Sprintf("no branch found for state %d-> ending trigger", state), ctx)
	}

	if log := audit_log.GetAuditLogFromContext(ctx); log != nil {
		(*log).AddExecState(execState, flowId)
	}

	if !ok {
		return nil
	}

	rule, ok := rules[branch.Rule]
	if !ok {
		return fmt.Errorf("rule %d for state %d not found", branch.Rule, state)
	} else {
		audit_log.AddExecLog(common.LogSystem, common.LogInfo, fmt.Sprintf("executing rule %d | %s", rule.ID, rule.Name), ctx)
	}

	if err := c.resolveArray(rule.Pre, ctx); err != nil {
		return fmt.Errorf("could not resolve pre: %s", err)
	}

	rVal, err := c.solveRuleSwitch(&rule.Switch, ctx)
	if err != nil {
		return fmt.Errorf("could not solve switch: %s", err)
	}

	nextState, ok := branch.States[rVal.(uint)]
	if !ok {
		audit_log.AddExecLog(common.LogSystem, common.LogInfo, fmt.Sprintf("no next state for rVal %d -> ending trigger", rVal), ctx)
		return nil
	}

	return c.execRule(nextState, branchMap, rules, flowId, ctx)
}

func (c *ServerCore) solveRuleSwitch(s *api.RuleSwitch, ctx context.Context) (any, error) {
	for _, currCase := range s.Cases {
		if ev, err := currCase.Condition.EvaluateGroup(ctx, c.ResolvableDependencies); err != nil {
			return nil, fmt.Errorf("method solveRuleSwitch: error in solving case: %s", err)
		} else if ev {
			if rVal, err := c.doRuleCase(&currCase, ctx); err != nil {
				return nil, err
			} else {
				return rVal, nil
			}
		}
	}
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
		audit_log.AddExecLog(common.LogSystem, common.LogInfo, fmt.Sprintf("resolving %s", r.ResolveType), ctx)
		if _, err := r.Resolve(ctx, c.ResolvableDependencies); err != nil {
			return err
		}
	}
	return nil
}
