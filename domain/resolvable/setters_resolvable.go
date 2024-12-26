package resolvable

import (
	"context"
	"fmt"
	"ifttt/handler/common"
)

type setResResolvable map[string]any

type setStoreResolvable map[string]any

type setLogResolvable struct {
	LogData any    `json:"logData" mapstructure:"logData"`
	LogType string `json:"logType" mapstructure:"logType"`
}

func (s *setResResolvable) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	resolvedMap, err := resolveMap((*map[string]any)(s), ctx, dependencies)
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

func (s *setStoreResolvable) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	resolvedMap, err := resolveMap((*map[string]any)(s), ctx, dependencies)
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

func (s *setLogResolvable) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	logDataResolved, err := resolveIfNested(s.LogData, ctx, dependencies)
	if err != nil {
		return nil, err
	}
	common.LogWithTracer(common.LogUser, "user resolvable log", fmt.Sprint(logDataResolved), false, ctx)
	return nil, nil
}
