package utils

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
		config.Data = make(map[string]interface{})
	}

	fRes := make(map[string]interface{})
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
	Data     interface{}
	Error    error
}

var Responses = map[string]Response{
	"Success":          {"00", "Success"},
	"ClientNotFound":   {"05", "Client Not Found"},
	"ApiNotFound":      {"10", "Api Not Found"},
	"ApiAlreadyExists": {"15", "API Already Exists"},
	"WrongTableFormat": {"20", "Wrong Table Format"},
	"TableNotFound":    {"25", "Table Not Found"},
	"IndexNotPossible": {"25", "Index Not Possible"},
	"IndexNotFound":    {"30", "Index Not Found"},

	"BadRequest":   {"400", "Bad request"},
	"Unauthorized": {"401", "Unauthorized"},
	"NotFound":     {"404", "Not Found"},
	"ServerError":  {"500", "Internal Server Error"},
}
