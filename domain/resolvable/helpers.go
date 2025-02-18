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
	case accessorGetErrors:
		return &getErrors{}
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
	case accessorSetStore:
		return &setStore{}
	case accessorSetLog:
		return &setLog{}
	case accessorResponse:
		return &Response{}
	case accessorStringInterpolation:
		return &stringInterpolation{}
	case accessorEncode:
		return &encode{}
	case accessorSetCache:
		return &setCache{}
	case accessorGetCache:
		return &getCache{}
	case accessorDeleteCache:
		return &deleteCache{}
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
	case accessorDateFunc:
		return &dateFunc{}
	case accessorConditional:
		return &conditional{}
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
					return resolveMapMaybeParallel(&mapCloned, ctx, dependencies)
				}
			case reflect.Slice, reflect.Array:
				{
					oArr := []any{}
					if err := mapstructure.Decode(o, &oArr); err != nil {
						return nil, err
					}
					return resolveArrayMaybeParallel(&oArr, ctx, dependencies)
				}
			default:
				return original, nil
			}
		}
	}
}

func resolveMapMustParallel(
	m *map[string]Resolvable, ctx context.Context, dependencies map[common.IntIota]any,
) (map[string]any, error) {
	var wg sync.WaitGroup
	cancelCtx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	resolvedMap := sync.Map{}

	for key, value := range *m {
		wg.Add(1)
		go func(k string, v Resolvable) {
			defer wg.Done()
			select {
			case <-cancelCtx.Done():
				return
			default:
				if resVal, err := v.Resolve(cancelCtx, dependencies); err != nil {
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

func resolveMapMaybeParallel(
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

func ResolveArrayMust(
	resolvables *[]Resolvable, ctx context.Context, dependencies map[common.IntIota]any,
) ([]any, error) {
	rVals := make([]any, len(*resolvables))
	for idx, r := range *resolvables {
		if v, err := r.Resolve(ctx, dependencies); err != nil {
			return nil, err
		} else {
			rVals[idx] = v
		}
	}
	return rVals, nil
}

func resolveArrayMustParallel(
	resolvables *[]Resolvable, ctx context.Context, dependencies map[common.IntIota]any,
) ([]any, error) {
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	wg := sync.WaitGroup{}
	mtx := sync.Mutex{}
	rVals := make([]any, len(*resolvables))

	for idx, res := range *resolvables {
		wg.Add(1)
		go func(i int, r Resolvable) {
			defer wg.Done()
			if v, err := r.Resolve(ctx, dependencies); err != nil {
				cancel(err)
				return
			} else {
				select {
				case <-ctx.Done():
					return
				default:
					mtx.Lock()
					rVals[i] = v
					mtx.Unlock()
				}
			}
		}(idx, res)
	}
	wg.Wait()

	if err := context.Cause(ctx); err != nil {
		return nil, err
	}
	return rVals, nil
}

func ResolveArrayMaybe(
	resolvables *[]any, ctx context.Context, dependencies map[common.IntIota]any,
) ([]any, error) {
	rVals := make([]any, len(*resolvables))
	for idx, r := range *resolvables {
		if v, err := resolveMaybe(r, ctx, dependencies); err != nil {
			return nil, err
		} else {
			rVals[idx] = v
		}
	}
	return rVals, nil
}

func resolveArrayMaybeParallel(
	resolvables *[]any, ctx context.Context, dependencies map[common.IntIota]any,
) ([]any, error) {
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	wg := sync.WaitGroup{}
	mtx := sync.Mutex{}
	rVals := make([]any, len(*resolvables))

	for idx, res := range *resolvables {
		wg.Add(1)
		go func(i int, r any) {
			defer wg.Done()
			if v, err := resolveMaybe(r, ctx, dependencies); err != nil {
				cancel(err)
				return
			} else {
				select {
				case <-ctx.Done():
					return
				default:
					mtx.Lock()
					rVals[i] = v
					mtx.Unlock()
				}
			}
		}(idx, res)
	}
	wg.Wait()

	if err := context.Cause(ctx); err != nil {
		return nil, err
	}
	return rVals, nil
}
