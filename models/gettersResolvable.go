package models

import (
	"context"
	"handler/common"
	"maps"
)

type RequestResolvable common.JsonObject

type ResponseResolvable common.JsonObject

type StoreResolvable common.JsonObject

type ApiResultsResolvable map[string]common.JsonObject

type QueryResultsResolvable map[string][]common.JsonObject

func getRequestData(ctx context.Context) *RequestData {
	reqData, _ := ctx.Value("request").(*RequestData)
	return reqData
}

func (r *RequestResolvable) Resolve(ctx context.Context) (any, error) {
	return getRequestData(ctx).ReqBody, nil
}

func (r *ResponseResolvable) Resolve(ctx context.Context) (any, error) {
	return maps.Clone(getRequestData(ctx).Response), nil
}

func (a *ApiResultsResolvable) Resolve(ctx context.Context) (any, error) {
	return getRequestData(ctx).ApiRes, nil
}

func (s *StoreResolvable) Resolve(ctx context.Context) (any, error) {
	return getRequestData(ctx).Store, nil
}

func (q *QueryResultsResolvable) Resolve(ctx context.Context) (any, error) {
	return getRequestData(ctx).QueryRes, nil
}

func ResolveConstant(c any, ctx context.Context) (any, error) {
	return c, nil
}
