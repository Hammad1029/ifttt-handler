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

		log := audit_log.APIAuditLog{}
		requestData := request_data.RequestData{}

		go func(l *audit_log.APIAuditLog) {
			<-ctx.Done()
			l.EndLog()
			if err := core.ConfigStore.APIAuditLogRepo.InsertLog(l); err != nil {
				fmt.Printf("token: %s | user: %s | type: %s | log: %s\n",
					l.GetRequestToken(), common.LogSystem, common.LogError,
					fmt.Sprintf("error in inserting log: %s", err))
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
			log.SetResponse(res.ResponseCode, res.ResponseDescription, structs.Map(res.Response))
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
			log.SetResponse(res.ResponseCode, res.ResponseDescription, structs.Map(res.Response))
			return c.JSON(res)
		}
		requestData.Headers = c.GetReqHeaders()

		resChan := make(chan resolvable.ResponseResolvable, 1)
		contextState.Store(common.ContextRequestData, &requestData)
		contextState.Store(common.ContextResponseChannel, resChan)

		if err := core.PreparePreConfig(api.PreConfig, ctx); err != nil {
			defer cancel(err)
			res := &resolvable.ResponseResolvable{
				ResponseCode:        "500",
				ResponseDescription: "Could not prepare pre config",
			}
			log.SetResponse(res.ResponseCode, res.ResponseDescription, structs.Map(res.Response))
			close(resChan)
			return c.JSON(res)
		}

		go func(l *audit_log.APIAuditLog) {
			defer cancel(nil)

			if err := core.InitMiddleWare(api.PreWare, ctx); err != nil {
				cancel(err)
			} else if err := core.InitMainWare(api.MainWare, ctx); err != nil {
				cancel(err)
			} else if err := core.InitMiddleWare(api.PostWare, ctx); err != nil {
				cancel(err)
			}

			if !l.ResponseSent {
				res := resolvable.ResponseResolvable{
					ResponseCode:        common.ResponseCodeSuccess,
					ResponseDescription: common.ResponseDescriptionSuccess,
				}
				if _, err := res.Resolve(ctx, core.ResolvableDependencies); err != nil {
					l.AddExecLog(common.LogSystem, common.LogError, err.Error())
					res.ResponseCode = common.ResponseCodeSystemError
					res.ResponseDescription = common.ResponseDescriptionSystemError
					l.SetResponse(res.ResponseCode, res.ResponseDescription, structs.Map(res.Response))
					if ok := l.SetResponseSent(); ok {
						resChan <- res
					}
				}
			}
		}(&log)

		res := <-resChan
		audit_log.AddExecLog(common.LogSystem, common.LogInfo,
			fmt.Sprintf("Sending response | Response Code: %s | Response Description: %s",
				res.ResponseCode, res.ResponseDescription), ctx)
		close(resChan)
		return c.JSON(res)
	}
}
