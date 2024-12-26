package resolvable

import (
	"context"
	"fmt"
	"ifttt/handler/common"

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
	AccessorDbDumpResolvable              = "dbDump"
	AccessorCastResolvable                = "cast"
	AccessorOrmResolvable                 = "orm"
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
