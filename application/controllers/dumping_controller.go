package controllers

import (
	"fmt"
	"handler/application/core"
	"handler/domain/api"
	"handler/domain/audit_log"
	"handler/domain/request_data"
	"handler/domain/resolvable"

	"github.com/gofiber/fiber/v2"
)

func NewDumpingController(router fiber.Router, core *core.ServerCore, api *api.Api) {
	controller := dumpingController(core)
	router.Post(api.Path, controller)
}

func dumpingController(core *core.ServerCore) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()

		log := audit_log.AuditLog{}
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
		log.Initialize(&requestData, api.Group, api.Name)

		requestData.Initialize()
		err = c.BodyParser(&requestData.ReqBody)
		if err != nil {
			return core.AddErrorToContext(err, ctx)
		}

		resChan := make(chan resolvable.ResponseResolvable)
		ctx.SetUserValue("log", &log)
		ctx.SetUserValue("request", &requestData)
		ctx.SetUserValue("resChan", resChan)

		go func() error {
			queryString, parameters, err := api.Dumping.CreateInsertQuery(ctx, core.ResolvableDependencies)
			if err != nil {
				core.AddErrorToContext(err, ctx)
			}
			if err := core.DataStore.RawQueryRepo.RawExecPositional(queryString, parameters); err != nil {
				core.AddErrorToContext(err, ctx)
			}
			response := &resolvable.ResponseResolvable{}
			response.Resolve(ctx, core.ResolvableDependencies)
			return nil
		}()

		res := <-resChan
		if postableLog, err := log.Post(); err != nil {
			fmt.Println(err)
		} else {
			core.ConfigStore.AuditLogRepo.InsertLog(postableLog, ctx)
			fmt.Printf("execution time: %v", postableLog.TimeTaken)
		}
		return c.JSON(res)
	}
}
