package resolvable

import (
	"context"
	"encoding/json"
	"fmt"
	"ifttt/handler/common"
	"sync"

	"github.com/itchyny/gojq"
)

type jqResolvable struct {
	Query any `json:"query" mapstructure:"query"`
	Input any `json:"input" mapstructure:"input"`
}

func (j *jqResolvable) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	queryResolved, err := resolveIfNested(j.Query, ctx, dependencies)
	if err != nil {
		return nil, err
	}
	inputResolved, err := resolveIfNested(j.Input, ctx, dependencies)
	if err != nil {
		return nil, err
	}

	return runJQQuery(fmt.Sprint(queryResolved), inputResolved)
}

func runJQQuery(queryString string, input any) (any, error) {
	jqInput, err := convertToGoJQCompatible(input)
	if err != nil {
		return nil, err
	}

	if queryString[0] != '.' {
		queryString = "." + queryString
	}
	query, err := gojq.Parse(queryString)
	if err != nil {
		return nil, fmt.Errorf("error parsing jq query: %s error: %s", queryString, err)
	}

	var resultVals []any
	resultIter := query.Run((any)(jqInput))

	for {
		v, ok := resultIter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			if haltErr, ok := err.(*gojq.HaltError); ok && haltErr.Value() == nil {
				break
			}
			return nil, fmt.Errorf("invalid jq setup with query %s: %s", queryString, err)
		}
		resultVals = append(resultVals, v)
	}

	switch len(resultVals) {
	case 0:
		return nil, nil
	case 1:
		return resultVals[0], nil
	default:
		return resultVals, nil
	}
}

func convertToGoJQCompatible(input any) (any, error) {
	switch o := input.(type) {
	case *sync.Map:
		return convertToGoJQCompatible(common.SyncMapUnsync(o))
	default:
		{
			marshalled, err := json.Marshal(input)
			if err != nil {
				return nil, err
			}
			var a any
			if err := json.Unmarshal(marshalled, &a); err != nil {
				return nil, err
			}
			return a, nil
		}
	}
}
