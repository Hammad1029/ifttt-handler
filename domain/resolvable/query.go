package resolvable

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"ifttt/handler/common"
	"ifttt/handler/domain/request_data"
	"reflect"
	"strings"
	"time"

	"github.com/fatih/structs"
)

type query struct {
	QueryString string       `json:"queryString" mapstructure:"queryString"`
	Scan        bool         `json:"scan" mapstructure:"scan"`
	Parameters  []Resolvable `json:"parameters" mapstructure:"parameters"`
	Async       bool         `json:"async" mapstructure:"async"`
	Timeout     uint         `json:"timeout" mapstructure:"timeout"`
}

type queryData struct {
	Request  *queryRequest     `json:"queryRequest" mapstructure:"queryRequest"`
	Metadata *queryMetadata    `json:"queryMetadata" mapstructure:"queryMetadata"`
	Results  *[]map[string]any `json:"results" mapstructure:"results"`
}

type queryRequest struct {
	Scan        bool   `json:"scan" mapstructure:"scan"`
	QueryString string `json:"queryString" mapstructure:"queryString"`
	Parameters  []any  `json:"parameters" mapstructure:"parameters"`
}

type queryMetadata struct {
	Start        time.Time `json:"start" mapstructure:"start"`
	End          time.Time `json:"end" mapstructure:"end"`
	TimeTaken    uint64    `json:"timeTaken" mapstructure:"timeTaken"`
	Timeout      uint      `json:"timeOut" mapstructure:"timeOut"`
	DidTimeout   bool      `json:"didTimeout" mapstructure:"didTimeout"`
	Async        bool      `json:"async" mapstructure:"async"`
	Error        string    `json:"error" mapstructure:"error"`
	RowsAffected int       `json:"rowsAffected" mapstructure:"rowsAffected"`
}

type RawQueryRepository interface {
	BeginTx(ctx context.Context) (*sql.Tx, error)
	Scan(tx *sql.Tx, queryString string, parameters []any, ctx context.Context) (*[]map[string]any, int, error)
	Exec(tx *sql.Tx, queryString string, parameters []any, ctx context.Context) (int, error)
}

func (q *query) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	resolved, err := resolveArrayMustParallel(&q.Parameters, ctx, dependencies)
	if err != nil {
		return nil, fmt.Errorf("could resolve parameters for query: %s", err)
	}

	rawQueryRepo, ok := dependencies[common.DependencyRawQueryRepo].(RawQueryRepository)
	if !ok {
		return nil, fmt.Errorf("method *QueryResolvable: could not cast raw query repo")
	}

	tx, err := rawQueryRepo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not begin tx: %s", err)
	}

	queryData, err := q.init(tx, resolved, ctx, dependencies)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return nil, fmt.Errorf("attempted rollback on error: %s. rollback failed %s", err, rollbackErr)
		}
		return nil, fmt.Errorf("rolled back. error: %s", err)
	} else if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit failed: %s", err)
	}

	return queryData, nil
}

func (q *query) init(tx *sql.Tx, parameters []any, ctx context.Context, dependencies map[common.IntIota]any) (*queryData, error) {
	queryData, err := q.createQueryData(parameters)
	if err != nil {
		return nil, fmt.Errorf("queryResolvable: could not create query data: %s", err)
	}

	if q.Async {
		go queryData.execute(tx, dependencies, ctx)
	} else if err := queryData.execute(tx, dependencies, ctx); err != nil {
		return nil, fmt.Errorf("queryResolvable: could not execute query: %s", err)
	}

	return queryData, nil
}

func (q *query) createQueryData(parameters []any) (*queryData, error) {
	req := queryRequest{
		Scan:        q.Scan,
		QueryString: q.QueryString,
		Parameters:  []any{},
	}
	req.preProcess(parameters)

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

func (q *queryData) execute(tx *sql.Tx, dependencies map[common.IntIota]any, ctx context.Context) error {
	defer func() {
		mapped := structs.Map(q)
		request_data.AddExternalTrip(common.ExternalTripQuery,
			q.Request.QueryString[20:],
			&mapped, q.Metadata.TimeTaken, ctx)
	}()

	q.Metadata.Start = time.Now()

	rawQueryRepo, ok := dependencies[common.DependencyRawQueryRepo].(RawQueryRepository)
	if !ok {
		return fmt.Errorf("method *QueryResolvable: could not cast raw query repo")
	}

	var (
		results      *[]map[string]any
		rowsAffected int
		err          error
	)
	if q.Metadata.Timeout > 0 {
		timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(q.Metadata.Timeout)*time.Millisecond)
		defer cancel()
		results, rowsAffected, err = q.Request.runQuery(tx, rawQueryRepo, timeoutCtx)
	} else {
		results, rowsAffected, err = q.Request.runQuery(tx, rawQueryRepo, ctx)
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
	q.Metadata.RowsAffected = rowsAffected
	q.Metadata.End = time.Now()
	q.Metadata.TimeTaken = uint64(q.Metadata.End.Sub(q.Metadata.Start).Milliseconds())

	return nil
}

func (q *queryRequest) preProcess(parameters []any) {
	var builder strings.Builder
	builder.Grow(len(q.QueryString))

	estimatedParams := strings.Count(q.QueryString, "?")
	q.Parameters = make([]any, 0, estimatedParams)

	var (
		afterParenthesis bool
		idx              int
	)

	for _, v := range q.QueryString {
		if v == '?' && idx < len(parameters) {
			if afterParenthesis {
				rv := reflect.ValueOf(parameters[idx])
				switch rv.Kind() {
				case reflect.Slice, reflect.Array:
					if rv.Len() == 0 {
						builder.WriteRune('?')
						q.Parameters = append(q.Parameters, nil)
					} else {
						if rv.Len() > 1 {
							builder.WriteRune('?')
							for i := 1; i < rv.Len(); i++ {
								builder.WriteString(",?")
							}
						} else {
							builder.WriteRune('?')
						}
						for i := 0; i < rv.Len(); i++ {
							q.Parameters = append(q.Parameters, rv.Index(i).Interface())
						}
					}
				default:
					builder.WriteRune('?')
					q.Parameters = append(q.Parameters, parameters[idx])
				}
			} else {
				builder.WriteRune('?')
				q.Parameters = append(q.Parameters, parameters[idx])
			}
			idx++
		} else {
			afterParenthesis = v == '('
			builder.WriteRune(v)
		}
	}

	q.QueryString = builder.String()
}

func (q *queryRequest) runQuery(tx *sql.Tx, rawQueryRepo RawQueryRepository, ctx context.Context) (*[]map[string]any, int, error) {
	var (
		results      *[]map[string]any
		rowsAffected int
		err          error
	)

	if q.Scan {
		results, rowsAffected, err = rawQueryRepo.Scan(tx, q.QueryString, q.Parameters, ctx)
	} else {
		rowsAffected, err = rawQueryRepo.Exec(tx, q.QueryString, q.Parameters, ctx)
	}

	return results, rowsAffected, err
}
