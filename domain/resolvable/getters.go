package resolvable

import (
	"context"
	"ifttt/handler/common"
	"ifttt/handler/domain/request_data"
)

type getErrors struct{}

type getStore struct{}

type getHeaders struct{}

type getConst struct {
	Value any `json:"value" mapstructure:"value"`
}

func (r *getErrors) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	err := request_data.GetRequestData(ctx).Errors
	errStr := make([]string, 0, len(err))
	for _, e := range err {
		errStr = append(errStr, e.Error())
	}
	return errStr, nil
}

func (s *getStore) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	return request_data.GetRequestData(ctx).Store, nil
}

func (c *getConst) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	return c.Value, nil
}

func (h *getHeaders) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	return request_data.GetRequestData(ctx).Headers, nil
}
