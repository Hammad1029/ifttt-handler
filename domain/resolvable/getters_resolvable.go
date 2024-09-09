package resolvable

import (
	"context"
	"ifttt/handler/domain/request_data"
	"maps"
)

type GetRequestResolvable map[string]any

type GetResponseResolvable map[string]any

type GetStoreResolvable map[string]any

type GetApiResultsResolvable map[string]map[string]any

type GetQueryResultsResolvable map[string][]map[string]any

type PreConfigResolvable map[string]any

type GetConstResolvable struct {
	Value any `json:"value" mapstructure:"value"`
}

func GetRequestData(ctx context.Context) *request_data.RequestData {
	reqData, _ := ctx.Value("request").(*request_data.RequestData)
	return reqData
}

func (r *GetRequestResolvable) Resolve(ctx context.Context, dependencies map[string]any) (any, error) {
	return GetRequestData(ctx).ReqBody, nil
}

func (r *GetResponseResolvable) Resolve(ctx context.Context, dependencies map[string]any) (any, error) {
	return maps.Clone(GetRequestData(ctx).Response), nil
}

func (a *GetApiResultsResolvable) Resolve(ctx context.Context, dependencies map[string]any) (any, error) {
	return GetRequestData(ctx).ApiRes, nil
}

func (s *GetStoreResolvable) Resolve(ctx context.Context, dependencies map[string]any) (any, error) {
	return GetRequestData(ctx).Store, nil
}

func (q *GetQueryResultsResolvable) Resolve(ctx context.Context, dependencies map[string]any) (any, error) {
	return GetRequestData(ctx).QueryRes, nil
}

func (c *GetConstResolvable) Resolve(ctx context.Context, dependencies map[string]any) (any, error) {
	return c.Value, nil
}

func (c *PreConfigResolvable) Resolve(ctx context.Context, dependencies map[string]any) (any, error) {
	return GetRequestData(ctx).PreConfig, nil
}
