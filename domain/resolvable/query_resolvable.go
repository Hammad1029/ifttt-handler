package resolvable

import (
	"context"
	"fmt"
	"handler/common"
	"handler/domain/request_data"
	"handler/store"

	jsontocql "github.com/Hammad1029/json-to-cql"
)

type QueryResolvable struct {
	QueryHash string `json:"query" mapstructure:"query"`
	Recall    bool   `json:"recall" mapstructure:"recall"`
}

func (q *QueryResolvable) Resolve(ctx context.Context) (any, error) {
	var results []common.JsonObject
	dataStore := *store.GetDataStore()
	reqData := ctx.Value("request").(*request_data.RequestData)
	queries := ctx.Value("queries").(map[string]jsontocql.ParameterizedQuery)

	if currQuery, ok := queries[q.QueryHash]; ok {
		var queryParameters []any
		var localResolvable Resolvable
		for _, param := range currQuery.Resolvables {
			localResolvable = Resolvable{
				ResolveType: param.ResolveType,
				ResolveData: param.ResolveData,
			}
			if p, err := localResolvable.Resolve(ctx); err != nil {
				return nil, fmt.Errorf("method resolveQuery: could not resolve query parameters: %s", err)
			} else {
				queryParameters = append(queryParameters, p)
			}
		}

		queryString, err := currQuery.ResolveQuery(queryParameters)
		if err != nil {
			return nil, fmt.Errorf("method RunSelect: error resolving query: %s", err)
		}

		switch currQuery.Type {
		case jsontocql.Select:
			{
				if oldRes, queryRan := reqData.QueryRes[q.QueryHash]; queryRan && !q.Recall {
					results = oldRes
				} else {
					if newRes, err := dataStore.RawSelect(queryString, queryParameters); err != nil {
						return nil, fmt.Errorf("method resolveQuery: could not run query: %s", err.Error())
					} else {
						reqData.QueryRes[q.QueryHash] = newRes
						results = newRes
					}
				}
			}
		default:
			{
				if res, err := dataStore.RawQuery(queryString, queryParameters); err != nil {
					return nil,
						fmt.Errorf("method resolveQuery: error running non select query | %s", err)
				} else {
					results = res
				}
			}
		}
		return results, nil
	} else {
		return nil, fmt.Errorf("method resolveQuery: query hash %s not found", q.QueryHash)
	}
}
