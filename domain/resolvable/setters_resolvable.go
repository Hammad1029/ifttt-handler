package resolvable

import (
	"context"
	"fmt"
	"ifttt/handler/domain/audit_log"
	"sync"
)

type SetResResolvable map[string]any

type SetStoreResolvable map[string]any

type SetLogResolvable struct {
	LogData string `json:"logData" mapstructure:"logData"`
	LogType string `json:"logType" mapstructure:"logType"`
}

func (s *SetResResolvable) Resolve(ctx context.Context, dependencies map[string]any) (any, error) {
	responseData := GetRequestData(ctx).Response
	var (
		once sync.Once
		wg   sync.WaitGroup
	)

	cancelCtx, cancel := context.WithCancelCause(ctx)

	for key, value := range *s {
		wg.Add(1)
		go func(k string, v any) {
			defer wg.Done()
			select {
			case <-cancelCtx.Done():
				return
			default:
				if resVal, err := resolveIfNested(v, cancelCtx, dependencies); err != nil {
					once.Do(func() {
						cancel(err)
					})
				} else {
					responseData[k] = resVal
				}
			}
		}(key, value)
	}

	<-cancelCtx.Done()
	wg.Wait()
	return nil, context.Cause(cancelCtx)
}

func (s *SetStoreResolvable) Resolve(ctx context.Context, dependencies map[string]any) (any, error) {
	store := GetRequestData(ctx).Store
	var (
		once sync.Once
		wg   sync.WaitGroup
	)

	cancelCtx, cancel := context.WithCancelCause(ctx)

	for key, value := range *s {
		wg.Add(1)
		go func(k string, v any) {
			defer wg.Done()
			select {
			case <-cancelCtx.Done():
				return
			default:
				if resVal, err := resolveIfNested(v, cancelCtx, dependencies); err != nil {
					once.Do(func() {
						cancel(err)
					})
				} else {
					store[k] = resVal
				}
			}
		}(key, value)
	}

	<-cancelCtx.Done()
	wg.Wait()
	return nil, context.Cause(cancelCtx)
}

func (s *SetLogResolvable) Resolve(ctx context.Context, dependencies map[string]any) (any, error) {
	if l, ok := ctx.Value("log").(*audit_log.AuditLog); ok {
		logTypeResolved, err := resolveIfNested(s.LogType, ctx, dependencies)
		if err != nil {
			return nil, err
		}
		logDataResolved, err := resolveIfNested(s.LogData, ctx, dependencies)
		if err != nil {
			return nil, err
		}
		l.AddExecLog("user", fmt.Sprint(logTypeResolved), fmt.Sprint(logDataResolved))
		return nil, nil
	}
	return nil, fmt.Errorf("method setUserLog: could not type cast log model")
}
