package resolvable

import (
	"context"
	"fmt"
	"ifttt/handler/common"

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
	jqCompatibleInput, err := common.ConvertToGoJQCompatible(inputResolved)
	if err != nil {
		return nil, err
	}

	return runJQQuery(fmt.Sprint(queryResolved), jqCompatibleInput)
}

func runJQQuery(queryString string, input any) (any, error) {
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
		return nil, nil
	case 1:
		return resultVals[0], nil
	default:
		return resultVals, nil
	}
}
