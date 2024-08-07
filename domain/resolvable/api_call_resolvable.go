package resolvable

import (
	"context"
	"encoding/json"
	"fmt"
	"handler/common"
	"handler/domain/request_data"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/samber/lo"
)

type ApiCallResolvable struct {
	Method  string            `json:"method" mapstructure:"method"`
	Url     string            `json:"url" mapstructure:"url"`
	Headers common.JsonObject `json:"headers" mapstructure:"headers"`
	Body    common.JsonObject `json:"body" mapstructure:"body"`
}

type apiCallResponse struct {
	StatusCode int                 `json:"statusCode" mapstructure:"statusCode"`
	Status     string              `json:"status" mapstructure:"status"`
	Headers    map[string][]string `json:"headers" mapstructure:"headers"`
	Body       common.JsonObject   `json:"body" mapstructure:"body"`
}

func (a *ApiCallResolvable) Resolve(ctx context.Context, optional ...any) (any, error) {
	reqData := ctx.Value("request").(*request_data.RequestData)

	callMethod := strings.ToUpper(a.Method)
	callURL := a.Url
	callHeaders := a.Headers
	callBody := a.Body

	allowedMethods := []string{"GET", "POST"}
	if !lo.Contains(allowedMethods, callMethod) {
		return nil, fmt.Errorf("method resolveApi: request method %s not found", callMethod)
	}

	var callBodyReader io.Reader
	callBodyResolved, err := resolveIfNested(callBody, ctx)

	if callBodyResolvedMap, ok := callBodyResolved.(common.JsonObject); ok {
		callBody = callBodyResolvedMap
	}
	if err != nil {
		return nil, fmt.Errorf("method resolveApi: could not resolve map: %s", err)
	}
	if callBodyStringified, err := json.Marshal(callBodyResolved); err == nil {
		callBodyReader = strings.NewReader(string(callBodyStringified))
	} else {
		return nil, fmt.Errorf("method resolveApi: couldn't stringify body: %s", err)
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
	respBodyMap := common.JsonObject{}

	for key, arr := range resp.Header {
		respHeadersMap[key] = arr
	}

	if err := json.NewDecoder(resp.Body).Decode(&respBodyMap); err != nil {
		return nil, fmt.Errorf("method resolveApi: could not decode response body to map: %s", err)
	}

	apiResponseStructured := common.JsonObject{
		"request": a,
		"response": apiCallResponse{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Headers:    respHeadersMap,
			Body:       respBodyMap,
		},
	}

	callSignature := fmt.Sprintf("%s - %s - %s", httpRequest.Method, httpRequest.URL.String(), time.Now().Format("yyyy-MM-dd HH:mm:ss"))
	reqData.ApiRes[callSignature] = apiResponseStructured
	return apiResponseStructured, nil
}
