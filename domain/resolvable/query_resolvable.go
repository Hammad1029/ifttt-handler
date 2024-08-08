package resolvable

import (
	"context"
	"fmt"
	"handler/common"
	"handler/domain/request_data"
	"time"

	"github.com/mitchellh/mapstructure"
)

type RawQueryRepository interface {
	RawQueryPositional(queryString string, parameters []any) (*[]common.JsonObject, error)
	RawQueryNamed(queryString string, parameters common.JsonObject) (*[]common.JsonObject, error)

	RawExecPositional(queryString string, parameters []any) error
	RawExecNamed(queryString string, parameters common.JsonObject) error
}

type QueryExecutor interface {
	Execute(rawQueryRepo RawQueryRepository, q *QueryResolvable, ctx context.Context, optional []any) (*[]common.JsonObject, error)
}

type NamedQueryExecutor struct{}

type PositionalQueryExecutor struct{}

type QueryResolvable struct {
	QueryString          string                `json:"queryString" mapstructure:"queryString"`
	QueryHash            string                `json:"queryHash" mapstructure:"queryHash"`
	Exec                 bool                  `json:"exec" mapstructure:"exec"`
	Named                bool                  `json:"named" mapstructure:"named"`
	NamedParameters      map[string]Resolvable `json:"namedParameters" mapstructure:"namedParameters"`
	PositionalParameters []Resolvable          `json:"positionalParameters" mapstructure:"positionalParameters"`
}

func (e *NamedQueryExecutor) Execute(rawQueryRepo RawQueryRepository, q *QueryResolvable, ctx context.Context, optional []any) (*[]common.JsonObject, error) {
	parametersResolved, err := resolveIfNested(q.NamedParameters, ctx, optional)
	if err != nil {
		return nil, fmt.Errorf("could not resolve named parameters: %s", err)
	}
	var namedParametersMap common.JsonObject
	if err := mapstructure.Decode(parametersResolved, &namedParametersMap); err != nil {
		return nil, fmt.Errorf("could not decode resolved named parameters to JsonObject: %s", err)
	}
	if q.Exec {
		return nil, rawQueryRepo.RawExecNamed(q.QueryString, namedParametersMap)
	}
	return rawQueryRepo.RawQueryNamed(q.QueryString, namedParametersMap)
}

func (e *PositionalQueryExecutor) Execute(rawQueryRepo RawQueryRepository, q *QueryResolvable, ctx context.Context, optional []any) (*[]common.JsonObject, error) {
	parametersResolved, err := resolveIfNested(q.PositionalParameters, ctx, optional)
	if err != nil {
		return nil, fmt.Errorf("could not resolve positional parameters: %s", err)
	}
	var positionalParametersSlice []any
	if err := mapstructure.Decode(parametersResolved, &positionalParametersSlice); err != nil {
		return nil, fmt.Errorf("could not decode resolved positional parameters to []any: %s", err)
	}

	if q.Exec {
		return nil, rawQueryRepo.RawExecPositional(q.QueryString, positionalParametersSlice)
	}
	return rawQueryRepo.RawQueryPositional(q.QueryString, positionalParametersSlice)
}

func (q *QueryResolvable) Resolve(ctx context.Context, optional ...any) (any, error) {
	var queryResult request_data.QueryResult
	queryResult.Start = time.Now()

	rawQueryRepo, ok := optional[0].(RawQueryRepository)
	if !ok {
		return nil, fmt.Errorf("method *QueryResolvable: could not cast raw query repo")
	}

	var queryExecutor QueryExecutor
	if q.Named {
		queryExecutor = &NamedQueryExecutor{}
	} else {
		queryExecutor = &PositionalQueryExecutor{}
	}

	results, err := queryExecutor.Execute(rawQueryRepo, q, ctx, optional)
	if err != nil {
		return nil, fmt.Errorf("method resolveQuery: could not run query %s: %s", q.QueryHash, err.Error())
	}
	queryResult.Results = results
	queryResult.End = time.Now()
	queryResult.TimeTaken = queryResult.End.Sub(queryResult.Start).Milliseconds()

	queryRes := ctx.Value("request").(*request_data.RequestData).QueryRes
	queryRes[q.QueryHash] = append(queryRes[q.QueryHash], queryResult)
	return results, nil
}
