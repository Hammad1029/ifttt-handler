package models

import (
	"context"
	"fmt"
	"handler/common"
	"maps"

	"github.com/mitchellh/mapstructure"
)

type ResolvableInterface interface {
	Resolve(ctx context.Context) (any, error)
}

type Resolvable struct {
	ResolveType string            `json:"resolveType" mapstructure:"resolveType"`
	ResolveData common.JsonObject `json:"resolveData" mapstructure:"resolveData"`
}

func (r *Resolvable) Resolve(ctx context.Context) (any, error) {
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

	return genericResolvable.Resolve(ctx)
}

func resolvableFactory(rType string) ResolvableInterface {
	switch rType {
	case "jq":
		return &JqResolvable{}
	case "setRes":
		return &SetResResolvable{}
	case "getRes":
		return &ResponseResolvable{}
	case "getReq":
		return &RequestResolvable{}
	default:
		return nil
	}
}

func resolveIfNested(original any, ctx context.Context) (any, error) {
	var err error

	switch o := original.(type) {
	case []any:
		{
			for key, item := range o {
				if o[key], err = resolveIfNested(item, ctx); err != nil {
					return nil, fmt.Errorf("method resolveIfNested: error in resolving nested array item: %s", err)
				}
			}
			return o, nil
		}
	case map[string]any:
		{
			var nestedResolvable Resolvable
			err = mapstructure.Decode(o, &nestedResolvable)
			if err == nil && nestedResolvable.ResolveType != "" && nestedResolvable.ResolveData != nil {
				return nestedResolvable.Resolve(ctx)
			}

			mapCloned := maps.Clone(o)
			for key, val := range mapCloned {
				if mapCloned[key], err = resolveIfNested(val, ctx); err != nil {
					return nil, fmt.Errorf("method resolveIfNested: error in resolving nested map: %s", err)
				}
			}
			return mapCloned, nil
		}
	default:
		{
			return original, nil
		}
	}
}
