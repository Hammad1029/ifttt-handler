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
	"net/http"
	"strings"
	"sync"

	"github.com/fatih/structs"
	"github.com/gofiber/fiber/v2"
)

func NewMainController(router fiber.Router, core *core.ServerCore, api *api.Api, ctx context.Context) error {
	controller := mainController(core, ctx)
	switch strings.ToUpper(api.Method) {
	case http.MethodGet:
		router.Get(api.Path, controller)
	case http.MethodPost:
		router.Post(api.Path, controller)
	case http.MethodPut:
		router.Put(api.Path, controller)
	case http.MethodDelete:
		router.Delete(api.Path, controller)
	default:
		return fmt.Errorf("method NewMainController: method %s not found", api.Method)
	}
	return nil
}

func mainController(core *core.ServerCore, parentCtx context.Context) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var contextState sync.Map
		ctx, cancel := context.WithCancelCause(parentCtx)
		defer cancel(nil)
		ctx = context.WithValue(ctx, common.ContextState, &contextState)

		go func() {
			<-c.Context().Done()
			cancel(nil)
		}()

		log := audit_log.APIAuditLog{}
		requestData := request_data.RequestData{}

		go func(l *audit_log.APIAuditLog) {
			<-ctx.Done()
			l.EndLog()
			if err := core.ConfigStore.APIAuditLogRepo.InsertLog(l); err != nil {
				fmt.Printf("error in inserting log: %s\n", err)
			}
		}(&log)

		log.Initialize(c.Path(), &requestData)
		requestData.Initialize()

		var interfaceLog audit_log.AuditLog = &log
		contextState.Store(common.ContextLog, &interfaceLog)

		api, err := core.CacheStore.APICacheRepo.GetApiByPath(c.Path(), ctx)
		if api == nil || err != nil {
			defer cancel(err)
			res := &resolvable.ResponseResolvable{
				ResponseCode:        "404",
				ResponseDescription: "API not found",
			}
			log.SetFinalResponse(structs.Map(res))
			return c.JSON(res)
		}

		log.ApiID = api.ID
		log.ApiPath = api.Path
		log.ApiName = api.Name

		if err := c.BodyParser(&requestData.ReqBody); err != nil {
			defer cancel(err)
			res := &resolvable.ResponseResolvable{
				ResponseCode:        "400",
				ResponseDescription: "Error in parsing body",
			}
			log.SetFinalResponse(structs.Map(res))
			return c.JSON(res)
		}

		resChan := make(chan resolvable.ResponseResolvable, 1)
		contextState.Store(common.ContextRequestData, &requestData)
		contextState.Store(common.ContextResponseChannel, resChan)

		if err := core.PreparePreConfig(api.PreConfig, ctx); err != nil {
			defer cancel(err)
			res := &resolvable.ResponseResolvable{
				ResponseCode:        "500",
				ResponseDescription: "Could not prepare pre config",
			}
			log.SetFinalResponse(structs.Map(res))
			close(resChan)
			return c.JSON(res)
		}

		go func() {
			defer cancel(nil)
			core.InitExec(api.TriggerFlows, ctx)
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

		res := <-resChan
		close(resChan)
		return c.JSON(res)
	}
}
