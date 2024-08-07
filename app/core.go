package app

import (
	"context"
	"fmt"
	"handler/common"
	"handler/domain/api"
	"handler/domain/audit_log"
	"handler/domain/configuration"
	"handler/domain/resolvable"
	infraStore "handler/infrastructure/store"
	"log"
	"sync"
)

type core struct {
	ConfigStore       infraStore.ConfigStore
	DataStore         infraStore.DataStore
	CacheStore        infraStore.CacheStore
	UserConfiguration configuration.UserConfiguration
}

func newCore() (*core, error) {
	var serverCore core

	if configStore, err := infraStore.NewConfigStore(); err != nil {
		return nil, fmt.Errorf("method newCore: could not create config store: %s", err)
	} else {
		serverCore.ConfigStore = *configStore
	}
	if dataStore, err := infraStore.NewDataStore(); err != nil {
		return nil, fmt.Errorf("method newCore: could not create data store: %s", err)
	} else {
		serverCore.DataStore = *dataStore
	}
	if cacheStore, err := infraStore.NewCacheStore(); err != nil {
		return nil, fmt.Errorf("method newCore: could not create cache store: %s", err)
	} else {
		serverCore.CacheStore = *cacheStore
	}
	if config, err := serverCore.ConfigStore.ConfigRepo.GetUserConfigFromDb(); err != nil {
		return nil, fmt.Errorf("method newCore: could not get user configuration: %s", err)
	} else {
		serverCore.UserConfiguration = config
	}

	return &serverCore, nil
}

func (c *core) initExec(startRules []string, ctx context.Context) {
	var wg sync.WaitGroup
	rules := ctx.Value("rules").(map[string]*api.Rule)
	for _, startId := range startRules {
		wg.Add(1)
		go c.prepRule(rules[startId], &wg, ctx, startId)
	}
	wg.Wait()
}

func (c *core) prepRule(rule *api.Rule, wg *sync.WaitGroup, ctx context.Context, ruleId string) {
	defer wg.Done()
	if l, ok := ctx.Value("log").(*audit_log.AuditLog); ok {
		l.ExecutionOrder = append(l.ExecutionOrder, ruleId)
	} else {
		log.Panic("method prepRule: could not type cast log model")
	}
	if err := c.execRule(rule, ctx); err != nil {
		c.addErrorToContext(err, ctx)
		return
	}
}

func (c *core) execRule(rule *api.Rule, ctx context.Context) error {
	if ev, err := rule.Conditions.EvaluateGroup(ctx); err != nil {
		return err
	} else if ev {
		return c.handleResolvableArray(rule.Then, ctx)
	} else {
		return c.handleResolvableArray(rule.Else, ctx)
	}
}

func (c *core) addErrorToContext(err error, ctx context.Context) error {
	if l, ok := ctx.Value("log").(*audit_log.AuditLog); ok {
		l.AddExecLog("system", "error", err.Error())
	} else {
		return fmt.Errorf("method addErrorToContext: could not type cast log model")
	}
	return nil
}

func (c *core) handleResolvableArray(resolvables []resolvable.Resolvable, ctx context.Context) error {
	for _, r := range resolvables {
		if _, err := c.callResolvable(r, ctx); err != nil {
			return fmt.Errorf("method *core.handleResolvableArray: error in resolving: %s", err)
		}
	}
	return nil
}

func (c *core) callResolvable(resolvable resolvable.Resolvable, ctx context.Context) (any, error) {
	switch resolvable.ResolveType {
	case "rule":
		c.handleActionRule(ctx, resolvable.ResolveData)
	case "db":
		resolvable.Resolve(ctx, c.DataStore.RawQueryRepository)
	default:
		return resolvable.Resolve(ctx)
	}
	return nil, nil
}

func (c *core) handleActionRule(ctx context.Context, actionData common.JsonObject) {
	var wg sync.WaitGroup
	wg.Add(1)
	rules := ctx.Value("rules").(map[string]*api.Rule)
	ruleId := fmt.Sprint(actionData["value"])
	go c.prepRule(rules[ruleId], &wg, ctx, ruleId)
}
