package resolvable

import (
	"context"
	"fmt"
	"ifttt/handler/common"
)

type setRes map[string]any

type setStore map[string]any

type setLog struct {
	LogData any    `json:"logData" mapstructure:"logData"`
	LogType string `json:"logType" mapstructure:"logType"`
}

func (s *setRes) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	resolvedMap, err := resolveMapMaybeParallel((*map[string]any)(s), ctx, dependencies)
	if err != nil {
		return nil, err
	}

	reqData := GetRequestData(ctx)
	reqData.Lock()
	defer reqData.Unlock()

	for k, v := range resolvedMap {
		if err := common.SyncMapSet(reqData.Response, k, v); err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (s *setStore) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	resolvedMap, err := resolveMapMaybeParallel((*map[string]any)(s), ctx, dependencies)
	if err != nil {
		return nil, err
	}

	reqData := GetRequestData(ctx)
	reqData.Lock()
	defer reqData.Unlock()

	for k, v := range resolvedMap {
		if err := common.SyncMapSet(reqData.Store, k, v); err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (s *setLog) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	logDataResolved, err := resolveMaybe(s.LogData, ctx, dependencies)
	if err != nil {
		return nil, err
	}
	common.LogWithTracer(common.LogUser, "user resolvable log", fmt.Sprint(logDataResolved), false, ctx)
	return nil, nil
}
