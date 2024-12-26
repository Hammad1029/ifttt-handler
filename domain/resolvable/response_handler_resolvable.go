package resolvable

import (
	"context"
	"fmt"
	"ifttt/handler/common"
	requestvalidator "ifttt/handler/domain/request_validator.go"

	"github.com/fatih/structs"
)

type ResponseResolvable struct {
	ResponseCode        string       `json:"responseCode" mapstructure:"responseCode"`
	ResponseDescription string       `json:"responseDescription" mapstructure:"responseDescription"`
	Data                responseData `json:"data" mapstructure:"data"`
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
		r.Data.Tracer = tracer.(string)
	}

	if r.Data.Errors == nil {
		r.Data.Errors = &errorsData{}
	}

	if r.ResponseCode == "" || r.ResponseDescription == "" {
		r.ResponseCode = common.ResponseCodeSuccess
		r.ResponseDescription = common.ResponseDescriptionSuccess
	}

	if len(r.Data.Errors.System) == 0 && len(r.Data.Errors.Validation) == 0 {
		r.Data.Data = common.SyncMapUnsync(reqData.Response)
	}

	resChanUncasted, ok := requestState.Load(common.ContextResponseChannel)
	if !ok {
		return nil, fmt.Errorf("response channel not found")
	}

	if responseChannel, ok := resChanUncasted.(chan ResponseResolvable); ok {
		r.channelSend(responseChannel, ctx)
	} else {
		return nil, fmt.Errorf("method Resolve: response channel type assertion failed")
	}

	return nil, nil
}

func (r *ResponseResolvable) ManualSend(resChan chan ResponseResolvable, pErr error, ctx context.Context) {
	if !common.GetResponseSent(ctx) {
		if pErr != nil {
			r.addError(pErr)
		}
		if _, err := r.Resolve(ctx, nil); err != nil {
			r.addError(err)
			common.LogWithTracer(common.LogSystem, "error in resolving response", err, true, ctx)
			r.ResponseCode = common.ResponseCodeSystemError
			r.ResponseDescription = common.ResponseDescriptionSystemError
			r.channelSend(resChan, ctx)
		}
	}
}

func (r *ResponseResolvable) channelSend(resChan chan ResponseResolvable, ctx context.Context) {
	if ok := common.SetResponseSent(ctx); ok {
		if reqData := GetRequestData(ctx); reqData != nil {
			reqData.AggregatedResponse = structs.Map(r)
		}
		common.LogWithTracer(common.LogSystem,
			fmt.Sprintf("Sending response | Response Code: %s | Response Description: %s",
				r.ResponseCode, r.ResponseDescription), r, false, ctx)
		resChan <- *r
		close(resChan)
	}
}

func (r *ResponseResolvable) AddValidationErrors(vErrs []requestvalidator.ValidationError) {
	if r.Data.Errors == nil {
		r.Data.Errors = &errorsData{}
	}
	for _, err := range vErrs {
		if err.Internal {
			r.Data.Errors.System = append(r.Data.Errors.System, err.ErrorInfo.Error())
		} else {
			r.Data.Errors.Validation = append(r.Data.Errors.Validation, err.ErrorInfo.Error())
		}
	}
}

func (r *ResponseResolvable) addError(err error) {
	if r.Data.Errors == nil {
		r.Data.Errors = &errorsData{}
	}
	r.Data.Errors.System = append(r.Data.Errors.System, err.Error())
}
