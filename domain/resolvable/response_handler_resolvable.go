package resolvable

import (
	"context"
	"fmt"
	"handler/domain/audit_log"
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

func (s *ResponseResolvable) Resolve(ctx context.Context, dependencies map[string]any) (any, error) {
	if log, ok := ctx.Value("log").(*audit_log.AuditLog); !ok {
		return nil, fmt.Errorf("method Resolve: log model type assertion failed")
	} else {
		s.Response.Errors.System = log.GetSystemErrorLogs()
		s.Response.Errors.User = log.GetUserErrorLogs()
	}

	s.Response.Data = getRequestData(ctx).Response

	if responseChannel, ok := ctx.Value("resChan").(chan ResponseResolvable); ok {
		responseChannel <- *s
		return nil, nil
	} else {
		return nil, fmt.Errorf("method Resolve: send res type assertion failed")
	}
}
