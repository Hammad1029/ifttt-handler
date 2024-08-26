package controllers

import (
	"fmt"
	"ifttt/handler/application/core"
	"ifttt/handler/common"
	"ifttt/handler/domain/api"
	"ifttt/handler/domain/audit_log"
	"ifttt/handler/domain/request_data"
	"ifttt/handler/domain/resolvable"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func NewRulesController(router fiber.Router, core *core.ServerCore, api *api.Api) error {
	controller := rulesController(core)
	switch strings.ToUpper(api.Method) {
	case common.RestMethodGet:
		router.Get(api.Path, controller)
	case common.RestMethodPost:
		router.Post(api.Path, controller)
	case common.RestMethodPut:
		router.Put(api.Path, controller)
	case common.RestMethodDelete:
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

		serializedApi, err := core.CacheStore.APICacheRepo.GetApiByPath(c.Path(), ctx)
		if err != nil {
			return core.AddErrorToContext(err, ctx)
		}

		api, err := serializedApi.Unserialize()
		if err != nil {
			return core.AddErrorToContext(err, ctx)
		}

		requestData := request_data.RequestData{}
		requestData.Initialize()
		log.Initialize(&requestData, api.Group, api.Name)

		err = c.BodyParser(&requestData.ReqBody)
		if err != nil {
			return core.AddErrorToContext(err, ctx)
		}

		resChan := make(chan resolvable.ResponseResolvable)
		ctx.SetUserValue("request", &requestData)
		ctx.SetUserValue("resChan", resChan)
		ctx.SetUserValue("rules", api.Rules)

		go core.InitExec(api.StartRules, ctx)

		res := <-resChan
		if postableLog, err := log.Post(); err != nil {
			fmt.Println(err)
		} else {
			core.ConfigStore.AuditLogRepo.InsertLog(postableLog, ctx)
			fmt.Printf("execution time: %v\n", postableLog.TimeTaken)
		}
		return c.JSON(res)
	}
}
