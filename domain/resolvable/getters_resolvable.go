package resolvable

import (
	"context"
	"ifttt/handler/common"
	"ifttt/handler/domain/request_data"
)

type getRequestResolvable struct{}

type getResponseResolvable struct{}

type getStoreResolvable struct{}

type getPreConfigResolvable struct{}

type getHeadersResolvable struct{}

type getConstResolvable struct {
	Value any `json:"value" mapstructure:"value"`
}

func GetRequestData(ctx context.Context) *request_data.RequestData {
	if reqData, ok := common.GetRequestState(ctx).Load(common.ContextRequestData); ok {
		return reqData.(*request_data.RequestData)
	}
	return nil
}

func (r *getRequestResolvable) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	return GetRequestData(ctx).ReqBody, nil
}

func (r *getResponseResolvable) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	return common.UnSyncMap(GetRequestData(ctx).Response), nil
}

func (s *getStoreResolvable) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	return GetRequestData(ctx).Store, nil
}

func (c *getConstResolvable) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	return c.Value, nil
}

func (c *getPreConfigResolvable) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	return GetRequestData(ctx).PreConfig, nil
}

func (h *getHeadersResolvable) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	return GetRequestData(ctx).Headers, nil
}
