package controllers

import (
	"context"
	"fmt"
	"ifttt/handler/application/core"
	"ifttt/handler/common"
	"ifttt/handler/domain/api"
	"ifttt/handler/domain/audit_log"
	"ifttt/handler/domain/request_data"
	"ifttt/handler/domain/resolvable"
	"sync"

	"github.com/fatih/structs"
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
	var contextState sync.Map
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)
	ctx = context.WithValue(ctx, common.ContextState, &contextState)

	log := audit_log.CronAuditLog{}
	requestData := request_data.RequestData{}

	go func(l *audit_log.CronAuditLog) {
		<-ctx.Done()
		l.EndLog()
		if err := core.ConfigStore.CronAuditLogRepo.InsertLog(l); err != nil {
			fmt.Printf("error in inserting log: %s\n", err)
		}
	}(&log)

	log.Initialize(cronJobName, &requestData)
	requestData.Initialize()

	var interfaceLog audit_log.AuditLog = &log
	contextState.Store(common.ContextLog, &interfaceLog)

	job, err := core.CacheStore.CronCacheRepo.GetCronByName(cronJobName, ctx)
	if job == nil || err != nil {
		defer cancel(err)
		res := &resolvable.ResponseResolvable{
			ResponseCode:        "404",
			ResponseDescription: "API not found",
		}
		log.SetFinalResponse(structs.Map(res))
		return
	}

	log.CronID = job.ID
	log.Name = job.Name

	resChan := make(chan resolvable.ResponseResolvable, 1)
	contextState.Store(common.ContextRequestData, &requestData)
	contextState.Store(common.ContextResponseChannel, resChan)

	if err := core.PreparePreConfig(job.PreConfig, ctx); err != nil {
		defer cancel(err)
		res := &resolvable.ResponseResolvable{
			ResponseCode:        "500",
			ResponseDescription: "Could not prepare pre config",
		}
		log.SetFinalResponse(structs.Map(res))
		close(resChan)
		return
	}

	go func() {
		defer cancel(nil)
		core.InitMainWare(job.TriggerFlows, ctx)
		res := &resolvable.ResponseResolvable{}
		if _, err := res.Resolve(ctx, core.ResolvableDependencies); err != nil {
			res = &resolvable.ResponseResolvable{
				ResponseCode:        "500",
				ResponseDescription: "Error in resolving response",
			}
			resChan <- *res
			cancel(err)
		}
	}()

	<-resChan
	close(resChan)
}
