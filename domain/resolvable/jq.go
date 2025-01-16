package resolvable

import (
	"context"
	"encoding/json"
	"fmt"
	"ifttt/handler/common"
	"strings"
	"sync"

	"github.com/itchyny/gojq"
)

type jq struct {
	Query any `json:"query" mapstructure:"query"`
	Input any `json:"input" mapstructure:"input"`
}

func (j *jq) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	queryResolved, err := resolveMaybe(j.Query, ctx, dependencies)
	if err != nil {
		return nil, err
	}
	inputResolved, err := resolveMaybe(j.Input, ctx, dependencies)
	if err != nil {
		return nil, err
	}
	jqCompatibleInput, err := convertToGoJQCompatible(inputResolved)
	if err != nil {
		return nil, err
	}

	return runJQQuery(fmt.Sprint(queryResolved), jqCompatibleInput)
}

func runJQQuery(queryString string, input any) (any, error) {
	arrayReturn := strings.Contains(queryString, "[]")
	if queryString[0] != '.' {
		queryString = "." + queryString
	}
	query, err := gojq.Parse(queryString)
	if err != nil {
		return nil, fmt.Errorf("error parsing jq query: %s error: %s", queryString, err)
	}

	var resultVals []any
	resultIter := query.Run((any)(input))

	for {
		v, ok := resultIter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			if haltErr, ok := err.(*gojq.HaltError); ok && haltErr.Value() == nil {
				break
			}
			return nil, err
		}
		resultVals = append(resultVals, v)
	}

	switch len(resultVals) {
	case 0:
		if arrayReturn {
			return []any{}, nil
		}
		return nil, nil
	case 1:
		if arrayReturn {
			return []any{resultVals[0]}, nil
		}
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
