package models

import (
	"context"
	"errors"
	"fmt"
	"handler/scylla"
	"strings"

	jsontocql "github.com/Hammad1029/json-to-cql"
)

type ResolvableUDT struct {
	Type string            `cql:"type" json:"type"`
	Data map[string]string `cql:"data" json:"data"`
}

func (r *ResolvableUDT) Resolve(ctx context.Context) (string, error) {
	switch r.Type {
	case "req":
		return r.resolveReq(ctx)
	case "db":
		return r.resolveQuery(ctx)
	case "const":
		return fmt.Sprint(r.Data["get"]), nil
	case "setRes":
		return "", r.setRes(ctx)
	case "store":
		return "", r.store(ctx)
	default:
		return "", fmt.Errorf("resolvable type %s not found", r.Type)
	}
}

func (r *ResolvableUDT) resolveReq(ctx context.Context) (string, error) {
	reqData := ctx.Value("request").(RequestData).ReqBody
	keys := strings.Split(fmt.Sprint(r.Data["get"]), ".")

	current := reqData
	for _, key := range keys {
		val, ok := current[key]
		if !ok {
			return "", nil
		}

		if nestedMap, ok := val.(map[string]interface{}); ok {
			current = nestedMap
		} else if stringVal, ok := val.(string); ok {
			return stringVal, nil
		} else {
			return "", errors.New("cannot typecast property")
		}
	}

	return "", errors.New("invalid accessor")
}

func (r *ResolvableUDT) resolveQuery(ctx context.Context) (string, error) {
	reqData := ctx.Value("request").(RequestData)
	queries := ctx.Value("queries").(map[string]QueryUDT)
	queryHash := fmt.Sprint(r.Data["query"])
	if currQuery, ok := queries[queryHash]; ok {
		var queryParameters []string
		for _, param := range currQuery.Resolvables {
			if p, err := param.Resolve(ctx); err != nil {
				return "", err
			} else {
				queryParameters = append(queryParameters, p)
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
						return "", err
					} else {
						reqData.QueryRes[queryHash] = newRes
						results = newRes
					}
				}
				if accessor, ok := r.Data["get"]; ok {
					if len(results) == 0 {
						return "", nil
					}
					if val, ok := results[0][fmt.Sprint(accessor)]; ok {
						return fmt.Sprint(val), nil
					}
					return "", errors.New("accessor not found in query result")
				}
				return "", errors.New("query accessor not found")
			}
		default:
			{
				return "", scylla.RunQuery(currQuery.QueryString, queryParameters)
			}
		}
	} else {
		return "", fmt.Errorf("query hash %s not found", queryHash)
	}
}

func (r *ResolvableUDT) setRes(ctx context.Context) error {
	if data, ok := ctx.Value("request").(RequestData); ok {
		responseData := data.Response
		for key, value := range r.Data {
			responseData[key] = value
		}
		return nil
	}
	return errors.New("set res type assertion failed")
}

func (r *ResolvableUDT) store(ctx context.Context) error {
	if data, ok := ctx.Value("request").(RequestData); ok {
		store := data.Store
		for key, value := range r.Data {
			store[key] = value
		}
		return nil
	}
	return errors.New("set res type assertion failed")
}
