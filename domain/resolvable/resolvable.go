package resolvable

import (
	"context"
	"fmt"
	"ifttt/handler/common"
	"reflect"

	"github.com/mitchellh/mapstructure"
)

type Resolvable struct {
	ResolveType string         `json:"resolveType" mapstructure:"resolveType"`
	ResolveData map[string]any `json:"resolveData" mapstructure:"resolveData"`
}

const (
	AccessorJqResolvable                  = "jq"
	AccessorGetRequestResolvable          = "getReq"
	AccessorGetResponseResolvable         = "getRes"
	AccessorGetQueryResultsResolvable     = "getQueryRes"
	AccessorGetApiResultsResolvable       = "getApiRes"
	AccessorGetStoreResolvable            = "getStore"
	AccessorGetConstResolvable            = "const"
	AccessorArithmetic                    = "arithmetic"
	AccessorQueryResolvable               = "query"
	AccessorApiCallResolvable             = "api"
	AccessorSetResResolvable              = "setRes"
	AccessorSetStoreResolvable            = "setStore"
	AccessorSetLogResolvable              = "log"
	AccessorResponseResolvable            = "sendRes"
	AccessorPreConfigResolvable           = "getPreConfig"
	AccessorStringInterpolationResolvable = "stringInterpolation"
	AccessorEncodeResolvable              = "encode"
	AccessorSetCacheResolvable            = "setCache"
	AccessorGetCacheResolvable            = "getCache"
	AccessorUUIDResolvable                = "uuid"
	AccessorHeadersResolvable             = "headers"
)

type resolvableInterface interface {
	Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error)
}

func (r *Resolvable) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	var genericResolvable resolvableInterface
	var err error

	if resolvableStruct := resolvableFactory(r.ResolveType); resolvableStruct != nil {
		if err = mapstructure.Decode(r.ResolveData, resolvableStruct); err != nil {
			return nil, fmt.Errorf("method Resolve: could not decode map to resolvable: %s", err)
		}
		genericResolvable = resolvableStruct
	} else {
		return nil, fmt.Errorf("method Resolve: resolveType %s not found", r.ResolveType)
	}

	return genericResolvable.Resolve(ctx, dependencies)
}

func resolvableFactory(rType string) resolvableInterface {
	switch rType {
	case AccessorJqResolvable:
		return &jqResolvable{}
	case AccessorGetRequestResolvable:
		return &getRequestResolvable{}
	case AccessorGetResponseResolvable:
		return &getResponseResolvable{}
	case AccessorGetQueryResultsResolvable:
		return &getQueryResultsResolvable{}
	case AccessorGetApiResultsResolvable:
		return &getApiResultsResolvable{}
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
		return &preConfigResolvable{}
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
		return &headersResolvable{}
	default:
		return nil
	}
}

func resolveIfNested(original any, ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	var err error

	switch o := original.(type) {
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

					var mapCloned map[string]any
					if err := mapstructure.Decode(o, &mapCloned); err != nil {
						return nil, err
					}
					for key, val := range mapCloned {
						if mapCloned[key], err = resolveIfNested(val, ctx, dependencies); err != nil {
							return nil, err
						}
					}
					return mapCloned, nil
				}
			case reflect.Slice, reflect.Array:
				{
					var oArr []any
					if err := mapstructure.Decode(o, &oArr); err != nil {
						return nil, err
					}
					for key, item := range oArr {
						if oArr[key], err = resolveIfNested(item, ctx, dependencies); err != nil {
							return nil, err
						}
					}
					return oArr, nil

				}
			default:
				return original, nil
			}
		}
	}
}
