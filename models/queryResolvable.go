package models

import (
	"context"
	"fmt"
	"handler/scylla"

	jsontocql "github.com/Hammad1029/json-to-cql"
)

type QueryResolvable struct {
	QueryHash string     `json:"query" mapstructure:"query"`
	Get       Resolvable `json:"get" mapstructure:"get"`
	Recall    bool       `json:"recall" mapstructure:"recall"`
}

func (q *QueryResolvable) Resolve(ctx context.Context) (any, error) {
	reqData := ctx.Value("request").(*RequestData)
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

		switch currQuery.Type {
		case jsontocql.Select:
			{
				if oldRes, queryRan := reqData.QueryRes[q.QueryHash]; queryRan && !q.Recall {
					return oldRes, nil
				} else {
					if newRes, err := scylla.RunSelect(currQuery, queryParameters); err != nil {
						return nil, fmt.Errorf("method resolveQuery: could not run query: %s", err.Error())
					} else {
						reqData.QueryRes[q.QueryHash] = newRes
						return newRes, nil
					}
				}
			}
		default:
			{
				if err := scylla.RunQuery(currQuery.QueryString, queryParameters); err != nil {
					return nil,
						fmt.Errorf("method resolveQuery: error running non select query | %s", err)
				}
				return nil, nil
			}
		}
	} else {
		return nil, fmt.Errorf("method resolveQuery: query hash %s not found", q.QueryHash)
	}
}
