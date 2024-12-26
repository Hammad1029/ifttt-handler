package resolvable

import (
	"context"
	"errors"
	"fmt"
	"ifttt/handler/common"
	"ifttt/handler/domain/request_data"
	"time"

	"github.com/fatih/structs"
	"github.com/mitchellh/mapstructure"
)

type queryResolvable struct {
	QueryString          string                `json:"queryString" mapstructure:"queryString"`
	Return               bool                  `json:"return" mapstructure:"return"`
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
	Return               bool           `json:"return" mapstructure:"return"`
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
	RawQueryPositional(queryString string, parameters []any, ctx context.Context) (*[]map[string]any, error)
	RawQueryNamed(queryString string, parameters map[string]any, ctx context.Context) (*[]map[string]any, error)

	RawExecPositional(queryString string, parameters []any, ctx context.Context) error
	RawExecNamed(queryString string, parameters map[string]any, ctx context.Context) error
}

func (q *queryResolvable) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	queryData, err := q.createQueryData(ctx, dependencies)
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

func (q *queryResolvable) createQueryData(ctx context.Context, dependencies map[common.IntIota]any) (*queryData, error) {
	var queryData queryData
	queryRequest, err := q.createQueryRequest(ctx, dependencies)
	if err != nil {
		return nil, fmt.Errorf("could not create query request: %s", err)
	}
	queryData.Request = queryRequest
	queryData.Metadata = q.createQueryMetadata()
	queryData.Results = &[]map[string]any{}
	return &queryData, nil
}

func (q *queryResolvable) createQueryRequest(ctx context.Context, dependencies map[common.IntIota]any) (*queryRequest, error) {
	var req queryRequest

	req.QueryString = q.QueryString
	req.Return = q.Return
	req.Named = q.Named

	if req.Named {
		parametersResolved, err := resolveIfNested(q.NamedParameters, ctx, dependencies)
		if err != nil {
			return nil, fmt.Errorf("could not resolve named parameters: %s", err)
		}
		if err := mapstructure.Decode(parametersResolved, &req.NamedParameters); err != nil {
			return nil, fmt.Errorf("could not decode resolved named parameters to map[string]any: %s", err)
		}
	} else {
		parametersResolved, err := resolveIfNested(q.PositionalParameters, ctx, dependencies)
		if err != nil {
			return nil, fmt.Errorf("could not resolve positional parameters: %s", err)
		}
		if err := mapstructure.Decode(parametersResolved, &req.PositionalParameters); err != nil {
			return nil, fmt.Errorf("could not decode resolved positional parameters to []any: %s", err)
		}
	}

	return &req, nil
}

func (q *queryResolvable) createQueryMetadata() *queryMetadata {
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
		results, err = q.Request.RunQuery(rawQueryRepo, timeoutCtx)
	} else {
		results, err = q.Request.RunQuery(rawQueryRepo, ctx)
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

func (q *queryRequest) RunQuery(rawQueryRepo RawQueryRepository, ctx context.Context) (*[]map[string]any, error) {
	if q.Named {
		namedParams := common.RegexNamedParameters.FindAllString(q.QueryString, -1)
		if len(namedParams) != len(q.NamedParameters) {
			return nil, fmt.Errorf("missing named parameters")
		}

		if q.Return {
			return rawQueryRepo.RawQueryNamed(q.QueryString, q.NamedParameters, ctx)
		} else {
			return nil, rawQueryRepo.RawExecNamed(q.QueryString, q.NamedParameters, ctx)
		}
	} else {
		positionalParams := common.RegexPositionalParameters.FindAllString(q.QueryString, -1)
		if len(positionalParams) != len(q.PositionalParameters) {
			return nil, fmt.Errorf("missing positional parameters")
		}

		if q.Return {
			return rawQueryRepo.RawQueryPositional(q.QueryString, q.PositionalParameters, ctx)
		} else {
			return nil, rawQueryRepo.RawExecPositional(q.QueryString, q.PositionalParameters, ctx)
		}
	}
}
