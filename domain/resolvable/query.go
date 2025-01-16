package resolvable

import (
	"context"
	"errors"
	"fmt"
	"ifttt/handler/common"
	"ifttt/handler/domain/request_data"
	"time"

	"github.com/fatih/structs"
)

type query struct {
	QueryString          string                `json:"queryString" mapstructure:"queryString"`
	Named                bool                  `json:"named" mapstructure:"named"`
	NamedParameters      map[string]Resolvable `json:"namedParameters" mapstructure:"namedParameters"`
	PositionalParameters []Resolvable          `json:"positionalParameters" mapstructure:"positionalParameters"`
	Async                bool                  `json:"async" mapstructure:"async"`
	Timeout              uint                  `json:"timeout" mapstructure:"timeout"`
}

type queryData struct {
	Request  *queryRequest     `json:"queryRequest" mapstructure:"queryRequest"`
	Metadata *queryMetadata    `json:"queryMetadata" mapstructure:"queryMetadata"`
	Results  *[]map[string]any `json:"results" mapstructure:"results"`
}

type queryRequest struct {
	QueryString          string         `json:"queryString" mapstructure:"queryString"`
	Named                bool           `json:"named" mapstructure:"named"`
	NamedParameters      map[string]any `json:"namedParameters" mapstructure:"namedParameters"`
	PositionalParameters []any          `json:"positionalParameters" mapstructure:"positionalParameters"`
}

type queryMetadata struct {
	Start      time.Time `json:"start" mapstructure:"start"`
	End        time.Time `json:"end" mapstructure:"end"`
	TimeTaken  uint64    `json:"timeTaken" mapstructure:"timeTaken"`
	Timeout    uint      `json:"timeOut" mapstructure:"timeOut"`
	DidTimeout bool      `json:"didTimeout" mapstructure:"didTimeout"`
	Async      bool      `json:"async" mapstructure:"async"`
	Error      string    `json:"error" mapstructure:"error"`
}

type RawQueryRepository interface {
	Positional(queryString string, parameters []any, ctx context.Context) (*[]map[string]any, error)
	Named(queryString string, parameters map[string]any, ctx context.Context) (*[]map[string]any, error)
}

func (q *query) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	var (
		namedResolved      map[string]any
		positionalResolved []any
		err                error
	)

	if q.Named {
		namedResolved, err = resolveMapMustParallel(&q.NamedParameters, ctx, dependencies)
	} else {
		positionalResolved, err = resolveArrayMustParallel(&q.PositionalParameters, ctx, dependencies)
	}

	if err != nil {
		return nil, fmt.Errorf("could resolve parameters for query: %s", err)
	} else {
		return q.init(positionalResolved, namedResolved, ctx, dependencies)
	}
}

func (q *query) init(positional []any, named map[string]any, ctx context.Context, dependencies map[common.IntIota]any) (*queryData, error) {
	queryData, err := q.createQueryData(positional, named, ctx, dependencies)
	if err != nil {
		return nil, fmt.Errorf("queryResolvable: could not create query data: %s", err)
	}

	if q.Async {
		go queryData.execute(dependencies, ctx)
	} else if err := queryData.execute(dependencies, ctx); err != nil {
		return nil, fmt.Errorf("queryResolvable: could not execute query: %s", err)
	}

	return queryData, nil
}

func (q *query) createQueryData(positional []any, named map[string]any, ctx context.Context, dependencies map[common.IntIota]any) (*queryData, error) {
	req := queryRequest{
		QueryString:          q.QueryString,
		Named:                q.Named,
		NamedParameters:      named,
		PositionalParameters: positional,
	}
	queryData := queryData{
		Request:  &req,
		Metadata: q.createQueryMetadata(),
		Results:  &[]map[string]any{},
	}
	return &queryData, nil
}

func (q *query) createQueryMetadata() *queryMetadata {
	return &queryMetadata{Timeout: q.Timeout, Async: q.Async}
}

func (q *queryData) execute(dependencies map[common.IntIota]any, ctx context.Context) error {
	defer func() {
		request_data.AddExternalTrip(common.ExternalTripQuery, structs.Map(q), q.Metadata.TimeTaken, ctx)
	}()

	q.Metadata.Start = time.Now()

	rawQueryRepo, ok := dependencies[common.DependencyRawQueryRepo].(RawQueryRepository)
	if !ok {
		return fmt.Errorf("method *QueryResolvable: could not cast raw query repo")
	}

	var (
		results *[]map[string]any
		err     error
	)
	if q.Metadata.Timeout > 0 {
		timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(q.Metadata.Timeout)*time.Millisecond)
		defer cancel()
		results, err = q.Request.runQuery(rawQueryRepo, timeoutCtx)
	} else {
		results, err = q.Request.runQuery(rawQueryRepo, ctx)
	}
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			q.Metadata.DidTimeout = true
		} else {
			common.LogWithTracer(common.LogUser, "error in executing query", err.Error(), true, ctx)
			q.Metadata.Error = err.Error()
			return err
		}
		return nil
	}

	q.Results = results
	q.Metadata.End = time.Now()
	q.Metadata.TimeTaken = uint64(q.Metadata.End.Sub(q.Metadata.Start).Milliseconds())

	return nil
}

func (q *queryRequest) runQuery(rawQueryRepo RawQueryRepository, ctx context.Context) (*[]map[string]any, error) {
	if q.Named {
		return rawQueryRepo.Named(q.QueryString, q.NamedParameters, ctx)
	} else {
		return rawQueryRepo.Positional(q.QueryString, q.PositionalParameters, ctx)
	}
}
