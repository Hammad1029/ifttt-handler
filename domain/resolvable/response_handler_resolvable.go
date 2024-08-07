package resolvable

import (
	"context"
	"fmt"
	"handler/domain/audit_log"
)

type ResponseResolvable struct {
	ResponseCode        string       `json:"responseCode" mapstructure:"responseCode"`
	ResponseDescription string       `json:"responseDescription" mapstructure:"responseDescription"`
	Response            ResponseData `json:"response" mapstructure:"response"`
}

type ResponseData struct {
	Data   any `json:"data" mapstructure:"data"`
	Errors any `json:"errors" mapstructure:"errors"`
}

func (s *ResponseResolvable) Resolve(ctx context.Context) (interface{}, error) {
	if log, ok := ctx.Value("log").(*audit_log.AuditLog); !ok {
		return nil, fmt.Errorf("method Resolve: log model type assertion failed")
	} else {
		s.Response.Errors = log.GetUserErrorLogs()
	}

	s.Response.Data = getRequestData(ctx).Response

	if responseChannel, ok := ctx.Value("resChan").(chan ResponseResolvable); ok {
		responseChannel <- *s
		return nil, nil
	} else {
		return nil, fmt.Errorf("method Resolve: send res type assertion failed")
	}
}
