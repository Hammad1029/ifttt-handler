package request_data

import (
	"context"
	"fmt"
	"ifttt/handler/common"
	"sync"
)

type RequestData struct {
	Mtx                sync.Mutex
	Errors             []error           `json:"errors" mapstructure:"errors"`
	ReqBody            map[string]any    `json:"reqBody" mapstructure:"reqBody"`
	Headers            map[string]string `json:"headers" mapstructure:"headers"`
	AggregatedResponse map[string]any    `json:"aggregatedResponse" mapstructure:"aggregatedResponse"`
	PreConfig          map[string]any    `json:"preConfig" mapstructure:"preConfig"`
	Store              map[string]any    `json:"store" mapstructure:"store"`
	Response           map[string]any    `json:"response" mapstructure:"response"`
	ExternalTrips      []ExternalTrip    `json:"externalTrips" mapstructure:"externalTrips"`
}

type ExternalTrip struct {
	Key        string          `json:"key" mapstructure:"key"`
	Identifier string          `json:"identifier" mapstructure:"identifier"`
	TimeTaken  uint64          `json:"timeTaken" mapstructure:"timeTaken"`
	Data       *map[string]any `json:"data" mapstructure:"data"`
}

func NewRequestData() *RequestData {
	r := RequestData{}
	r.Mtx = sync.Mutex{}
	r.Errors = []error{}
	r.ReqBody = make(map[string]any)
	r.Headers = make(map[string]string)
	r.AggregatedResponse = make(map[string]any)
	r.PreConfig = make(map[string]any)
	r.Store = make(map[string]any)
	r.Response = make(map[string]any)
	r.ExternalTrips = []ExternalTrip{}
	return &r
}

func (e *RequestData) AddErrors(errArgs ...error) {
	if len(errArgs) != 0 {
		e.Errors = append(e.Errors, errArgs...)
	}
}

func AddExternalTrip(
	key string, identifier string, data *map[string]any, timeTaken uint64, ctx context.Context,
) {
	errorMsg := "error in adding external trip"
	reqData, ok := common.GetCtxState(ctx).Load(common.ContextRequestData)
	if !ok {
		common.LogWithTracer(common.LogSystem, errorMsg, fmt.Errorf("could not get request data"), true, ctx)
	}
	r, ok := reqData.(*RequestData)
	if !ok {
		common.LogWithTracer(common.LogSystem, errorMsg, fmt.Errorf("could not cast request data"), true, ctx)
	}

	r.Mtx.Lock()
	defer r.Mtx.Unlock()

	r.ExternalTrips = append(r.ExternalTrips,
		ExternalTrip{Key: key, Identifier: identifier, Data: data, TimeTaken: timeTaken})
	if ctxState := common.GetCtxState(ctx); ctxState == nil {
		common.LogWithTracer(common.LogSystem, errorMsg, fmt.Errorf("could not load external trips"), true, ctx)
	} else if externalExecTime, ok := ctxState.Load(common.ContextExternalExecTime); ok {
		ctxState.Store(common.ContextExternalExecTime, externalExecTime.(uint64)+timeTaken)
	}
}
