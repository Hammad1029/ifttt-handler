package resolvable

import (
	"context"
	"fmt"
	"ifttt/handler/common"
	"ifttt/handler/domain/audit_log"

	"github.com/fatih/structs"
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

func (r *ResponseResolvable) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	requestState := common.GetRequestState(ctx)

	if r.ResponseCode == "" {
		r.ResponseCode = "00"
	}
	if r.ResponseDescription == "" {
		r.ResponseDescription = "SUCCESS"
	}
	r.Response.Data = common.UnSyncMap(GetRequestData(ctx).Response)

	logUncasted, ok := requestState.Load(common.ContextLog)
	if !ok {
		return nil, fmt.Errorf("log data not found in map")
	}
	if log, ok := logUncasted.(*audit_log.AuditLog); !ok {
		return nil, fmt.Errorf("log model type assertion failed")
	} else {
		r.Response.Errors.System = (*log).GetSystemErrorLogs()
		r.Response.Errors.User = (*log).GetUserErrorLogs()
		(*log).SetFinalResponse(structs.Map(r))
	}

	resChanUncasted, ok := requestState.Load(common.ContextResponseChannel)
	if !ok {
		return nil, fmt.Errorf("log data not found in map")
	}
	if responseChannel, ok := resChanUncasted.(chan ResponseResolvable); ok {
		responseChannel <- *r
		return nil, nil
	} else {
		return nil, fmt.Errorf("method Resolve: send res type assertion failed")
	}
}
