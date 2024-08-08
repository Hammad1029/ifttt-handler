package resolvable

import (
	"context"
	"fmt"
	"handler/common"
	"handler/domain/request_data"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
)

const (
	dmlSelect = "SELECT"
	dmlInsert = "INSERT"
	dmlUpdate = "UPDATE"
	dmlDelete = "DELETE"
)

const (
	basicQuery      = "BASIC"
	namedQuery      = "NAMED"
	positionalQuery = "POSITIONAL"
)

type RawQueryRepository interface {
	RawSelect(queryString string) ([]common.JsonObject, error)
	RawSelectNamed(queryString string, parameters common.JsonObject) ([]common.JsonObject, error)
	RawSelectPositional(queryString string, parameters []any) ([]common.JsonObject, error)

	RawQuery(queryString string) ([]common.JsonObject, error)
	RawQueryNamed(queryString string, parameters common.JsonObject) ([]common.JsonObject, error)
	RawQueryPositional(queryString string, parameters []any) ([]common.JsonObject, error)
}

type QueryExecutor interface {
	Execute(rawQueryRepo RawQueryRepository, q *QueryResolvable, ctx context.Context, optional []any) ([]common.JsonObject, error)
}

type BasicQueryExecutor struct{}

type NamedQueryExecutor struct{}

type PositionalQueryExecutor struct{}

type QueryResolvable struct {
	QueryString          string                `json:"queryString" mapstructure:"queryString"`
	QueryHash            string                `json:"queryHash" mapstructure:"queryHash"`
	Dml                  string                `json:"dml" mapstructure:"dml"`
	Dynamic              string                `json:"dynamic" mapstructure:"dynamic"`
	NamedParameters      map[string]Resolvable `json:"namedParameters" mapstructure:"namedParameters"`
	PositionalParameters []Resolvable          `json:"positionalParameters" mapstructure:"positionalParameters"`
}

func (e *BasicQueryExecutor) Execute(rawQueryRepo RawQueryRepository, q *QueryResolvable, ctx context.Context, optional []any) ([]common.JsonObject, error) {
	if q.Dml == dmlSelect {
		return rawQueryRepo.RawSelect(q.QueryString)
	}
	return rawQueryRepo.RawQuery(q.QueryString)
}

func (e *NamedQueryExecutor) Execute(rawQueryRepo RawQueryRepository, q *QueryResolvable, ctx context.Context, optional []any) ([]common.JsonObject, error) {
	parametersResolved, err := resolveIfNested(q.NamedParameters, ctx, optional)
	if err != nil {
		return nil, fmt.Errorf("could not resolve named parameters: %s", err)
	}
	var namedParametersMap common.JsonObject
	if err := mapstructure.Decode(parametersResolved, &namedParametersMap); err != nil {
		return nil, fmt.Errorf("could not decode resolved named parameters to JsonObject: %s", err)
	}
	if q.Dml == dmlSelect {
		return rawQueryRepo.RawSelectNamed(q.QueryString, namedParametersMap)
	}
	return rawQueryRepo.RawQueryNamed(q.QueryString, namedParametersMap)
}

func (e *PositionalQueryExecutor) Execute(rawQueryRepo RawQueryRepository, q *QueryResolvable, ctx context.Context, optional []any) ([]common.JsonObject, error) {
	parametersResolved, err := resolveIfNested(q.PositionalParameters, ctx, optional)
	if err != nil {
		return nil, fmt.Errorf("could not resolve positional parameters: %s", err)
	}
	var positionalParametersSlice []any
	if err := mapstructure.Decode(parametersResolved, &positionalParametersSlice); err != nil {
		return nil, fmt.Errorf("could not decode resolved positional parameters to []any: %s", err)
	}

	if q.Dml == dmlSelect {
		return rawQueryRepo.RawSelectPositional(q.QueryString, positionalParametersSlice)
	}
	return rawQueryRepo.RawQueryPositional(q.QueryString, positionalParametersSlice)
}

func (q *QueryResolvable) Resolve(ctx context.Context, optional ...any) (any, error) {
	var queryResult request_data.QueryResult
	queryResult.Start = time.Now()

	q.Dml = strings.ToUpper(q.Dml)
	if q.Dml != dmlSelect && q.Dml != dmlInsert && q.Dml != dmlUpdate && q.Dml != dmlDelete {
		return nil, fmt.Errorf("method *BasicQueryExecutor.Execute: unsupported DML type %s", q.Dml)
	}
	q.Dynamic = strings.ToUpper(q.Dynamic)

	rawQueryRepo, ok := optional[0].(RawQueryRepository)
	if !ok {
		return nil, fmt.Errorf("method *QueryResolvable: could not cast raw query repo")
	}

	queryExecutor, err := queryExecutorFactory(q.Dynamic)
	if err != nil {
		return nil, fmt.Errorf("method *QueryResolvable.Resolve: %v", err)
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

func queryExecutorFactory(dynamic string) (QueryExecutor, error) {
	switch dynamic {
	case basicQuery:
		return &BasicQueryExecutor{}, nil
	case namedQuery:
		return &NamedQueryExecutor{}, nil
	case positionalQuery:
		return &PositionalQueryExecutor{}, nil
	default:
		return nil, fmt.Errorf("method queryExecutorFactory: unsupported dynamic type: %s", dynamic)
	}
}
