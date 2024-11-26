package resolvable

import (
	"context"
	"fmt"
	"ifttt/handler/common"
	requestvalidator "ifttt/handler/domain/request_validator.go"
)

type ResponseResolvable struct {
	ResponseCode        string       `json:"responseCode" mapstructure:"responseCode"`
	ResponseDescription string       `json:"responseDescription" mapstructure:"responseDescription"`
	Response            responseData `json:"response" mapstructure:"response"`
}

type responseData struct {
	Tracer string      `json:"tracer" mapstructure:"tracer"`
	Data   any         `json:"data" mapstructure:"data"`
	Errors *errorsData `json:"errors" mapstructure:"errors"`
}

type errorsData struct {
	System     []string `json:"system" mapstructure:"system"`
	Validation []string `json:"validation" mapstructure:"validation"`
}

func (r *ResponseResolvable) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	requestState := common.GetCtxState(ctx)
	reqData := GetRequestData(ctx)

	if tracer, ok := requestState.Load(common.ContextTracer); ok {
		r.Response.Tracer = tracer.(string)
	}

	if r.Response.Errors == nil {
		r.Response.Errors = &errorsData{}
	}

	if r.ResponseCode == "" || r.ResponseDescription == "" {
		r.ResponseCode = common.ResponseCodeSuccess
		r.ResponseDescription = common.ResponseDescriptionSuccess
	}

	r.Response.Data = common.UnSyncMap(reqData.Response)

	resChanUncasted, ok := requestState.Load(common.ContextResponseChannel)
	if !ok {
		return nil, fmt.Errorf("response channel not found")
	}

	if responseChannel, ok := resChanUncasted.(chan ResponseResolvable); ok {
		if ok := common.SetResponseSent(ctx); ok {
			common.LogWithTracer(common.LogSystem,
				fmt.Sprintf("Sending response | Response Code: %s | Response Description: %s",
					r.ResponseCode, r.ResponseDescription), r, false, ctx)
			responseChannel <- *r
			close(responseChannel)
			return nil, nil
		}
	} else {
		return nil, fmt.Errorf("method Resolve: response channel type assertion failed")
	}

	return nil, nil
}

func (r *ResponseResolvable) ManualSend(resChan chan ResponseResolvable, dependencies map[common.IntIota]any, ctx context.Context) {
	if !common.GetResponseSent(ctx) {
		if _, err := r.Resolve(ctx, dependencies); err != nil {
			r.AddError(err)
			common.LogWithTracer(common.LogSystem, "error in resolving response", err, true, ctx)
			r.ResponseCode = common.ResponseCodeSystemError
			r.ResponseDescription = common.ResponseDescriptionSystemError
			if ok := common.SetResponseSent(ctx); ok {
				common.LogWithTracer(common.LogSystem,
					fmt.Sprintf("Sending response | Response Code: %s | Response Description: %s",
						r.ResponseCode, r.ResponseDescription), r, false, ctx)
				resChan <- *r
				close(resChan)
			}
		}
	}
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

func (r *ResponseResolvable) AddError(err error) {
	if r.Response.Errors == nil {
		r.Response.Errors = &errorsData{}
	}
	r.Response.Errors.System = append(r.Response.Errors.System, err.Error())
}
