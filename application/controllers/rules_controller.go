package controllers

import (
	"fmt"
	"ifttt/handler/application/core"
	"ifttt/handler/domain/api"
	"ifttt/handler/domain/audit_log"
	"ifttt/handler/domain/request_data"
	"ifttt/handler/domain/resolvable"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func NewRulesController(router fiber.Router, core *core.ServerCore, api *api.Api) error {
	controller := rulesController(core)
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
		return fmt.Errorf("method NewRulesController: method %s not found", api.Method)
	}
	return nil
}

func rulesController(core *core.ServerCore) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()

		log := audit_log.AuditLog{}
		ctx.SetUserValue("log", &log)
		log.StartLog()

		api, err := core.CacheStore.APICacheRepo.GetApiByPath(c.Path(), ctx)
		if api == nil || err != nil {
			res := &resolvable.ResponseResolvable{
				ResponseCode:        "404",
				ResponseDescription: "API not found",
			}
			return res.SendResponse(c)
		}

		requestData := request_data.RequestData{}
		requestData.Initialize()
		// log.Initialize(&requestData, api.Group, api.Name)

		err = c.BodyParser(&requestData.ReqBody)
		if err != nil {
			res := &resolvable.ResponseResolvable{
				ResponseCode:        "400",
				ResponseDescription: "Error in parsing body",
			}
			return res.SendResponse(c)
		}

		resChan := make(chan resolvable.ResponseResolvable, 1)
		ctx.SetUserValue("request", &requestData)
		ctx.SetUserValue("resChan", resChan)
		ctx.SetUserValue("api", api)

		if err := core.PreparePreConfig(api.PreConfig, ctx); err != nil {
			res := &resolvable.ResponseResolvable{
				ResponseCode:        "500",
				ResponseDescription: "Could not prepare pre config",
			}
			return res.SendResponse(c)
		}

		go core.InitExec(api.TriggerFlows, ctx)

		res := <-resChan
		close(resChan)
		// if postableLog, err := log.Post(); err != nil {
		// 	fmt.Println(err)
		// } else {
		// 	core.ConfigStore.AuditLogRepo.InsertLog(postableLog, ctx)
		// 	fmt.Printf("execution time: %v\n", postableLog.TimeTaken)
		// }
		return res.SendResponse(c)
	}
}
