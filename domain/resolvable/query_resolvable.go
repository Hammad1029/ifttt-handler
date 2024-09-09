package resolvable

import (
	"context"
	"fmt"
	"ifttt/handler/common"
	"ifttt/handler/domain/request_data"
	"time"

	"github.com/mitchellh/mapstructure"
)

type RawQueryRepository interface {
	RawQueryPositional(queryString string, parameters []any) (*[]map[string]any, error)
	RawQueryNamed(queryString string, parameters map[string]any) (*[]map[string]any, error)

	RawExecPositional(queryString string, parameters []any) error
	RawExecNamed(queryString string, parameters map[string]any) error
}

type QueryExecutor interface {
	Execute(rawQueryRepo RawQueryRepository, q *QueryResolvable, ctx context.Context, dependencies map[string]any) (*[]map[string]any, error)
}

type NamedQueryExecutor struct{}

type PositionalQueryExecutor struct{}

type QueryResolvable struct {
	QueryString          string                `json:"queryString" mapstructure:"queryString"`
	QueryHash            string                `json:"queryHash" mapstructure:"queryHash"`
	Return               bool                  `json:"return" mapstructure:"return"`
	Named                bool                  `json:"named" mapstructure:"named"`
	NamedParameters      map[string]Resolvable `json:"namedParameters" mapstructure:"namedParameters"`
	PositionalParameters []Resolvable          `json:"positionalParameters" mapstructure:"positionalParameters"`
}

func (e *NamedQueryExecutor) Execute(rawQueryRepo RawQueryRepository, q *QueryResolvable, ctx context.Context, dependencies map[string]any) (*[]map[string]any, error) {
	parametersResolved, err := resolveIfNested(q.NamedParameters, ctx, dependencies)
	if err != nil {
		return nil, fmt.Errorf("could not resolve named parameters: %s", err)
	}
	var namedParametersMap map[string]any
	if err := mapstructure.Decode(parametersResolved, &namedParametersMap); err != nil {
		return nil, fmt.Errorf("could not decode resolved named parameters to map[string]any: %s", err)
	}
	if q.Return {
		return rawQueryRepo.RawQueryNamed(q.QueryString, namedParametersMap)
	}
	return nil, rawQueryRepo.RawExecNamed(q.QueryString, namedParametersMap)
}

func (e *PositionalQueryExecutor) Execute(rawQueryRepo RawQueryRepository, q *QueryResolvable, ctx context.Context, dependencies map[string]any) (*[]map[string]any, error) {
	parametersResolved, err := resolveIfNested(q.PositionalParameters, ctx, dependencies)
	if err != nil {
		return nil, fmt.Errorf("could not resolve positional parameters: %s", err)
	}
	var positionalParametersSlice []any
	if err := mapstructure.Decode(parametersResolved, &positionalParametersSlice); err != nil {
		return nil, fmt.Errorf("could not decode resolved positional parameters to []any: %s", err)
	}

	if q.Return {
		return rawQueryRepo.RawQueryPositional(q.QueryString, positionalParametersSlice)
	}
	return nil, rawQueryRepo.RawExecPositional(q.QueryString, positionalParametersSlice)
}

func (q *QueryResolvable) Resolve(ctx context.Context, dependencies map[string]any) (any, error) {
	var queryResult request_data.QueryResult
	queryResult.Start = time.Now()

	rawQueryRepo, ok := dependencies[common.DependencyRawQueryRepo].(RawQueryRepository)
	if !ok {
		return nil, fmt.Errorf("method *QueryResolvable: could not cast raw query repo")
	}

	var queryExecutor QueryExecutor
	if q.Named {
		queryExecutor = &NamedQueryExecutor{}
	} else {
		queryExecutor = &PositionalQueryExecutor{}
	}

	results, err := queryExecutor.Execute(rawQueryRepo, q, ctx, dependencies)
	if err != nil {
		return nil, fmt.Errorf("method resolveQuery: could not run query %s: %s", q.QueryHash, err.Error())
	}
	queryResult.Results = results
	queryResult.End = time.Now()
	queryResult.TimeTaken = queryResult.End.Sub(queryResult.Start).Milliseconds()

	queryRes := GetRequestData(ctx).QueryRes
	queryRes[q.QueryHash] = append(queryRes[q.QueryHash], queryResult)
	return results, nil
}
