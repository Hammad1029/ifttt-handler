package resolvable

import (
	"context"
	"ifttt/handler/domain/request_data"
	"maps"
)

type getRequestResolvable map[string]any

type getResponseResolvable map[string]any

type getStoreResolvable map[string]any

type getApiResultsResolvable map[string]map[string]any

type getQueryResultsResolvable map[string][]map[string]any

type preConfigResolvable map[string]any

type getConstResolvable struct {
	Value any `json:"value" mapstructure:"value"`
}

func GetRequestData(ctx context.Context) *request_data.RequestData {
	reqData, _ := ctx.Value("request").(*request_data.RequestData)
	return reqData
}

func (r *getRequestResolvable) Resolve(ctx context.Context, dependencies map[string]any) (any, error) {
	return GetRequestData(ctx).ReqBody, nil
}

func (r *getResponseResolvable) Resolve(ctx context.Context, dependencies map[string]any) (any, error) {
	return maps.Clone(GetRequestData(ctx).Response), nil
}

func (a *getApiResultsResolvable) Resolve(ctx context.Context, dependencies map[string]any) (any, error) {
	return GetRequestData(ctx).ApiRes, nil
}

func (s *getStoreResolvable) Resolve(ctx context.Context, dependencies map[string]any) (any, error) {
	return GetRequestData(ctx).Store, nil
}

func (q *getQueryResultsResolvable) Resolve(ctx context.Context, dependencies map[string]any) (any, error) {
	return GetRequestData(ctx).QueryRes, nil
}

func (c *getConstResolvable) Resolve(ctx context.Context, dependencies map[string]any) (any, error) {
	return c.Value, nil
}

func (c *preConfigResolvable) Resolve(ctx context.Context, dependencies map[string]any) (any, error) {
	return GetRequestData(ctx).PreConfig, nil
}
