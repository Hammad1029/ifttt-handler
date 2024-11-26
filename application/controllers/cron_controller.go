package controllers

import (
	"context"
	"fmt"
	"ifttt/handler/application/core"
	"ifttt/handler/common"
	"ifttt/handler/domain/api"
	"ifttt/handler/domain/request_data"
	"ifttt/handler/domain/resolvable"
	"sync"
	"time"

	"github.com/google/uuid"
)

func NewCronController(cron *api.Cron, core *core.ServerCore, ctx context.Context) error {
	cronName := cron.Name
	if _, err := core.Cron.AddFunc(cron.Cron, func() {
		cronController(cronName, core, ctx)
	}); err != nil {
		return err
	}
	return nil
}

func cronController(cronJobName string, core *core.ServerCore, ctx context.Context) {
	startTime := time.Now()

	var contextState sync.Map
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)
	ctx = context.WithValue(ctx, common.ContextState, &contextState)

	tracer, err := uuid.NewRandom()
	if err != nil {
		core.Logger.Info(fmt.Sprintf(
			"Cron received: %s | Start time: %s", cronJobName, startTime.String(),
		))
		core.Logger.Error("could not assign tracer", err)
		return
	}

	contextState.Store(common.ContextLogger, core.Logger)
	contextState.Store(common.ContextTracer, tracer.String())
	common.LogWithTracer(common.LogSystem, fmt.Sprintf(
		"Cron recieved: %s | Start time: %s", cronJobName, startTime.String(),
	), nil, false, ctx)

	requestData := request_data.RequestData{}
	requestData.Initialize()
	resChan := make(chan resolvable.ResponseResolvable, 1)

	contextState.Store(common.ContextResponseChannel, resChan)
	contextState.Store(common.ContextExternalExecTime, uint64(0))
	contextState.Store(common.ContextRequestData, &requestData)
	contextState.Store(common.ContextResponseSent, false)

	go func(requestData *request_data.RequestData) {
		<-ctx.Done()
		end := time.Now()
		executionTime := uint64(end.Sub(startTime).Milliseconds())
		externalExecTime, ok := contextState.Load(common.ContextExternalExecTime)
		if !ok {
			return
		}
		internalExecTime := executionTime - externalExecTime.(uint64)
		common.LogWithTracer(common.LogSystem, "cron end", map[string]any{
			"start":            startTime,
			"end":              end,
			"executionTime":    executionTime,
			"internalExecTime": internalExecTime,
			"externalExecTime": externalExecTime,
			"requestData":      requestData,
		}, false, ctx)
	}(&requestData)

	job, err := core.CacheStore.CronCacheRepo.GetCronByName(cronJobName, ctx)
	if job == nil || err != nil {
		defer cancel(err)
		common.LogWithTracer(common.LogSystem,
			fmt.Sprintf("cron not found | path: %s", cronJobName), err, true, ctx)
		res := &resolvable.ResponseResolvable{
			ResponseCode:        "404",
			ResponseDescription: "Cron not found",
		}
		res.ManualSend(resChan, core.ResolvableDependencies, ctx)
		return
	}
	common.LogWithTracer(common.LogSystem,
		fmt.Sprintf("Cron found | name: %s | schedule: %s", job.Name, job.Cron),
		job, false, ctx)

	if err := core.PreparePreConfig(job.PreConfig, ctx); err != nil {
		defer cancel(err)
		common.LogWithTracer(common.LogSystem, "could not prepare pre config", err, true, ctx)
		res := &resolvable.ResponseResolvable{
			ResponseCode:        "500",
			ResponseDescription: "Could not prepare pre config",
		}
		res.ManualSend(resChan, core.ResolvableDependencies, ctx)
		return
	}

	go func() {
		defer cancel(nil)

		if err := core.InitMainWare(job.TriggerFlows, ctx); err != nil {
			cancel(err)
		}

		var res resolvable.ResponseResolvable
		if err := context.Cause(ctx); err != nil {
			res = resolvable.ResponseResolvable{
				ResponseCode:        common.ResponseCodeSystemError,
				ResponseDescription: common.ResponseDescriptionSystemError,
			}
			res.AddError(err)
		}
		res.ManualSend(resChan, core.ResolvableDependencies, ctx)
	}()

	<-resChan
}
