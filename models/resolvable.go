package models

import (
	"context"
	"fmt"
	"handler/scylla"

	jsontocql "github.com/Hammad1029/json-to-cql"
	"github.com/PaesslerAG/jsonpath"
	"github.com/mitchellh/mapstructure"
	"github.com/samber/lo"
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
		return r.ResolveData["get"], nil
	case "arithmetic":
		return r.arithmetic(ctx)
	case "getStore":
		return r.getStore(ctx)
	// actions
	case "setRes":
		return nil, r.setRes(ctx)
	case "setStore":
		return nil, r.setStore(ctx)
	case "log":
		return nil, r.saveUserLog(ctx)
	default:
		return nil, fmt.Errorf("resolveType %s not found", r.ResolveType)
	}
}

func resolveIfNested(value interface{}, ctx context.Context) (interface{}, error) {
	var nestedResolvable Resolvable
	if err := mapstructure.Decode(value, &nestedResolvable); err != nil {
		return nil, err
	}
	if nestedResolvable.ResolveType != "" && nestedResolvable.ResolveData != nil {
		return nestedResolvable.Resolve(ctx)
	}
	return value, nil
}

func (r *Resolvable) resolveReq(ctx context.Context) (interface{}, error) {
	reqData := ctx.Value("request").(*RequestData).ReqBody
	reqPath, err := resolveIfNested(r.ResolveData["get"], ctx)
	if err != nil {
		return nil, err
	}

	if val, err := jsonpath.Get(fmt.Sprint(reqPath), reqData); err != nil {
		return nil, fmt.Errorf("method resolveReq: %s", err.Error())
	} else {
		return val, nil
	}
}

func (r *Resolvable) resolveQuery(ctx context.Context) (interface{}, error) {
	reqData := ctx.Value("request").(*RequestData)
	queries := ctx.Value("queries").(map[string]jsontocql.ParameterizedQuery)
	queryHash := fmt.Sprint(r.ResolveData["query"])
	if currQuery, ok := queries[queryHash]; ok {
		var queryParameters []interface{}
		var localResolvable Resolvable
		for _, param := range currQuery.Resolvables {
			localResolvable = Resolvable{
				ResolveType: param.ResolveType,
				ResolveData: param.ResolveData,
			}
			if p, err := localResolvable.Resolve(ctx); err != nil {
				return nil, fmt.Errorf("method resolveQuery: could not resolve query parameters: %s", err)
			} else {
				queryParameters = append(queryParameters, p)
			}
		}

		switch currQuery.Type {
		case jsontocql.Select:
			{
				var results []map[string]interface{}
				queryRes := reqData.QueryRes

				callAgain, callAgainOk := r.ResolveData["callAgain"]
				oldRes, queryRan := queryRes[queryHash]

				if (!callAgainOk || callAgain != true) && queryRan {
					results = oldRes
				} else {
					if newRes, err := scylla.RunSelect(currQuery, queryParameters); err != nil {
						return nil, fmt.Errorf("method resolveQuery: could not run query: %s", err.Error())
					} else {
						reqData.QueryRes[queryHash] = newRes
						results = newRes
					}
				}

				if len(results) == 0 {
					return nil, nil
				}

				var accessorResolvable Resolvable
				if accessor, ok := r.ResolveData["get"]; ok {
					if err := mapstructure.Decode(accessor, &accessorResolvable); err != nil {
						return nil, err
					}
				}
				resolved, err := accessorResolvable.Resolve(ctx)
				if err != nil {
					return nil, fmt.Errorf("method resolveQuery: could not resolve accessor | %s", err.Error())
				}
				accessorResolved := fmt.Sprint(resolved)
				if accessorResolved == "*" {
					return results, nil
				} else {
					interfaceResults := lo.Map(results, func(m map[string]interface{}, _ int) interface{} {
						return m
					})
					if filteredResults, err := jsonpath.Get(accessorResolved, interfaceResults); err != nil {
						return nil, fmt.Errorf("method resolveQuery: could not access accessor | %s", err.Error())
					} else {
						return filteredResults, nil
					}

				}
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
	return fmt.Errorf("method setRes: setRes resolveType assertion failed")
}

func (r *Resolvable) setStore(ctx context.Context) error {
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
	return fmt.Errorf("method store: setRes resolveType assertion failed")

}

func (r *Resolvable) saveUserLog(ctx context.Context) error {
	logData := r.ResolveData["logData"]
	logType := r.ResolveData["logType"]
	if l, ok := ctx.Value("log").(*LogData); ok {
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
	return fmt.Errorf("method saveUserLog: could not type cast log model")
}

func (r *Resolvable) arithmetic(ctx context.Context) (interface{}, error) {
	var arithmetic Arithmetic
	mapstructure.Decode(r.ResolveData, &arithmetic)
	return arithmetic.Arithmetic(ctx)
}

func (r *Resolvable) getStore(ctx context.Context) (interface{}, error) {
	store := ctx.Value("request").(*RequestData).Store
	storePath, err := resolveIfNested(r.ResolveData["get"], ctx)
	if err != nil {
		return nil, err
	}

	if val, err := jsonpath.Get(fmt.Sprint(storePath), store); err != nil {
		return nil, fmt.Errorf("method getStore: %s", err.Error())
	} else {
		return val, nil
	}
}
