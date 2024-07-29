package common

import (
	"github.com/gofiber/fiber/v2"
)

func ResponseHandler(c *fiber.Ctx, params ...ResponseConfig) error {

	var config ResponseConfig

	if len(params) > 0 {
		config = params[0]
	} else {
		config = ResponseConfig{}
	}

	if config.Response.Code == "" || config.Response.Description == "" {
		config.Response = Responses["Success"]
	}

	if config.Error != nil {
		HandleError(config.Error)
		config.Response = Responses["ServerError"]
		config.Data = config.Error
	}

	if config.Data == nil {
		config.Data = make(JsonObject)
	}

	fRes := make(JsonObject)
	fRes["responseCode"] = config.Response.Code
	fRes["responseDescription"] = config.Response.Description
	fRes["data"] = config.Data

	return c.JSON(fRes)
}

type Response struct {
	Code        string
	Description string
}

type ResponseConfig struct {
	Response Response
	Data     any
	Error    error
}

var Responses = map[string]Response{
	"Success":       {"00", "Success"},
	"ServerError":   {"500", "Internal Server Error"},
	"TypeCastError": {"505", "Error in typecasting property"},
}
