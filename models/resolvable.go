package models

import (
	"context"
	"errors"
	"fmt"
	"handler/scylla"
	"strings"

	jsontocql "github.com/Hammad1029/json-to-cql"
	"github.com/mitchellh/mapstructure"
)

type Resolvable struct {
	ResolveType string                 `json:"resolveType" mapstructure:"resolveType"`
	ResolveData map[string]interface{} `json:"resolveData" mapstructure:"resolveData"`
}

func (r *Resolvable) Resolve(ctx context.Context) (interface{}, error) {
	switch r.ResolveType {
	case "req":
		return r.resolveReq(ctx)
	case "db":
		return r.resolveQuery(ctx)
	case "const":
		return fmt.Sprint(r.ResolveData["get"]), nil
	case "arithmetic":
		return r.arithmetic(ctx)
	// actions
	case "setRes":
		return nil, r.setRes(ctx)
	case "store":
		return nil, r.store(ctx)
	case "log":
		return nil, r.saveUserLog(ctx)
	default:
		return nil, fmt.Errorf("resolveType %s not found", r.ResolveType)
	}
}

func resolveIfNested(value interface{}, ctx context.Context) (interface{}, error) {
	if valueMap, ok := value.(map[string]interface{}); ok {
		if resolvableType, ok := valueMap["resolveType"]; ok {
			if resolvableData, ok := valueMap["resolveData"]; ok {
				if castedType, ok := resolvableType.(string); ok {
					if castedData, ok := resolvableData.(map[string]interface{}); ok {
						nestedResolvable := Resolvable{ResolveType: castedType, ResolveData: castedData}
						return nestedResolvable.Resolve(ctx)
					}
				}
			}
		}
	}
	return value, nil
}

func (r *Resolvable) resolveReq(ctx context.Context) (interface{}, error) {
	reqData := ctx.Value("request").(*RequestData).ReqBody
	reqPath, err := resolveIfNested(r.ResolveData["get"], ctx)
	if err != nil {
		return nil, err
	}
	keys := strings.Split(fmt.Sprint(reqPath), ".")

	current := reqData

	for _, key := range keys {
		val, ok := current[key]
		if !ok {
			return nil, nil
		}

		if nestedMap, ok := val.(map[string]interface{}); ok {
			current = nestedMap
		} else {
			return val, nil
		}
	}

	return nil, errors.New("method resolveReq: invalid accessor")
}

func (r *Resolvable) resolveQuery(ctx context.Context) (interface{}, error) {
	reqData := ctx.Value("request").(*RequestData)
	queries := ctx.Value("queries").(map[string]QueryUDT)
	queryHash := fmt.Sprint(r.ResolveData["query"])
	if currQuery, ok := queries[queryHash]; ok {
		var queryParameters []interface{}
		for _, param := range currQuery.Resolvables {
			if p, err := param.Resolve(ctx); err != nil {
				return nil, err
			} else {
				queryParameters = append(queryParameters, fmt.Sprint(p))
			}
		}

		switch currQuery.Type {
		case jsontocql.Select:
			{
				queryRes := reqData.QueryRes
				var results []map[string]interface{}
				if oldRes, queryRan := queryRes[queryHash]; queryRan {
					results = oldRes
				} else {
					if newRes, err := scylla.RunSelect(currQuery.QueryString, queryParameters, 1); err != nil {
						return nil, err
					} else {
						reqData.QueryRes[queryHash] = newRes
						results = newRes
					}
				}
				if accessor, ok := r.ResolveData["get"]; ok {
					if len(results) == 0 {
						return nil, nil
					}
					if val, ok := results[0][fmt.Sprint(accessor)]; ok {
						return val, nil

					}
					return nil, errors.New("method resolveQuery: accessor not found in query result")
				}
				return nil, errors.New("method resolveQuery: query accessor not found")
			}
		default:
			{
				if err := scylla.RunQuery(currQuery.QueryString, queryParameters); err != nil {
					return nil,
						fmt.Errorf("method resolveQuery: error running non select query | %s", err)
				}
				return nil, nil
			}
		}
	} else {
		return nil, fmt.Errorf("method resolveQuery: query hash %s not found", queryHash)
	}
}

func (r *Resolvable) setRes(ctx context.Context) error {
	if reqData, ok := ctx.Value("request").(*RequestData); ok {
		responseData := reqData.Response
		for key, value := range r.ResolveData {
			resVal, err := resolveIfNested(value, ctx)
			if err != nil {
				return err
			}
			responseData[key] = resVal
		}
		return nil
	}
	return errors.New("method setRes: setRes resolveType assertion failed")
}

func (r *Resolvable) store(ctx context.Context) error {
	if reqData, ok := ctx.Value("request").(*RequestData); ok {
		store := reqData.Store
		for key, value := range r.ResolveData {
			resVal, err := resolveIfNested(value, ctx)
			if err != nil {
				return err
			}
			store[key] = resVal
		}
		return nil
	}
	return errors.New("method store: setRes resolveType assertion failed")

}

func (r *Resolvable) saveUserLog(ctx context.Context) error {
	logData := r.ResolveData["logData"]
	logType := r.ResolveData["logType"]
	if l, ok := ctx.Value("log").(*LogModel); ok {
		logTypeResolved, err := resolveIfNested(logType, ctx)
		if err != nil {
			return err
		}
		logDataResolved, err := resolveIfNested(logData, ctx)
		if err != nil {
			return err
		}
		l.AddExecLog("user", fmt.Sprint(logTypeResolved), fmt.Sprint(logDataResolved))
		return nil
	}
	return errors.New("method saveUserLog: could not type cast log model")
}

func (r *Resolvable) arithmetic(ctx context.Context) (interface{}, error) {
	var arithmetic Arithmetic
	mapstructure.Decode(r.ResolveData, &arithmetic)
	return arithmetic.Arithmetic(ctx)
}
