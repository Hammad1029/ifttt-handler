package request_data

import (
	"context"
	"fmt"
	"ifttt/handler/common"
	"sync"
)

type RequestData struct {
	sync.Mutex
	ReqBody            map[string]any      `json:"reqBody" mapstructure:"reqBody"`
	Headers            map[string][]string `json:"headers" mapstructure:"headers"`
	AggregatedResponse map[string]any      `json:"aggregatedResponse" mapstructure:"aggregatedResponse"`
	PreConfig          *sync.Map           `json:"preConfig" mapstructure:"preConfig"`
	Store              *sync.Map           `json:"store" mapstructure:"store"`
	Response           *sync.Map           `json:"response" mapstructure:"response"`
	ExternalTrips      *sync.Map           `json:"externalTrips" mapstructure:"externalTrips"`
}

func (r *RequestData) Initialize() {
	r.ReqBody = make(map[string]any)
	r.Headers = make(map[string][]string)
	r.AggregatedResponse = make(map[string]any)
	r.PreConfig = &sync.Map{}
	r.Store = &sync.Map{}
	r.Response = &sync.Map{}
	r.ExternalTrips = &sync.Map{}
	r.ExternalTrips.Store(common.ExternalTripDump, &[]map[string]any{})
	r.ExternalTrips.Store(common.ExternalTripApi, &[]map[string]any{})
	r.ExternalTrips.Store(common.ExternalTripQuery, &[]map[string]any{})
}

func (r *RequestData) UnSync() *map[string]any {
	return &map[string]any{
		"request":             r.ReqBody,
		"headers":             r.Headers,
		"aggregated_response": r.AggregatedResponse,
		"preConfig":           common.UnSyncMap(r.PreConfig),
		"store":               common.UnSyncMap(r.Store),
		"response":            common.UnSyncMap(r.Response),
		"externalTrips":       common.UnSyncMap(r.ExternalTrips),
	}
}

func AddExternalTrip(
	key string, data map[string]any, timeTaken uint64, ctx context.Context,
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

	r.Lock()
	defer r.Unlock()
	if result, ok := r.ExternalTrips.Load(key); ok {
		trips := result.(*[]map[string]any)
		*trips = append(*trips, data)
		r.ExternalTrips.Store(key, trips)

		ctxState := common.GetCtxState(ctx)
		if ctxState != nil {
			if externalExecTime, ok := ctxState.Load(common.ContextExternalExecTime); ok {
				ctxState.Store(common.ContextExternalExecTime, externalExecTime.(uint64)+timeTaken)
			}
		}
	} else {
		common.LogWithTracer(common.LogSystem, errorMsg, fmt.Errorf("could not load external trips"), true, ctx)
	}
}
