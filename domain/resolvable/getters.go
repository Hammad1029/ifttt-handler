package resolvable

import (
	"context"
	"ifttt/handler/common"
	"ifttt/handler/domain/request_data"
)

type getErrors struct{}

type getRequest struct{}

type getResponse struct{}

type getStore struct{}

type getPreConfig struct{}

type getHeaders struct{}

type getConst struct {
	Value any `json:"value" mapstructure:"value"`
}

func GetRequestData(ctx context.Context) *request_data.RequestData {
	if reqData, ok := common.GetCtxState(ctx).Load(common.ContextRequestData); ok {
		return reqData.(*request_data.RequestData)
	}
	return nil
}

func (r *getErrors) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	return GetRequestData(ctx).Errors, nil
}

func (r *getRequest) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	return GetRequestData(ctx).ReqBody, nil
}

func (r *getResponse) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	return common.SyncMapGet(GetRequestData(ctx).Response, "."), nil
}

func (s *getStore) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	return GetRequestData(ctx).Store, nil
}

func (c *getConst) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	return c.Value, nil
}

func (c *getPreConfig) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	return GetRequestData(ctx).PreConfig, nil
}

func (h *getHeaders) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	return GetRequestData(ctx).Headers, nil
}
