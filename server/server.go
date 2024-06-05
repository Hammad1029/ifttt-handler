package server

import (
	"fmt"
	"handler/config"
	"handler/models"
	"handler/utils"

	"github.com/gofiber/fiber/v2"
)

func Init() {
	port := config.GetConfigProp("app.port")
	app := fiber.New()
	app.Post("/:apiName", apiHandler)
	app.Listen(fmt.Sprintf(":%s", port))
	fmt.Printf("Handler running on port: %s", port)
}

func apiHandler(c *fiber.Ctx) error {

	api, err := getApiFromRedis(c)
	ctx := c.Context()

	switch {
	case err != nil:
		return utils.ResponseHandler(c, utils.ResponseConfig{Error: err})
	case api == nil:
		return utils.ResponseHandler(c, utils.ResponseConfig{Response: utils.Responses["ApiNotFound"]})
	}

	requestData := models.RequestData{}
	requestData.Initialize(api)
	err = c.BodyParser(&requestData.ReqBody)
	if err != nil {
		return utils.ResponseHandler(c, utils.ResponseConfig{Error: err})
	}

	log := models.LogModel{}
	log.Initialize(&requestData, api)

	resChan := make(chan string)
	ctx.SetUserValue("log", log)
	ctx.SetUserValue("request", requestData)
	ctx.SetUserValue("resChan", resChan)
	ctx.SetUserValue("rules", api.Rules)
	ctx.SetUserValue("queries", api.Queries)

	go initExec(api.StartRules, ctx)

	responseCode := <-resChan
	utils.ResponseHandler(c, utils.ResponseConfig{
		Response: utils.Response{Code: responseCode},
		Data: map[string]interface{}{
			"response": requestData.Response,
		},
	})
	log.Post()

	return nil
}
