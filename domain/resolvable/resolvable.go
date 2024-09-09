package resolvable

import (
	"context"
	"fmt"
	"reflect"

	"github.com/mitchellh/mapstructure"
)

const (
	AccessorRuleResolvable            = "rule"
	AccessorJqResolvable              = "jq"
	AccessorGetRequestResolvable      = "getReq"
	AccessorGetResponseResolvable     = "getRes"
	AccessorGetQueryResultsResolvable = "getQueryRes"
	AccessorGetApiResultsResolvable   = "getApiRes"
	AccessorGetStoreResolvable        = "getStore"
	AccessorGetConstResolvable        = "const"
	AccessorArithmetic                = "arithmetic"
	AccessorQueryResolvable           = "query"
	AccessorApiCallResolvable         = "api"
	AccessorSetResResolvable          = "setRes"
	AccessorSetStoreResolvable        = "setStore"
	AccessorSetLogResolvable          = "log"
	AccessorResponseResolvable        = "sendRes"
	AccessorPreConfigResolvable       = "getPreConfig"
)

type ResolvableInterface interface {
	Resolve(ctx context.Context, dependencies map[string]any) (any, error)
}

type Resolvable struct {
	ResolveType string         `json:"resolveType" mapstructure:"resolveType"`
	ResolveData map[string]any `json:"resolveData" mapstructure:"resolveData"`
}

func (r *Resolvable) Resolve(ctx context.Context, dependencies map[string]any) (any, error) {
	var genericResolvable ResolvableInterface
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

func resolvableFactory(rType string) ResolvableInterface {
	switch rType {
	case AccessorJqResolvable:
		return &JqResolvable{}
	case AccessorGetRequestResolvable:
		return &GetRequestResolvable{}
	case AccessorGetResponseResolvable:
		return &GetResponseResolvable{}
	case AccessorGetQueryResultsResolvable:
		return &GetQueryResultsResolvable{}
	case AccessorGetApiResultsResolvable:
		return &GetApiResultsResolvable{}
	case AccessorGetStoreResolvable:
		return &GetStoreResolvable{}
	case AccessorGetConstResolvable:
		return &GetConstResolvable{}
	case AccessorArithmetic:
		return &Arithmetic{}
	case AccessorQueryResolvable:
		return &QueryResolvable{}
	case AccessorApiCallResolvable:
		return &ApiCallResolvable{}
	case AccessorSetResResolvable:
		return &SetResResolvable{}
	case AccessorSetStoreResolvable:
		return &SetStoreResolvable{}
	case AccessorSetLogResolvable:
		return &SetLogResolvable{}
	case AccessorResponseResolvable:
		return &ResponseResolvable{}
	case AccessorRuleResolvable:
		return &CallRuleResolvable{}
	case AccessorPreConfigResolvable:
		return &PreConfigResolvable{}
	default:
		return nil
	}
}

func resolveIfNested(original any, ctx context.Context, dependencies map[string]any) (any, error) {
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
						return nil, fmt.Errorf("method resolveIfNested: error in cloning map: %s", err)
					}
					for key, val := range mapCloned {
						if mapCloned[key], err = resolveIfNested(val, ctx, dependencies); err != nil {
							return nil, fmt.Errorf("method resolveIfNested: error in resolving nested map: %s", err)
						}
					}
					return mapCloned, nil
				}
			case reflect.Slice, reflect.Array:
				{
					var oArr []any
					if err := mapstructure.Decode(o, &oArr); err != nil {
						return nil, fmt.Errorf("method resolveIfNested: error in decoding to []any: %s", err)
					}
					for key, item := range oArr {
						if oArr[key], err = resolveIfNested(item, ctx, dependencies); err != nil {
							return nil, fmt.Errorf("method resolveIfNested: error in resolving nested array item: %s", err)
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
