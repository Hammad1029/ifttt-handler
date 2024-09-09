package resolvable

import (
	"context"
	"fmt"
	"ifttt/handler/domain/audit_log"

	"github.com/gofiber/fiber/v2"
)

type ResponseResolvable struct {
	ResponseCode        string       `json:"responseCode" mapstructure:"responseCode"`
	ResponseDescription string       `json:"responseDescription" mapstructure:"responseDescription"`
	Response            responseData `json:"response" mapstructure:"response"`
}

type responseData struct {
	Data   any        `json:"data" mapstructure:"data"`
	Errors errorsData `json:"errors" mapstructure:"errors"`
}

type errorsData struct {
	User   any `json:"user" mapstructure:"user"`
	System any `json:"system" mapstructure:"system"`
}

func (r *ResponseResolvable) Resolve(ctx context.Context, dependencies map[string]any) (any, error) {
	if log, ok := ctx.Value("log").(*audit_log.AuditLog); !ok {
		return nil, fmt.Errorf("method Resolve: log model type assertion failed")
	} else {
		r.Response.Errors.System = log.GetSystemErrorLogs()
		r.Response.Errors.User = log.GetUserErrorLogs()
	}

	if r.ResponseCode == "" {
		r.ResponseCode = "00"
	}
	if r.ResponseDescription == "" {
		r.ResponseDescription = "SUCCESS"
	}
	r.Response.Data = GetRequestData(ctx).Response

	if responseChannel, ok := ctx.Value("resChan").(chan ResponseResolvable); ok {
		responseChannel <- *r
		return nil, nil
	} else {
		return nil, fmt.Errorf("method Resolve: send res type assertion failed")
	}
}

func (r *ResponseResolvable) SendResponse(ctx *fiber.Ctx) error {
	return ctx.JSON(r)
}
