package resolvable

import (
	"context"
	"encoding/json"
	"fmt"
	"ifttt/handler/domain/request_data"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/samber/lo"
)

type ApiCallResolvable struct {
	Method  string         `json:"method" mapstructure:"method"`
	Url     string         `json:"url" mapstructure:"url"`
	Headers map[string]any `json:"headers" mapstructure:"headers"`
	Body    map[string]any `json:"body" mapstructure:"body"`
	Start   time.Time      `json:"start" mapstructure:"start"`
}

type apiCallResponse struct {
	StatusCode int                 `json:"statusCode" mapstructure:"statusCode"`
	Status     string              `json:"status" mapstructure:"status"`
	Headers    map[string][]string `json:"headers" mapstructure:"headers"`
	Body       map[string]any      `json:"body" mapstructure:"body"`
	End        time.Time           `json:"end" mapstructure:"end"`
	TimeTaken  int64               `json:"timeTaken" mapstructure:"timeTaken"`
}

func (a *ApiCallResolvable) Resolve(ctx context.Context, dependencies map[string]any) (any, error) {
	var response apiCallResponse
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
	if callBodyResolved, err := resolveIfNested(callBody, ctx, dependencies); err != nil {
		return nil, fmt.Errorf("method resolveApi: could not resolve map: %s", err)
	} else {
		if err := mapstructure.Decode(callBodyResolved, &a.Body); err != nil {
			return nil, fmt.Errorf("method resolveApi: could not decode resolved request body to a.Body: %s", err)
		}
	}

	if callBodyStringified, err := json.Marshal(a.Body); err == nil {
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

	a.Start = time.Now()
	resp, err := http.DefaultClient.Do(httpRequest)
	if err != nil {
		return nil, fmt.Errorf("method resolveApi: api request failed: %s", err)
	}
	response.End = time.Now()
	response.TimeTaken = response.End.Sub(a.Start).Milliseconds()
	defer resp.Body.Close()

	respHeadersMap := make(map[string][]string)
	respBodyMap := map[string]any{}

	for key, arr := range resp.Header {
		respHeadersMap[key] = arr
	}

	if err := json.NewDecoder(resp.Body).Decode(&respBodyMap); err != nil {
		return nil, fmt.Errorf("method resolveApi: could not decode response body to map: %s", err)
	}

	response.StatusCode = resp.StatusCode
	response.Status = resp.Status
	response.Headers = respHeadersMap
	response.Body = respBodyMap

	apiResponseStructured := map[string]any{
		"request":  a,
		"response": response,
	}

	callSignature := fmt.Sprintf("%s|%s|%s", httpRequest.Method, httpRequest.URL.String(), a.Start)
	reqData.ApiRes[callSignature] = apiResponseStructured
	return apiResponseStructured, nil
}
