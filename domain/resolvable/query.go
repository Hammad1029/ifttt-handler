package resolvable

import (
	"context"
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
	QueryString          string                `json:"queryString" mapstructure:"queryString"`
	Scan                 bool                  `json:"scan" mapstructure:"scan"`
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
	Operation            string         `json:"operation" mapstructure:"operation"`
	Table                string         `json:"table" mapstructure:"table"`
	Scan                 bool           `json:"scan" mapstructure:"scan"`
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
	ScanPositional(queryString string, parameters []any, ctx context.Context) (*[]map[string]any, error)
	ScanNamed(queryString string, parameters map[string]any, ctx context.Context) (*[]map[string]any, error)
	ExecPositional(queryString string, parameters []any, ctx context.Context) error
	ExecNamed(queryString string, parameters map[string]any, ctx context.Context) error
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
	queryData, err := q.createQueryData(positional, named)
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

func (q *query) createQueryData(positional []any, named map[string]any) (*queryData, error) {
	operation, table := common.ExtractFromQuery(q.QueryString)
	req := queryRequest{
		Operation:            operation,
		Table:                table,
		Scan:                 q.Scan,
		QueryString:          q.QueryString,
		Named:                q.Named,
		PositionalParameters: []any{},
		NamedParameters:      map[string]any{},
	}
	req.preProcess(positional, named)

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
		mapped := structs.Map(q)
		request_data.AddExternalTrip(common.ExternalTripQuery, fmt.Sprintf("%s:%s", q.Request.Operation, q.Request.Table), &mapped, q.Metadata.TimeTaken, ctx)
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
func (q *queryRequest) preProcess(positional []any, named map[string]any) {
	if q.Named || named != nil {
		return
	}

	var builder strings.Builder
	builder.Grow(len(q.QueryString))

	estimatedParams := strings.Count(q.QueryString, "?")
	q.PositionalParameters = make([]any, 0, estimatedParams)

	var (
		afterParenthesis bool
		idx              int
	)

	for _, v := range q.QueryString {
		if v == '?' && idx < len(positional) {
			if afterParenthesis {
				rv := reflect.ValueOf(positional[idx])
				switch rv.Kind() {
				case reflect.Slice, reflect.Array:
					if rv.Len() == 0 {
						builder.WriteRune('?')
						q.PositionalParameters = append(q.PositionalParameters, nil)
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
							q.PositionalParameters = append(q.PositionalParameters, rv.Index(i).Interface())
						}
					}
				default:
					builder.WriteRune('?')
					q.PositionalParameters = append(q.PositionalParameters, positional[idx])
				}
			} else {
				builder.WriteRune('?')
				q.PositionalParameters = append(q.PositionalParameters, positional[idx])
			}
			idx++
		} else {
			afterParenthesis = v == '('
			builder.WriteRune(v)
		}
	}

	q.QueryString = builder.String()
}

func (q *queryRequest) runQuery(rawQueryRepo RawQueryRepository, ctx context.Context) (*[]map[string]any, error) {
	switch {
	case q.Named && q.Scan:
		return rawQueryRepo.ScanNamed(q.QueryString, q.NamedParameters, ctx)
	case q.Named && !q.Scan:
		return nil, rawQueryRepo.ExecNamed(q.QueryString, q.NamedParameters, ctx)
	case !q.Named && q.Scan:
		return rawQueryRepo.ScanPositional(q.QueryString, q.PositionalParameters, ctx)
	case !q.Named && !q.Scan:
		return nil, rawQueryRepo.ExecPositional(q.QueryString, q.PositionalParameters, ctx)
	default:
		return nil, fmt.Errorf("no run query case matched")
	}
}
