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
	accessorGetRequest          = "getReq"
	accessorGetResponse         = "getRes"
	accessorGetStore            = "getStore"
	accessorGetConst            = "const"
	accessorArithmetic          = "arithmetic"
	accessorQuery               = "query"
	accessorApiCall             = "api"
	accessorSetRes              = "setRes"
	accessorSetStore            = "setStore"
	accessorSetLog              = "log"
	accessorResponse            = "sendRes"
	accessorPreConfig           = "getPreConfig"
	accessorStringInterpolation = "stringInterpolation"
	accessorEncode              = "encode"
	accessorSetCache            = "setCache"
	accessorGetCache            = "getCache"
	accessorUUID                = "uuid"
	accessorHeaders             = "headers"
	accessorCast                = "cast"
	accessorOrm                 = "orm"
	accessorForEach             = "forEach"
	accessorGetIter             = "getIter"
	accessorDateInput           = "dateInput"
	accessorDateManipulator     = "dateManipulator"
	accessorDateFunc            = "dateFunc"
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
