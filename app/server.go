package app

import (
	"context"
	"fmt"
	"handler/common"
	"handler/config"
	"handler/domain/api"
	"handler/domain/audit_log"
	"handler/domain/request_data"
	"handler/domain/resolvable"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/mitchellh/mapstructure"
)

var serverCore *core

func Init() {

	if core, err := newCore(); err != nil {
		panic(fmt.Errorf("could not create core: %s", err))
	} else {
		serverCore = core
	}

	port := config.GetConfigProp("app.port")
	app := fiber.New()
	ctx := context.Background()

	apis, err := serverCore.ConfigStore.APIPersistentRepo.GetAllApis(ctx)
	if err != nil {
		panic(fmt.Errorf("could not get apis from persistent storage: %s", err))
	} else {
		serverCore.CacheStore.APICacheRepo.StoreApis(apis, ctx)
	}

	app.Post("/testApi", testHandler)
	for _, api := range apis {
		app.Post(fmt.Sprintf("/%s/%s", api.ApiGroup, api.ApiName), apiHandler)
	}

	app.Listen(fmt.Sprintf(":%s", port))
	fmt.Printf("Handler running on port: %s", port)
}

func testHandler(c *fiber.Ctx) error {
	log := audit_log.AuditLog{}
	log.StartLog()

	ctx := c.Context()

	requestData := request_data.RequestData{}

	requestData.Initialize()
	var reqBody common.JsonObject
	if err := c.BodyParser(&reqBody); err != nil {
		return serverCore.addErrorToContext(fmt.Errorf("method testHandler: could not parse reqBody: %s", err), ctx)
	}

	if requestInterface, ok := reqBody["request"]; ok {
		var requestBodyData common.JsonObject
		if err := mapstructure.Decode(requestInterface, &requestBodyData); err == nil {
			requestData.ReqBody = requestBodyData
		} else {
			return serverCore.addErrorToContext(fmt.Errorf("method testHandler: could not decode request data to JsonObject: %s", err), ctx)
		}
	} else {
		return serverCore.addErrorToContext(fmt.Errorf("method testHandler: request data not found in body"), ctx)
	}

	var api api.Api
	if err := mapstructure.Decode(reqBody["api"], &api); err != nil {
		return serverCore.addErrorToContext(fmt.Errorf("method testHandler: could not decode api from request body"), ctx)
	}

	log.Initialize(&requestData, api.ApiGroup, api.ApiName)

	resChan := make(chan resolvable.ResponseResolvable)
	ctx.SetUserValue("log", &log)
	ctx.SetUserValue("request", &requestData)
	ctx.SetUserValue("resChan", resChan)
	ctx.SetUserValue("rules", api.Rules)
	ctx.SetUserValue("queries", api.Queries)

	go serverCore.initExec(api.StartRules, ctx)

	res := <-resChan
	if postableLog, err := log.Post(); err != nil {
		fmt.Println(err)
	} else {
		serverCore.ConfigStore.AuditLogRepo.InsertLog(postableLog, ctx)
		fmt.Printf("execution time: %s", postableLog.TimeTaken)
	}
	return c.JSON(res)
}

func apiHandler(c *fiber.Ctx) error {
	ctx := c.Context()

	log := audit_log.AuditLog{}
	log.StartLog()

	apiParts := strings.Split(c.Path(), "/")
	api, err := serverCore.CacheStore.APICacheRepo.GetApiByGroupAndName(apiParts[0], apiParts[1], ctx)
	if err != nil {
		return serverCore.addErrorToContext(err, ctx)
	}

	requestData := request_data.RequestData{}
	log.Initialize(&requestData, api.ApiGroup, api.ApiName)

	requestData.Initialize()
	err = c.BodyParser(&requestData.ReqBody)
	if err != nil {
		return serverCore.addErrorToContext(err, ctx)
	}

	resChan := make(chan resolvable.ResponseResolvable)
	ctx.SetUserValue("log", &log)
	ctx.SetUserValue("request", &requestData)
	ctx.SetUserValue("resChan", resChan)
	ctx.SetUserValue("rules", api.Rules)
	ctx.SetUserValue("queries", api.Queries)

	go serverCore.initExec(api.StartRules, ctx)

	res := <-resChan
	if postableLog, err := log.Post(); err != nil {
		fmt.Println(err)
	} else {
		serverCore.ConfigStore.AuditLogRepo.InsertLog(postableLog, ctx)
		fmt.Printf("execution time: %s", postableLog.TimeTaken)
	}
	return c.JSON(res)
}
