package resolvable

import (
	"context"
	"fmt"
	"ifttt/handler/common"
	"ifttt/handler/domain/audit_log"
	requestvalidator "ifttt/handler/domain/request_validator.go"

	"github.com/fatih/structs"
)

type ResponseResolvable struct {
	ResponseCode        string       `json:"responseCode" mapstructure:"responseCode"`
	ResponseDescription string       `json:"responseDescription" mapstructure:"responseDescription"`
	Response            responseData `json:"response" mapstructure:"response"`
}

type responseData struct {
	RequestToken string      `json:"requestToken" mapstructure:"requestToken"`
	Data         any         `json:"data" mapstructure:"data"`
	Errors       *errorsData `json:"errors" mapstructure:"errors"`
}

type errorsData struct {
	User       []string `json:"user" mapstructure:"user"`
	System     []string `json:"system" mapstructure:"system"`
	Validation []string `json:"validation" mapstructure:"validation"`
}

func (r *ResponseResolvable) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	requestState := common.GetRequestState(ctx)
	reqData := GetRequestData(ctx)
	log := audit_log.GetAuditLogFromContext(ctx)
	if log == nil {
		return nil, fmt.Errorf("log model type assertion failed")
	}

	r.Response.RequestToken = (*log).GetRequestToken()

	if r.Response.Errors == nil {
		r.Response.Errors = &errorsData{}
	}
	execLogs := (*log).GetLogs()
	for _, v := range *execLogs {
		if v.LogType == common.LogError {
			switch v.LogUser {
			case common.LogUser:
				r.Response.Errors.User = append(r.Response.Errors.User, v.LogData)
			case common.LogSystem:
				r.Response.Errors.System = append(r.Response.Errors.System, v.LogData)
			}
		}
	}

	if r.ResponseCode == "" || r.ResponseDescription == "" {
		r.ResponseCode = common.ResponseCodeSuccess
		r.ResponseDescription = common.ResponseDescriptionSuccess
	}
	if len(r.Response.Errors.User) != 0 {
		r.ResponseCode = common.ResponseCodeUserError
		r.ResponseDescription = common.ResponseDescriptionUserError
	}
	if len(r.Response.Errors.System) != 0 {
		r.ResponseCode = common.ResponseCodeSystemError
		r.ResponseDescription = common.ResponseDescriptionSystemError
	}

	r.Response.Data = common.UnSyncMap(reqData.Response)

	(*log).SetResponse(r.ResponseCode, r.ResponseDescription, structs.Map(r.Response))

	resChanUncasted, ok := requestState.Load(common.ContextResponseChannel)
	if !ok {
		return nil, fmt.Errorf("log data not found in map")
	}

	if ok := (*log).SetResponseSent(); ok {
		if responseChannel, ok := resChanUncasted.(chan ResponseResolvable); ok {
			responseChannel <- *r
			return nil, nil
		} else {
			return nil, fmt.Errorf("method Resolve: send res type assertion failed")
		}
	}

	return nil, nil
}

func (r *ResponseResolvable) AddValidationErrors(vErrs []requestvalidator.ValidationError) {
	if r.Response.Errors == nil {
		r.Response.Errors = &errorsData{}
	}
	for _, err := range vErrs {
		if err.Internal {
			r.Response.Errors.System = append(r.Response.Errors.System, err.ErrorInfo.Error())
		} else {
			r.Response.Errors.Validation = append(r.Response.Errors.Validation, err.ErrorInfo.Error())
		}
	}
}
