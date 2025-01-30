package request_data

import (
	"context"
	"fmt"
	"ifttt/handler/common"
	"sync"
)

func NewRequestData() *RequestData {
	r := RequestData{}
	r.Mtx = sync.Mutex{}
	r.Errors = []error{}
	r.Headers = make(map[string]string)
	r.AggregatedResponse = make(map[string]any)
	r.Store = make(map[string]any)
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

func GetRequestData(ctx context.Context) *RequestData {
	if reqData, ok := common.GetCtxState(ctx).Load(common.ContextRequestData); ok {
		return reqData.(*RequestData)
	}
	return nil
}

func (r *RequestData) SetStore(key string, value any) error {
	r.Mtx.Lock()
	defer r.Mtx.Unlock()
	if err := common.MapJQSet(&r.Store, key, value); err != nil {
		return err
	}
	return nil
}
