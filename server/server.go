package server

import (
	"fmt"
	"handler/config"
	"handler/models"
	redisUtils "handler/rediisUtils"
	"handler/utils"

	"github.com/gofiber/fiber/v2"
)

func Init() {
	port := config.GetConfigProp("app.port")
	app := fiber.New()

	apis, err := redisUtils.GetAllApis()
	if err != nil {
		utils.HandleError(err, "failed to get apis from redis")
		return
	}
	for _, api := range apis {
		app.Post(fmt.Sprintf("/%s/%s", api.ApiGroup, api.ApiName), apiHandler)
	}

	app.Listen(fmt.Sprintf(":%s", port))
	fmt.Printf("Handler running on port: %s", port)
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

	requestData.Initialize(api)
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
