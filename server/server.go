package server

import (
	"fmt"
	"handler/common"
	"handler/config"
	"handler/models"
	"handler/redisUtils"

	"github.com/gofiber/fiber/v2"
	"github.com/mitchellh/mapstructure"
)

func Init() {
	userConfig := models.UserConfigurationResolvable{IsActive: true}
	if err := userConfig.ReadUserConfig(); err != nil {
		common.HandleError(err, "failed to read user configuration")
		return
	}

	port := config.GetConfigProp("app.port")
	app := fiber.New()

	apis, err := redisUtils.GetAllApis()
	if err != nil {
		common.HandleError(err, "failed to get apis from redis")
		return
	}
	app.Post("/testApi", testHandler)
	for _, api := range apis {
		app.Post(fmt.Sprintf("/%s/%s", api.ApiGroup, api.ApiName), apiHandler)
	}

	app.Listen(fmt.Sprintf(":%s", port))
	fmt.Printf("Handler running on port: %s", port)
}

func testHandler(c *fiber.Ctx) error {
	log := models.LogData{}
	log.StartLog()

	ctx := c.Context()

	requestData := models.RequestData{}

	requestData.Initialize()
	var reqBody common.JsonObject
	if err := c.BodyParser(&reqBody); err != nil {
		return addErrorToContext(fmt.Errorf("method testHandler: could not parse reqBody: %s", err), ctx)
	}

	if requestInterface, ok := reqBody["request"]; ok {
		var requestBodyData common.JsonObject
		if err := mapstructure.Decode(requestInterface, &requestBodyData); err == nil {
			requestData.ReqBody = requestBodyData
		} else {
			return addErrorToContext(fmt.Errorf("method testHandler: could not decode request data to JsonObject: %s", err), ctx)
		}
	} else {
		return addErrorToContext(fmt.Errorf("method testHandler: request data not found in body"), ctx)
	}

	var api models.ApiModel
	if err := mapstructure.Decode(reqBody["api"], &api); err != nil {
		return addErrorToContext(fmt.Errorf("method testHandler: could not decode api from request body"), ctx)
	}

	log.Initialize(&requestData, &api)

	resChan := make(chan models.ResponseResolvable)
	ctx.SetUserValue("log", &log)
	ctx.SetUserValue("request", &requestData)
	ctx.SetUserValue("resChan", resChan)
	ctx.SetUserValue("rules", api.Rules)
	ctx.SetUserValue("queries", api.Queries)

	go initExec(api.StartRules, ctx)

	res := <-resChan
	// go log.Post()
	if err := log.Post(); err != nil {
		fmt.Println(err)
	}
	return c.JSON(res)
}

func apiHandler(c *fiber.Ctx) error {
	log := models.LogData{}
	log.StartLog()

	api, err := redisUtils.GetApi(c)
	ctx := c.Context()

	if err != nil {
		return addErrorToContext(err, ctx)
	}

	requestData := models.RequestData{}
	log.Initialize(&requestData, api)

	requestData.Initialize()
	err = c.BodyParser(&requestData.ReqBody)
	if err != nil {
		return addErrorToContext(err, ctx)
	}

	resChan := make(chan models.ResponseResolvable)
	ctx.SetUserValue("log", &log)
	ctx.SetUserValue("request", &requestData)
	ctx.SetUserValue("resChan", resChan)
	ctx.SetUserValue("rules", api.Rules)
	ctx.SetUserValue("queries", api.Queries)

	go initExec(api.StartRules, ctx)

	res := <-resChan
	if err := c.JSON(res); err != nil {
		return err
	}
	if err := log.Post(); err != nil {
		return err
	}

	return nil
}
