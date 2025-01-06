package resolvable

import (
	"context"
	"fmt"
	"ifttt/handler/common"
	"reflect"
	"sync"

	"github.com/mitchellh/mapstructure"
)

func resolvableFactory(rType string) resolvableInterface {
	switch rType {
	case accessorJq:
		return &jq{}
	case accessorGetRequest:
		return &getRequest{}
	case accessorGetResponse:
		return &getResponse{}
	case accessorGetStore:
		return &getStore{}
	case accessorGetConst:
		return &getConst{}
	case accessorArithmetic:
		return &arithmetic{}
	case accessorQuery:
		return &query{}
	case accessorApiCall:
		return &apiCall{}
	case accessorSetRes:
		return &setRes{}
	case accessorSetStore:
		return &setStore{}
	case accessorSetLog:
		return &setLog{}
	case accessorResponse:
		return &Response{}
	case accessorPreConfig:
		return &getPreConfig{}
	case accessorStringInterpolation:
		return &stringInterpolation{}
	case accessorEncode:
		return &encode{}
	case accessorSetCache:
		return &setCache{}
	case accessorGetCache:
		return &getCache{}
	case accessorUUID:
		return &generateUUID{}
	case accessorHeaders:
		return &getHeaders{}
	case accessorCast:
		return &cast{}
	case accessorOrm:
		return &orm{}
	case accessorForEach:
		return &forEach{}
	case accessorGetIter:
		return &getIter{}
	case accessorDateInput:
		return &dateInput{}
	case accessorDateManipulator:
		return &dateManipulator{}
	case accessorDateFunc:
		return &dateFunc{}
	default:
		return nil
	}
}

func resolveMaybe(original any, ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	var err error

	switch o := original.(type) {
	case nil:
		return nil, nil
	case Resolvable:
		return o.Resolve(ctx, dependencies)
	default:
		{
			switch reflect.TypeOf(o).Kind() {
			case reflect.Map:
				{
					var nestedResolvable Resolvable
					err = mapstructure.Decode(o, &nestedResolvable)
					if err == nil && nestedResolvable.ResolveType != "" && nestedResolvable.ResolveData != nil {
						return nestedResolvable.Resolve(ctx, dependencies)
					}

					mapCloned := map[string]any{}
					if err := mapstructure.Decode(o, &mapCloned); err != nil {
						return nil, err
					}
					return resolveMapMaybe(&mapCloned, ctx, dependencies)
				}
			case reflect.Slice, reflect.Array:
				{
					oArr := []any{}
					if err := mapstructure.Decode(o, &oArr); err != nil {
						return nil, err
					}
					return resolveSliceMaybe(&oArr, ctx, dependencies)
				}
			default:
				return original, nil
			}
		}
	}
}

func resolveMapMaybe(
	m *map[string]any, ctx context.Context, dependencies map[common.IntIota]any,
) (map[string]any, error) {
	var wg sync.WaitGroup
	cancelCtx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	resolvedMap := sync.Map{}

	for key, value := range *m {
		wg.Add(1)
		go func(k string, v any) {
			defer wg.Done()
			select {
			case <-cancelCtx.Done():
				return
			default:
				if resVal, err := resolveMaybe(v, cancelCtx, dependencies); err != nil {
					cancel(err)
				} else {
					resolvedMap.Store(k, resVal)
				}
				return
			}
		}(key, value)
	}

	wg.Wait()
	if err := context.Cause(cancelCtx); err != nil {
		return nil, fmt.Errorf("could not perform concurrent resolve on map: %s", err)
	}
	return common.SyncMapUnsync(&resolvedMap), nil
}

func resolveSliceMaybe(s *[]any, ctx context.Context, dependencies map[common.IntIota]any,
) ([]any, error) {
	var wg sync.WaitGroup
	cancelCtx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	resolvedSlice := make([]any, len(*s))
	mtx := sync.Mutex{}

	for idx, value := range *s {
		wg.Add(1)
		go func(i int, v any) {
			defer wg.Done()
			select {
			case <-cancelCtx.Done():
				return
			default:
				if resVal, err := resolveMaybe(v, cancelCtx, dependencies); err != nil {
					cancel(err)
				} else {
					mtx.Lock()
					resolvedSlice[i] = resVal
					mtx.Unlock()
				}
				return
			}
		}(idx, value)
	}

	wg.Wait()
	if err := context.Cause(cancelCtx); err != nil {
		return nil, fmt.Errorf("could not perform concurrent resolve on slice: %s", err)
	}
	return resolvedSlice, nil
}

func ResolveArrayMust(
	resolvables *[]Resolvable, ctx context.Context, dependencies map[common.IntIota]any,
) error {
	for _, r := range *resolvables {
		common.LogWithTracer(common.LogSystem, fmt.Sprintf("resolving %s", r.ResolveType), r, false, ctx)
		if _, err := r.Resolve(ctx, dependencies); err != nil {
			return err
		}
	}
	return nil
}
