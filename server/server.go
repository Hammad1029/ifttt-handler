package server

import (
	"fmt"
	"handler/config"
	"handler/models"
	"handler/redisUtils"
	"handler/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/mitchellh/mapstructure"
)

func Init() {
	port := config.GetConfigProp("app.port")
	app := fiber.New()

	apis, err := redisUtils.GetAllApis()
	if err != nil {
		utils.HandleError(err, "failed to get apis from redis")
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
	reqBody := make(map[string]interface{})
	if err := c.BodyParser(&reqBody); err != nil {
		return utils.ResponseHandler(c, utils.ResponseConfig{Error: fmt.Errorf("could parse reqBody: %s", err)})
	}

	if reqBodyMap, ok := reqBody["request"].(map[string]interface{}); !ok {
		return utils.ResponseHandler(c, utils.ResponseConfig{Error: fmt.Errorf("could not cast reqbody to map")})
	} else {
		requestData.ReqBody = reqBodyMap
	}

	var api models.ApiModel
	if err := mapstructure.Decode(reqBody["api"], &api); err != nil {
		return utils.ResponseHandler(c, utils.ResponseConfig{Error: fmt.Errorf("could not decode api from request body")})
	}

	log.Initialize(&requestData, &api)

	resChan := make(chan utils.Response)
	ctx.SetUserValue("log", &log)
	ctx.SetUserValue("request", &requestData)
	ctx.SetUserValue("resChan", resChan)
	ctx.SetUserValue("rules", api.Rules)
	ctx.SetUserValue("queries", api.Queries)

	go initExec(api.StartRules, ctx)

	res := <-resChan
	err := utils.ResponseHandler(c, utils.ResponseConfig{
		Response: res,
		Data: map[string]interface{}{
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
		return utils.ResponseHandler(c, utils.ResponseConfig{Error: err})
	case api == nil:
		return utils.ResponseHandler(c, utils.ResponseConfig{Response: utils.Responses["ApiNotFound"]})
	}

	requestData := models.RequestData{}
	log.Initialize(&requestData, api)

	requestData.Initialize()
	err = c.BodyParser(&requestData.ReqBody)
	if err != nil {
		return utils.ResponseHandler(c, utils.ResponseConfig{Error: err})
	}

	resChan := make(chan utils.Response)
	ctx.SetUserValue("log", &log)
	ctx.SetUserValue("request", &requestData)
	ctx.SetUserValue("resChan", resChan)
	ctx.SetUserValue("rules", api.Rules)
	ctx.SetUserValue("queries", api.Queries)

	go initExec(api.StartRules, ctx)

	res := <-resChan
	utils.ResponseHandler(c, utils.ResponseConfig{
		Response: res,
		Data: map[string]interface{}{
			"response": requestData.Response,
			"errors":   log.GetUserErrorLogs(),
		},
	})
	log.Post()

	return nil
}
