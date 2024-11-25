package resolvable

import (
	"context"
	"fmt"
	"ifttt/handler/common"
	"ifttt/handler/domain/audit_log"
	"reflect"
	"sync"

	"github.com/fatih/structs"
	"github.com/itchyny/gojq"
)

type jqResolvable struct {
	Query Resolvable `json:"query" mapstructure:"query"`
	Input any        `json:"input" mapstructure:"input"`
}

func (j *jqResolvable) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	queryResolved, err := j.Query.Resolve(ctx, dependencies)
	if err != nil {
		audit_log.AddExecLog(common.LogUser, common.LogError, err.Error(), ctx)
		return nil, nil
	}
	inputResolved, err := resolveIfNested(j.Input, ctx, dependencies)
	if err != nil {
		return nil, err
	}

	return runJQQuery(fmt.Sprint(queryResolved), inputResolved, ctx)
}

func runJQQuery(queryString string, input any, ctx context.Context) (any, error) {
	jqInput := convertToGoJQCompatible(input)

	if queryString[0] != '.' {
		queryString = "." + queryString
	}
	query, err := gojq.Parse(queryString)
	if err != nil {
		return nil, fmt.Errorf("error parsing jq query: %s error: %s", queryString, err)
	}

	var resultVals []any
	resultIter := query.Run(jqInput)

	for {
		v, ok := resultIter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			if haltErr, ok := err.(*gojq.HaltError); ok && haltErr.Value() == nil {
				break
			}
			audit_log.AddExecLog(common.LogUser, common.LogError, err, ctx)
			return nil, nil
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
	if syncMap, ok := input.(*sync.Map); ok {
		unsyncedValue := common.UnSyncMap(syncMap)
		return convertMapToGoJQCompatible(reflect.ValueOf(unsyncedValue))
	}

	v := reflect.Indirect(reflect.ValueOf(input))
	switch v.Kind() {
	case reflect.Map:
		return convertMapToGoJQCompatible(v)
	case reflect.Slice:
		return convertSliceToGoJQCompatible(v)
	case reflect.Struct:
		return convertMapToGoJQCompatible(reflect.ValueOf(structs.Map(input)))
	default:
		return input
	}
}

func convertMapToGoJQCompatible(v reflect.Value) map[string]any {
	compatibleMap := make(map[string]any)
	for _, key := range v.MapKeys() {
		value := v.MapIndex(key).Interface()
		if keyStr, ok := key.Interface().(string); ok {
			compatibleMap[keyStr] = convertToGoJQCompatible(value)
		}
	}
	return compatibleMap
}

func convertSliceToGoJQCompatible(v reflect.Value) []any {
	compatibleSlice := make([]any, v.Len())
	for i := 0; i < v.Len(); i++ {
		value := v.Index(i).Interface()
		compatibleSlice[i] = convertToGoJQCompatible(value)
	}
	return compatibleSlice
}
