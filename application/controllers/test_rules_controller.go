package controllers

import (
	"ifttt/handler/application/core"

	"github.com/gofiber/fiber/v2"
)

const TestRulesRoute = "/test/rules"

func NewTestRulesController(router fiber.Router, core *core.ServerCore) {
	controller := testRulesController(core)
	router.Post(TestRulesRoute, controller)
}

func testRulesController(core *core.ServerCore) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		return nil
		// ctx := c.Context()

		// log := audit_log.AuditLog{}
		// ctx.SetUserValue("log", &log)
		// log.StartLog()

		// requestData := request_data.RequestData{}
		// ctx.SetUserValue("request", &requestData)
		// requestData.Initialize()

		// var reqBody map[string]any
		// if err := c.BodyParser(&reqBody); err != nil {
		// 	return core.AddErrorToContext(fmt.Errorf("method testController: could not parse reqBody: %s", err), ctx)
		// }

		// if requestInterface, ok := reqBody["request"]; ok {
		// 	var requestBodyData map[string]any
		// 	if err := mapstructure.Decode(requestInterface, &requestBodyData); err == nil {
		// 		requestData.ReqBody = requestBodyData
		// 	} else {
		// 		return core.AddErrorToContext(fmt.Errorf("method testController: could not decode request data to JsonObject: %s", err), ctx)
		// 	}
		// } else {
		// 	return core.AddErrorToContext(fmt.Errorf("method testController: request data not found in body"), ctx)
		// }

		// var api api.Api
		// if err := mapstructure.Decode(reqBody["api"], &api); err != nil {
		// 	return core.AddErrorToContext(fmt.Errorf("method testController: could not decode api from request body"), ctx)
		// }

		// log.Initialize(&requestData, api.Group, api.Name)

		// resChan := make(chan resolvable.ResponseResolvable)
		// ctx.SetUserValue("resChan", resChan)
		// ctx.SetUserValue("rules", api.Rules)

		// go core.InitExec(api.StartRules, ctx)

		// res := <-resChan
		// if postableLog, err := log.Post(); err != nil {
		// 	fmt.Println(err)
		// } else {
		// 	core.ConfigStore.AuditLogRepo.InsertLog(postableLog, ctx)
		// 	fmt.Printf("execution time: %v", postableLog.TimeTaken)
		// }
		// return c.JSON(res)
	}
}
