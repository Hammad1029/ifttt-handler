package resolvable

import (
	"context"
	"fmt"
	"ifttt/handler/common"
	"sync"
)

type setResResolvable map[string]any

type setStoreResolvable map[string]any

type setLogResolvable struct {
	LogData any    `json:"logData" mapstructure:"logData"`
	LogType string `json:"logType" mapstructure:"logType"`
}

func (s *setResResolvable) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	responseData := GetRequestData(ctx).Response
	var wg sync.WaitGroup

	cancelCtx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	for key, value := range *s {
		wg.Add(1)
		go func(k string, v any) {
			defer wg.Done()
			select {
			case <-cancelCtx.Done():
				return
			default:
				if resVal, err := resolveIfNested(v, cancelCtx, dependencies); err != nil {
					cancel(err)
				} else {
					responseData.Store(k, resVal)
				}
			}
		}(key, value)
	}

	wg.Wait()
	return nil, context.Cause(cancelCtx)
}

func (s *setStoreResolvable) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	store := GetRequestData(ctx).Store
	var wg sync.WaitGroup

	cancelCtx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	for key, value := range *s {
		wg.Add(1)
		go func(k string, v any) {
			defer wg.Done()
			select {
			case <-cancelCtx.Done():
				return
			default:
				if resVal, err := resolveIfNested(v, cancelCtx, dependencies); err != nil {
					cancel(err)
				} else {
					store.Store(k, resVal)
				}
			}
		}(key, value)
	}

	wg.Wait()
	return nil, context.Cause(cancelCtx)
}

func (s *setLogResolvable) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	logDataResolved, err := resolveIfNested(s.LogData, ctx, dependencies)
	if err != nil {
		return nil, err
	}
	common.LogWithTracer(common.LogUser, "user resolvable log", fmt.Sprint(logDataResolved), false, ctx)
	return nil, nil
}
