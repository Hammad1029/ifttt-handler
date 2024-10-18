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

func mainController(core *core.ServerCore, ctx context.Context) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {

		var contextState sync.Map
		ctx, cancel := context.WithCancel(ctx)
		ctx = context.WithValue(ctx, common.ContextState, &contextState)

		log := audit_log.AuditLog{}
		requestData := request_data.RequestData{}

		log.Initialize(c.Path(), &requestData)
		requestData.Initialize()

		go func(l *audit_log.AuditLog) {
			<-ctx.Done()
			l.EndLog()
			if err := core.ConfigStore.AuditLogRepo.InsertLog(l); err != nil {
				fmt.Printf("error in inserting log: %s\n", err)
			}
		}(&log)

		contextState.Store(common.ContextLog, &log)

		api, err := core.CacheStore.APICacheRepo.GetApiByPath(c.Path(), ctx)
		if api == nil || err != nil {
			defer cancel()
			res := &resolvable.ResponseResolvable{
				ResponseCode:        "404",
				ResponseDescription: "API not found",
			}
			return res.SendResponse(c)
		} else {
			log.ApiID = api.ID
			log.ApiPath = api.Path
			log.ApiName = api.Name
		}

		if err := c.BodyParser(&requestData.ReqBody); err != nil {
			defer cancel()
			res := &resolvable.ResponseResolvable{
				ResponseCode:        "400",
				ResponseDescription: "Error in parsing body",
			}
			return res.SendResponse(c)
		}

		resChan := make(chan resolvable.ResponseResolvable, 1)
		contextState.Store(common.ContextRequestData, &requestData)
		contextState.Store(common.ContextResponseChannel, resChan)
		contextState.Store(common.ContextApiData, api)

		if err := core.PreparePreConfig(api.PreConfig, ctx); err != nil {
			defer cancel()
			res := &resolvable.ResponseResolvable{
				ResponseCode:        "500",
				ResponseDescription: "Could not prepare pre config",
			}
			return res.SendResponse(c)
		}

		go func() {
			defer cancel()
			core.InitExec(api.TriggerFlows, ctx, c)
		}()

		res := <-resChan
		close(resChan)
		return res.SendResponse(c)
	}
}
