package resolvable

import (
	"context"
	"fmt"
	"handler/common"
	"handler/domain/audit_log"
	"handler/domain/request_data"
)

type SetResResolvable common.JsonObject

type SetStoreResolvable common.JsonObject

type SetLogResolvable struct {
	LogData string `json:"logData" mapstructure:"logData"`
	LogType string `json:"logType" mapstructure:"logType"`
}

func (s *SetResResolvable) Resolve(ctx context.Context) (any, error) {
	if reqData, ok := ctx.Value("request").(*request_data.RequestData); ok {
		responseData := reqData.Response
		for key, value := range *s {
			resVal, err := resolveIfNested(value, ctx)
			if err != nil {
				return nil, err
			}
			responseData[key] = resVal
		}
		return nil, nil
	}
	return nil, fmt.Errorf("method setRes: setRes resolveType assertion failed")
}

func (s *SetStoreResolvable) Resolve(ctx context.Context) (any, error) {
	if reqData, ok := ctx.Value("request").(*request_data.RequestData); ok {
		store := reqData.Store
		for key, value := range *s {
			resVal, err := resolveIfNested(value, ctx)
			if err != nil {
				return nil, err
			}
			store[key] = resVal
		}
		return nil, nil
	}
	return nil, fmt.Errorf("method store: setRes resolveType assertion failed")
}

func (s *SetLogResolvable) Resolve(ctx context.Context) (any, error) {
	if l, ok := ctx.Value("log").(*audit_log.AuditLog); ok {
		logTypeResolved, err := resolveIfNested(s.LogType, ctx)
		if err != nil {
			return nil, err
		}
		logDataResolved, err := resolveIfNested(s.LogData, ctx)
		if err != nil {
			return nil, err
		}
		l.AddExecLog("user", fmt.Sprint(logTypeResolved), fmt.Sprint(logDataResolved))
		return nil, nil
	}
	return nil, fmt.Errorf("method setUserLog: could not type cast log model")
}
