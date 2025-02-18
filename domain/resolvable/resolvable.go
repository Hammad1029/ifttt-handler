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
	accessorJq                  = "jq"
	accessorGetErrors           = "getErrors"
	accessorGetStore            = "getStore"
	accessorGetConst            = "const"
	accessorArithmetic          = "arithmetic"
	accessorQuery               = "query"
	accessorApiCall             = "api"
	accessorSetStore            = "setStore"
	accessorSetLog              = "log"
	accessorResponse            = "response"
	accessorStringInterpolation = "stringInterpolation"
	accessorEncode              = "encode"
	accessorSetCache            = "setCache"
	accessorGetCache            = "getCache"
	accessorDeleteCache         = "deleteCache"
	accessorUUID                = "uuid"
	accessorHeaders             = "headers"
	accessorCast                = "cast"
	accessorOrm                 = "orm"
	accessorForEach             = "forEach"
	accessorGetIter             = "getIter"
	accessorDateFunc            = "dateFunc"
	accessorConditional         = "conditional"
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
