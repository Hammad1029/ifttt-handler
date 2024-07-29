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
		return common.ResponseHandler(c, common.ResponseConfig{Error: fmt.Errorf("method testHandler: could not parse reqBody: %s", err)})
	}

	if requestInterface, ok := reqBody["request"]; ok {
		var requestBodyData common.JsonObject
		if err := mapstructure.Decode(requestInterface, &requestBodyData); err == nil {
			requestData.ReqBody = requestBodyData
		} else {
			return common.ResponseHandler(c, common.ResponseConfig{Error: fmt.Errorf("method testHandler: could not decode request data to JsonObject: %s", err)})
		}
	} else {
		return common.ResponseHandler(c, common.ResponseConfig{Error: fmt.Errorf("method testHandler: request data not found in body")})
	}

	var api models.ApiModel
	if err := mapstructure.Decode(reqBody["api"], &api); err != nil {
		return common.ResponseHandler(c, common.ResponseConfig{Error: fmt.Errorf("could not decode api from request body")})
	}

	log.Initialize(&requestData, &api)

	resChan := make(chan common.Response)
	ctx.SetUserValue("log", &log)
	ctx.SetUserValue("request", &requestData)
	ctx.SetUserValue("resChan", resChan)
	ctx.SetUserValue("rules", api.Rules)
	ctx.SetUserValue("queries", api.Queries)

	go initExec(api.StartRules, ctx)

	res := <-resChan
	err := common.ResponseHandler(c, common.ResponseConfig{
		Response: res,
		Data: common.JsonObject{
			"response": requestData.Response,
			"errors":   log.GetUserErrorLogs(),
		},
	})
	log.Post()

	return err
}

func apiHandler(c *fiber.Ctx) error {
	log := models.LogData{}
	log.StartLog()

	api, err := redisUtils.GetApi(c)
	ctx := c.Context()

	switch {
	case err != nil:
		return common.ResponseHandler(c, common.ResponseConfig{Error: err})
	case api == nil:
		return common.ResponseHandler(c, common.ResponseConfig{Response: common.Responses["ApiNotFound"]})
	}

	requestData := models.RequestData{}
	log.Initialize(&requestData, api)

	requestData.Initialize()
	err = c.BodyParser(&requestData.ReqBody)
	if err != nil {
		return common.ResponseHandler(c, common.ResponseConfig{Error: err})
	}

	resChan := make(chan common.Response)
	ctx.SetUserValue("log", &log)
	ctx.SetUserValue("request", &requestData)
	ctx.SetUserValue("resChan", resChan)
	ctx.SetUserValue("rules", api.Rules)
	ctx.SetUserValue("queries", api.Queries)

	go initExec(api.StartRules, ctx)

	res := <-resChan
	common.ResponseHandler(c, common.ResponseConfig{
		Response: res,
		Data: common.JsonObject{
			"response": requestData.Response,
			"errors":   log.GetUserErrorLogs(),
		},
	})
	log.Post()

	return nil
}
