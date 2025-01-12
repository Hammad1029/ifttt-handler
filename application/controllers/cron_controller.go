package controllers

import (
	"context"
	"ifttt/handler/application/core"
	"ifttt/handler/domain/api"
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

func cronController(cronJobName string, core *core.ServerCore, parentCtx context.Context) {
	// startTime := time.Now()

	// requestData := request_data.RequestData{}
	// requestData.Initialize()
	// resChan := make(chan resolvable.Response, 1)

	// var contextState sync.Map
	// contextState.Store(common.ContextLogStage, common.LogStageInitation)
	// contextState.Store(common.ContextLogger, core.Logger)
	// contextState.Store(common.ContextResponseChannel, resChan)
	// contextState.Store(common.ContextExternalExecTime, uint64(0))
	// contextState.Store(common.ContextResponseSent, false)
	// contextState.Store(common.ContextRequestData, &requestData)

	// cancelCtx, cancel := context.WithCancelCause(parentCtx)
	// ctx := context.WithValue(cancelCtx, common.ContextState, &contextState)

	// go func(requestData *request_data.RequestData, ctxState *sync.Map) {
	// 	<-cancelCtx.Done()
	// 	ctxState.Store(common.ContextLogStage, common.LogStageEnding)
	// 	end := time.Now()
	// 	executionTime := uint64(end.Sub(startTime).Milliseconds())
	// 	externalExecTime, ok := contextState.Load(common.ContextExternalExecTime)
	// 	if !ok {
	// 		return
	// 	}
	// 	internalExecTime := executionTime - externalExecTime.(uint64)
	// 	cancelCause := context.Cause(cancelCtx)
	// 	common.LogWithTracer(common.LogSystem, "cron end",
	// 		map[string]any{
	// 			"start":            startTime,
	// 			"end":              end,
	// 			"executionTime":    executionTime,
	// 			"internalExecTime": internalExecTime,
	// 			"externalExecTime": externalExecTime,
	// 			"requestData":      requestData,
	// 			"error":            cancelCause,
	// 		}, cancelCause == nil, ctx)
	// }(&requestData, &contextState)

	// tracer, err := uuid.NewRandom()
	// if err != nil {
	// 	cancel(err)
	// 	core.Logger.Info(fmt.Sprintf(
	// 		"Cron received: %s | Start time: %s", cronJobName, startTime.String(),
	// 	))
	// 	core.Logger.Error("could not assign tracer", err)
	// 	res := &resolvable.Response{
	// 		ResponseCode:        "500",
	// 		ResponseDescription: "Could not assign tracer",
	// 	}
	// 	res.ManualSend(resChan, err, ctx)
	// 	return
	// } else {
	// 	common.LogWithTracer(common.LogSystem, fmt.Sprintf(
	// 		"Cron recieved: %s | Start time: %s", cronJobName, startTime.String(),
	// 	), nil, false, ctx)
	// 	contextState.Store(common.ContextTracer, tracer.String())
	// }

	// contextState.Store(common.ContextLogStage, common.LogStageMemload)
	// job, err := core.CacheStore.CronRepo.GetCronByName(cronJobName, ctx)
	// if job == nil || err != nil {
	// 	defer cancel(err)
	// 	common.LogWithTracer(common.LogSystem,
	// 		fmt.Sprintf("cron not found | path: %s", cronJobName), err, true, ctx)
	// 	res := &resolvable.Response{
	// 		ResponseCode:        "404",
	// 		ResponseDescription: "Cron not found",
	// 	}
	// 	res.ManualSend(resChan, nil, ctx)
	// 	return
	// }
	// common.LogWithTracer(common.LogSystem,
	// 	fmt.Sprintf("Cron found | name: %s | schedule: %s", job.Name, job.Cron),
	// 	job, false, ctx)

	// contextState.Store(common.ContextLogStage, common.LogStagePreConfig)
	// if err := core.PreparePreConfig(job.PreConfig, ctx); err != nil {
	// 	defer cancel(err)
	// 	common.LogWithTracer(common.LogSystem, "could not prepare pre config", err, true, ctx)
	// 	res := &resolvable.Response{
	// 		ResponseCode:        "500",
	// 		ResponseDescription: "Could not prepare pre config",
	// 	}
	// 	res.ManualSend(resChan, err, ctx)
	// 	return
	// }

	// go func() {
	// 	defer cancel(nil)

	// 	contextState.Store(common.ContextLogStage, common.LogStageExecution)
	// 	if err := core.InitExecution(job.TriggerFlows, ctx); err != nil {
	// 		cancel(err)
	// 	}

	// 	var res resolvable.Response
	// 	err := context.Cause(cancelCtx)
	// 	if err != nil {
	// 		res = resolvable.Response{
	// 			ResponseCode: common.ResponseCodes[common.ResponseCodeSystemMalfunction],
	// 		}
	// 	}
	// 	res.ManualSend(resChan, err, ctx)
	// }()

	// <-resChan
}
