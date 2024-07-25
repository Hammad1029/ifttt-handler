package models

import (
	"context"
	"encoding/json"
	"fmt"
	"handler/scylla"
	"handler/utils"
	"io"
	"maps"
	"net/http"
	"strings"
	"time"

	jsontocql "github.com/Hammad1029/json-to-cql"
	"github.com/mitchellh/mapstructure"
	"github.com/samber/lo"
)

type Resolvable struct {
	ResolveType string                 `json:"resolveType" mapstructure:"resolveType"`
	ResolveData map[string]interface{} `json:"resolveData" mapstructure:"resolveData"`
}

func (r *Resolvable) Resolve(ctx context.Context) (interface{}, error) {
	switch r.ResolveType {
	// getters
	case "jq":
		return r.resolveJq(ctx)
	case "getReq":
		return r.getReq(ctx), nil
	case "getRes":
		return r.getRes(ctx), nil
	case "getQueryRes":
		return r.getQueryRes(ctx), nil
	case "getApiRes":
		return r.getApiRes(ctx), nil
	case "getStore":
		return r.getStore(ctx), nil
	case "const":
		return r.getConst()
	case "arithmetic":
		return r.arithmetic(ctx)
	// actions & getters both
	case "db":
		return r.resolveQuery(ctx)
	case "api":
		return r.resolveApi(ctx)
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

func resolveIfNested(original interface{}, ctx context.Context) (interface{}, error) {
	var err error

	switch o := original.(type) {
	case []interface{}:
		{
			for key, item := range o {
				if o[key], err = resolveIfNested(item, ctx); err != nil {
					return nil, fmt.Errorf("method resolveIfNested: error in resolving nested array item: %s", err)
				}
			}
			return o, nil
		}
	case map[string]interface{}:
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

func (r *Resolvable) resolveJq(ctx context.Context) (interface{}, error) {
	jqQuery, ok := r.ResolveData["query"]
	if !ok {
		return nil, fmt.Errorf("method resolveJq: jq query not found")
	}
	inputData, ok := r.ResolveData["input"]
	if !ok {
		return nil, fmt.Errorf("method resolveJq: input data not found")
	}

	inputResolved, err := resolveIfNested(inputData, ctx)
	if err != nil {
		return nil, fmt.Errorf("method resolveJq: couldn't resolve input: %s", err)
	}

	return utils.RunJQQuery(fmt.Sprint(jqQuery), inputResolved)
}

func (r *Resolvable) getReq(ctx context.Context) map[string]interface{} {
	return ctx.Value("request").(*RequestData).ReqBody
}

func (r *Resolvable) getRes(ctx context.Context) map[string]interface{} {
	return maps.Clone(ctx.Value("request").(*RequestData).Response)
}

func (r *Resolvable) getApiRes(ctx context.Context) map[string]map[string]interface{} {
	return ctx.Value("request").(*RequestData).ApiRes
}

func (r *Resolvable) getStore(ctx context.Context) map[string]interface{} {
	return ctx.Value("request").(*RequestData).Store
}

func (r *Resolvable) getQueryRes(ctx context.Context) map[string][]map[string]interface{} {
	return ctx.Value("request").(*RequestData).QueryRes
}

func (r *Resolvable) getConst() (interface{}, error) {
	if val, ok := r.ResolveData["get"]; ok {
		return val, nil
	}
	return nil, fmt.Errorf("method getConst: could not find constant in path get")
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
				var accessorResolvable Resolvable
				if accessor, ok := r.ResolveData["get"]; ok {
					if err := mapstructure.Decode(accessor, &accessorResolvable); err != nil {
						return nil, err
					}
				}

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

				return results, nil
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

func (r *Resolvable) resolveApi(ctx context.Context) (interface{}, error) {
	reqData := ctx.Value("request").(*RequestData)

	callMethod := strings.ToUpper(fmt.Sprint(r.ResolveData["method"]))
	callHeaders := r.ResolveData["headers"]
	callBody := r.ResolveData["body"]
	callURL, ok := r.ResolveData["url"]
	if !ok {
		return nil, fmt.Errorf("method resolveApi: request URL not provided")
	}

	allowedMethods := []string{"GET", "POST"}
	if !lo.Contains(allowedMethods, callMethod) {
		return nil, fmt.Errorf("method resolveApi: request method %s not found", callMethod)
	}

	var callBodyReader io.Reader
	callBodyMap, ok := callBody.(map[string]interface{})
	if ok {
		callBodyResolved, err := resolveIfNested(callBodyMap, ctx)

		if callBodyResolvedMap, ok := callBodyResolved.(map[string]interface{}); ok {
			callBodyMap = callBodyResolvedMap
		}
		if err != nil {
			return nil, fmt.Errorf("method resolveApi: could not resolve map: %s", err)
		}
		if callBodyMapStringified, err := json.Marshal(callBodyResolved); err == nil {
			callBodyReader = strings.NewReader(string(callBodyMapStringified))
		} else {
			return nil, fmt.Errorf("method resolveApi: couldn't stringify body: %s", err)
		}
	} else {
		return nil, fmt.Errorf("method resolveApi: couldn't cast body to a map[string]interface{}")
	}

	httpRequest, err := http.NewRequest(callMethod, fmt.Sprint(callURL), callBodyReader)
	if err != nil {
		return nil, fmt.Errorf("method resolveApi: could not create api request: %s", err)
	}

	var callHeadersMap map[string]string
	if err := mapstructure.Decode(callHeaders, &callHeadersMap); err != nil {
		return nil, fmt.Errorf("method resolveApi: couldn't decode headers to map[string]string")
	}
	for key, val := range callHeadersMap {
		httpRequest.Header.Add(key, val)
	}

	resp, err := http.DefaultClient.Do(httpRequest)
	if err != nil {
		return nil, fmt.Errorf("method resolveApi: api request failed: %s", err)
	}
	defer resp.Body.Close()

	respHeadersMap := make(map[string][]string)
	respBodyMap := map[string]interface{}{}

	for key, arr := range resp.Header {
		respHeadersMap[key] = arr
	}

	if err := json.NewDecoder(resp.Body).Decode(&respBodyMap); err != nil {
		return nil, fmt.Errorf("method resolveApi: could not decode response body to map: %s", err)
	}

	apiResponseStructured := map[string]interface{}{
		"request": map[string]interface{}{
			"url":     httpRequest.URL.String(),
			"method":  httpRequest.Method,
			"headers": callHeadersMap,
			"body":    callBodyMap,
		},
		"response": map[string]interface{}{
			"statusCode": resp.StatusCode,
			"status":     resp.Status,
			"headers":    respHeadersMap,
			"body":       respBodyMap,
		},
	}

	callSignature := fmt.Sprintf("%s - %s - %s", httpRequest.Method, httpRequest.URL.String(), time.Now().Format("yyyy-MM-dd HH:mm:ss"))
	reqData.ApiRes[callSignature] = apiResponseStructured
	return apiResponseStructured, nil
}
