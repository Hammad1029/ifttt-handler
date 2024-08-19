package core

import (
	"context"
	"fmt"
	"handler/common"
	"handler/domain/api"
	"handler/domain/audit_log"
	"handler/domain/configuration"
	"handler/domain/resolvable"
	infraStore "handler/infrastructure/store"
	"sync"

	"github.com/mitchellh/mapstructure"
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
	if config, err := serverCore.ConfigStore.ConfigRepo.GetConfigFromDb(); err != nil {
		return nil, fmt.Errorf("method newCore: could not get user configuration: %s", err)
	} else {
		serverCore.Configuration = config
	}
	serverCore.ResolvableDependencies = map[string]any{
		common.DependencyRawQueryRepo: serverCore.DataStore.RawQueryRepo,
	}

	return &serverCore, nil
}

func (c *ServerCore) InitExec(startRules []string, ctx context.Context) {
	var wg sync.WaitGroup
	for _, startId := range startRules {
		wg.Add(1)
		go c.prepRule(startId, &wg, ctx)
	}
	wg.Wait()
}

func (c *ServerCore) prepRule(ruleId string, wg *sync.WaitGroup, ctx context.Context) error {
	defer wg.Done()

	rules := ctx.Value("rules").(map[string]*api.Rule)
	currRule, ok := rules[ruleId]
	if !ok {
		return c.AddErrorToContext(fmt.Errorf("method *core.prepRule: rule %s not found", ruleId), ctx)
	}

	if l, ok := ctx.Value("log").(*audit_log.AuditLog); ok {
		l.ExecutionOrder = append(l.ExecutionOrder, ruleId)
	} else {
		return c.AddErrorToContext(fmt.Errorf("method *core.prepRule: could not type cast log model"), ctx)
	}

	if err := c.execRule(currRule, ctx); err != nil {
		return c.AddErrorToContext(fmt.Errorf("method *core.prepRule: error in rule %s execution: %s", ruleId, err), ctx)
	}

	return nil
}

func (c *ServerCore) execRule(rule *api.Rule, ctx context.Context) error {
	if ev, err := rule.Conditions.EvaluateGroup(ctx, c.ResolvableDependencies); err != nil {
		return err
	} else if ev {
		return c.handleResolvableArray(rule.Then, ctx)
	} else {
		return c.handleResolvableArray(rule.Else, ctx)
	}
}

func (c *ServerCore) AddErrorToContext(err error, ctx context.Context) error {
	if l, ok := ctx.Value("log").(*audit_log.AuditLog); ok {
		l.AddExecLog("system", "error", err.Error())
		errorResponse := &resolvable.Resolvable{
			ResolveType: resolvable.AccessorResponseResolvable,
			ResolveData: map[string]any{
				"responseCode": "500", "responseDescription": "Server Errror",
			}}
		c.callResolvable(errorResponse, ctx)
	} else {
		return fmt.Errorf("method *ServerCore.AddErrorToContext: could not type cast log model")
	}
	return nil
}

func (c *ServerCore) handleResolvableArray(resolvables []resolvable.Resolvable, ctx context.Context) error {
	for _, r := range resolvables {
		if _, err := c.callResolvable(&r, ctx); err != nil {
			return fmt.Errorf("method *core.handleResolvableArray: error in resolving: %s", err)
		}
	}
	return nil
}

func (c *ServerCore) callResolvable(r *resolvable.Resolvable, ctx context.Context) (any, error) {
	switch r.ResolveType {
	case resolvable.AccessorRuleResolvable:
		return nil, c.handleActionRule(ctx, r.ResolveData)
	default:
		return r.Resolve(ctx, c.ResolvableDependencies)
	}
}

func (c *ServerCore) handleActionRule(ctx context.Context, data map[string]any) error {
	var ruleRes resolvable.CallRuleResolvable
	if err := mapstructure.Decode(data, &ruleRes); err != nil {
		return fmt.Errorf("method *core.handleActionRule: could not decode resolvable data to resolvable.CallRuleResolvable: %s", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go c.prepRule(ruleRes.RuleId, &wg, ctx)
	return nil
}
