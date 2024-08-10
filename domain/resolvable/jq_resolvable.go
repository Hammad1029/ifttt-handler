package resolvable

import (
	"context"
	"fmt"

	"github.com/itchyny/gojq"
)

type JqResolvable struct {
	Query string `json:"query" mapstructure:"query"`
	Input any    `json:"input" mapstructure:"input"`
}

func (j *JqResolvable) Resolve(ctx context.Context, dependencies map[string]any) (any, error) {
	queryResolved, err := resolveIfNested(j.Query, ctx, dependencies)
	if err != nil {
		return nil, fmt.Errorf("method resolveJq: couldn't resolve input: %s", err)
	}
	inputResolved, err := resolveIfNested(j.Input, ctx, dependencies)
	if err != nil {
		return nil, fmt.Errorf("method resolveJq: couldn't resolve input: %s", err)
	}

	return runJQQuery(fmt.Sprint(queryResolved), inputResolved)
}

func runJQQuery(queryString string, input any) (any, error) {
	input = convertToGoJQCompatible(input)

	query, err := gojq.Parse(queryString)
	if err != nil {
		return nil, fmt.Errorf("method runJQQuery: could not parse gojq query: %s", err)
	}

	var resultVals []any
	resultIter := query.Run(input)

	for {
		v, ok := resultIter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			if err, ok := err.(*gojq.HaltError); ok && err.Value() == nil {
				break
			}
			return nil, fmt.Errorf("method runJQQuery: error in running gojq iter: %s", err)
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

func convertToGoJQCompatible(input any) any {
	switch v := input.(type) {
	case map[string]any:
		return convertMapToGoJQCompatible(v)
	case []any:
		return convertSliceToGoJQCompatible(v)
	default:
		return input
	}
}

func convertMapToGoJQCompatible(m map[string]any) map[string]any {
	compatibleMap := make(map[string]any)
	for key, value := range m {
		compatibleMap[key] = convertToGoJQCompatible(value)
	}
	return compatibleMap
}

func convertSliceToGoJQCompatible(slice []any) []any {
	compatibleSlice := make([]any, len(slice))
	for i, value := range slice {
		compatibleSlice[i] = convertToGoJQCompatible(value)
	}
	return compatibleSlice
}
