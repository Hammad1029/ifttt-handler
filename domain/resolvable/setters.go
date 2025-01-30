package resolvable

import (
	"context"
	"fmt"
	"ifttt/handler/common"
	"ifttt/handler/domain/request_data"
)

type setRes map[string]any

type setStore map[string]any

type setLog struct {
	LogData any    `json:"logData" mapstructure:"logData"`
	LogType string `json:"logType" mapstructure:"logType"`
}

func (s *setStore) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	resolvedMap, err := resolveMapMaybeParallel((*map[string]any)(s), ctx, dependencies)
	if err != nil {
		return nil, err
	}

	reqData := request_data.GetRequestData(ctx)
	for k, v := range resolvedMap {
		if err := reqData.SetStore(k, v); err != nil {
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
