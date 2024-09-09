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
	var (
		mu    sync.Mutex
		wg    sync.WaitGroup
		errCh = make(chan error, 1)
	)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for key, r := range config {
		wg.Add(1)
		go func(key string, r resolvable.Resolvable) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				return
			default:
				if val, err := r.Resolve(ctx, c.ResolvableDependencies); err != nil {
					mu.Lock()
					select {
					case errCh <- err:
						cancel()
					default:
					}
					mu.Unlock()
				} else {
					mu.Lock()
					preConfig[key] = val
					mu.Unlock()
				}
			}
		}(key, r)
	}

	wg.Wait()

	select {
	case err := <-errCh:
		return err
	default:
		return nil
	}
}

func (c *ServerCore) InitExec(triggerFlows *[]api.TriggerFlow, ctx context.Context) {
	var fwg sync.WaitGroup
	for _, flow := range *triggerFlows {
		go c.startTriggerFlow(&flow, &fwg, ctx)
	}
	fwg.Wait()
}

func (c *ServerCore) startTriggerFlow(tFlow *api.TriggerFlow, fwg *sync.WaitGroup, ctx context.Context) {
	fwg.Add(1)
	defer fwg.Done()

	var rwg sync.WaitGroup
	for _, startRule := range tFlow.StartRules {
		go c.prepRule(startRule, &rwg, tFlow, ctx)
	}
	rwg.Wait()
}

func (c *ServerCore) prepRule(ruleId uint, rwg *sync.WaitGroup, tFlow *api.TriggerFlow, ctx context.Context) error {
	rwg.Add(1)
	defer rwg.Done()

	currRule, ok := tFlow.AllRules[ruleId]
	if !ok {
		return c.AddErrorToContext(fmt.Errorf("method *core.prepRule: rule %d not found", ruleId), ctx)
	}

	if l, ok := ctx.Value("log").(*audit_log.AuditLog); ok {
		l.ExecutionOrder = append(l.ExecutionOrder, ruleId)
	} else {
		return c.AddErrorToContext(fmt.Errorf("method *core.prepRule: could not type cast log model"), ctx)
	}

	if err := c.execRule(currRule, tFlow, ctx); err != nil {
		return c.AddErrorToContext(fmt.Errorf("method *core.prepRule: error in rule %d execution: %s", ruleId, err), ctx)
	}

	return nil
}

func (c *ServerCore) execRule(rule *api.Rule, tFlow *api.TriggerFlow, ctx context.Context) error {
	if ev, err := rule.Conditions.EvaluateGroup(ctx, c.ResolvableDependencies); err != nil {
		return err
	} else if ev {
		return c.handleResolvableArray(rule.Then, tFlow, ctx)
	} else {
		return c.handleResolvableArray(rule.Else, tFlow, ctx)
	}
}

func (c *ServerCore) AddErrorToContext(err error, ctx context.Context) error {
	if l, ok := ctx.Value("log").(*audit_log.AuditLog); ok {
		l.AddExecLog("system", "error", err.Error())
		errorResponse := resolvable.ResponseResolvable{
			ResponseCode:        "500",
			ResponseDescription: "Server Errror",
		}
		errorResponse.Resolve(ctx, c.ResolvableDependencies)
	} else {
		return fmt.Errorf("method *ServerCore.AddErrorToContext: could not type cast log model")
	}
	return nil
}

func (c *ServerCore) handleResolvableArray(resolvables []resolvable.Resolvable, tFlow *api.TriggerFlow, ctx context.Context) error {
	var rwg sync.WaitGroup
	for _, r := range resolvables {
		switch r.ResolveType {
		case resolvable.AccessorRuleResolvable:
			ruleId, _ := r.Resolve(ctx, c.ResolvableDependencies)
			if ruleIdUint, ok := ruleId.(uint); ok {
				go c.prepRule(ruleIdUint, &rwg, tFlow, ctx)
				return nil
			}
			return fmt.Errorf("method *core.handleResolvableArray: error getting/casting rule id")
		default:
			if _, err := r.Resolve(ctx, c.ResolvableDependencies); err != nil {
				return fmt.Errorf("method *core.handleResolvableArray: error in resolving: %s", err)
			}
		}
	}
	rwg.Wait()
	return nil
}
