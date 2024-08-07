package resolvable

import (
	"context"
	"handler/common"
	"handler/domain/request_data"
	"maps"
)

type GetRequestResolvable common.JsonObject

type GetResponseResolvable common.JsonObject

type GetStoreResolvable common.JsonObject

type GetApiResultsResolvable map[string]common.JsonObject

type GetQueryResultsResolvable map[string][]common.JsonObject

type GetConstResolvable struct {
	Value any `json:"value" mapstructure:"value"`
}

func getRequestData(ctx context.Context) *request_data.RequestData {
	reqData, _ := ctx.Value("request").(*request_data.RequestData)
	return reqData
}

func (r *GetRequestResolvable) Resolve(ctx context.Context) (any, error) {
	return getRequestData(ctx).ReqBody, nil
}

func (r *GetResponseResolvable) Resolve(ctx context.Context) (any, error) {
	return maps.Clone(getRequestData(ctx).Response), nil
}

func (a *GetApiResultsResolvable) Resolve(ctx context.Context) (any, error) {
	return getRequestData(ctx).ApiRes, nil
}

func (s *GetStoreResolvable) Resolve(ctx context.Context) (any, error) {
	return getRequestData(ctx).Store, nil
}

func (q *GetQueryResultsResolvable) Resolve(ctx context.Context) (any, error) {
	return getRequestData(ctx).QueryRes, nil
}

func (c *GetConstResolvable) Resolve(ctx context.Context) (any, error) {
	return c.Value, nil
}
