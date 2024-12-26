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
	case AccessorJqResolvable:
		return &jqResolvable{}
	case AccessorGetRequestResolvable:
		return &getRequestResolvable{}
	case AccessorGetResponseResolvable:
		return &getResponseResolvable{}
	case AccessorGetStoreResolvable:
		return &getStoreResolvable{}
	case AccessorGetConstResolvable:
		return &getConstResolvable{}
	case AccessorArithmetic:
		return &arithmetic{}
	case AccessorQueryResolvable:
		return &queryResolvable{}
	case AccessorApiCallResolvable:
		return &apiCallResolvable{}
	case AccessorSetResResolvable:
		return &setResResolvable{}
	case AccessorSetStoreResolvable:
		return &setStoreResolvable{}
	case AccessorSetLogResolvable:
		return &setLogResolvable{}
	case AccessorResponseResolvable:
		return &ResponseResolvable{}
	case AccessorPreConfigResolvable:
		return &getPreConfigResolvable{}
	case AccessorStringInterpolationResolvable:
		return &stringInterpolationResolvable{}
	case AccessorEncodeResolvable:
		return &encodeResolvable{}
	case AccessorSetCacheResolvable:
		return &setCacheResolvable{}
	case AccessorGetCacheResolvable:
		return &getCacheResolvable{}
	case AccessorUUIDResolvable:
		return &uuidResolvable{}
	case AccessorHeadersResolvable:
		return &getHeadersResolvable{}
	case AccessorDbDumpResolvable:
		return &dbDumpResolvable{}
	case AccessorCastResolvable:
		return &castResolvable{}
	case AccessorOrmResolvable:
		return &ormResolvable{}
	default:
		return nil
	}
}

func resolveIfNested(original any, ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
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
					return resolveMap(&mapCloned, ctx, dependencies)
				}
			case reflect.Slice, reflect.Array:
				{
					oArr := []any{}
					if err := mapstructure.Decode(o, &oArr); err != nil {
						return nil, err
					}
					return resolveSlice(&oArr, ctx, dependencies)
				}
			default:
				return original, nil
			}
		}
	}
}

func resolveMap(
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
				if resVal, err := resolveIfNested(v, cancelCtx, dependencies); err != nil {
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

func resolveSlice(s *[]any, ctx context.Context, dependencies map[common.IntIota]any,
) ([]any, error) {
	var wg sync.WaitGroup
	cancelCtx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	resolvedSlice := []any{}
	mtx := sync.Mutex{}

	for _, value := range *s {
		wg.Add(1)
		go func(v any) {
			defer wg.Done()
			select {
			case <-cancelCtx.Done():
				return
			default:
				if resVal, err := resolveIfNested(v, cancelCtx, dependencies); err != nil {
					cancel(err)
				} else {
					mtx.Lock()
					resolvedSlice = append(resolvedSlice, resVal)
					mtx.Unlock()
				}
				return
			}
		}(value)
	}

	wg.Wait()
	if err := context.Cause(cancelCtx); err != nil {
		return nil, fmt.Errorf("could not perform concurrent resolve on slice: %s", err)
	}
	return resolvedSlice, nil
}
