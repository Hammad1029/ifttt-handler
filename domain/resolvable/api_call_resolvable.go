package resolvable

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"ifttt/handler/common"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/fatih/structs"
	"github.com/mitchellh/mapstructure"
	"github.com/samber/lo"
)

type apiCallResolvable struct {
	Method  string         `json:"method" mapstructure:"method"`
	URL     Resolvable     `json:"url" mapstructure:"url"`
	Headers map[string]any `json:"headers" mapstructure:"headers"`
	Body    map[string]any `json:"body" mapstructure:"body"`
	Async   bool           `json:"async" mapstructure:"async"`
	Timeout uint           `json:"timeout" mapstructure:"timeout"`
}

type callData struct {
	Metadata *apiMetadata     `json:"metadata" mapstructure:"metadata"`
	Request  *apiRequest      `json:"request" mapstructure:"request"`
	Response *apiCallResponse `json:"response" mapstructure:"response"`
}

type apiRequest struct {
	Method  string            `json:"method" mapstructure:"method"`
	URL     string            `json:"url" mapstructure:"url"`
	Headers map[string]string `json:"headers" mapstructure:"headers"`
	Body    map[string]any    `json:"body" mapstructure:"body"`
}

type apiCallResponse struct {
	StatusCode int                 `json:"statusCode" mapstructure:"statusCode"`
	Status     string              `json:"status" mapstructure:"status"`
	Headers    map[string][]string `json:"headers" mapstructure:"headers"`
	Body       any                 `json:"body" mapstructure:"body"`
}

type apiMetadata struct {
	Start      time.Time `json:"start" mapstructure:"start"`
	End        time.Time `json:"end" mapstructure:"end"`
	TimeTaken  uint64    `json:"timeTaken" mapstructure:"timeTaken"`
	Timeout    uint      `json:"timeout" mapstructure:"timeout"`
	DidTimeout bool      `json:"didTimeout" mapstructure:"didTimeout"`
	Async      bool      `json:"async" mapstructure:"async"`
	Error      string    `json:"error" mapstructure:"error"`
}

func (a *apiCallResolvable) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	callData, err := a.createCallData(ctx, dependencies)
	if err != nil {
		return nil, fmt.Errorf("apiCallResolvable: could not create calldata: %s", err)
	}

	if a.Async {
		go callData.doRequest(ctx)
	} else if err := callData.doRequest(ctx); err != nil {
		return nil, err
	}

	return callData, nil
}

func (a *apiCallResolvable) createCallData(ctx context.Context, dependencies map[common.IntIota]any) (*callData, error) {
	var callData callData
	callData.Metadata = a.createMetadata()
	request, err := a.createRequest(ctx, dependencies)
	if err != nil {
		return nil, fmt.Errorf("could not create request: %s", err)
	}
	callData.Request = request
	callData.Response = &apiCallResponse{}
	return &callData, nil
}

func (a *apiCallResolvable) createMetadata() *apiMetadata {
	return &apiMetadata{
		Timeout: a.Timeout,
		Async:   a.Async,
	}
}

func (a *apiCallResolvable) createRequest(ctx context.Context, dependencies map[common.IntIota]any) (*apiRequest, error) {
	var request apiRequest

	allowedMethods := []string{"GET", "POST"}
	if !lo.Contains(allowedMethods, a.Method) {
		return nil, fmt.Errorf("request method %s not found", a.Method)
	}
	request.Method = strings.ToUpper(a.Method)

	resolvedURL, err := a.URL.Resolve(ctx, dependencies)
	if err != nil {
		return nil, fmt.Errorf("could not resolve url: %s", err)
	}
	request.URL = fmt.Sprint(resolvedURL)

	if bodyResolved, err := resolveIfNested(a.Body, ctx, dependencies); err != nil {
		return nil, fmt.Errorf("could not resolve request body: %s", err)
	} else if err := mapstructure.Decode(bodyResolved, &request.Body); err != nil {
		return nil, fmt.Errorf("could not decode resolved request body: %s", err)
	}

	if headersResolved, err := resolveIfNested(a.Headers, ctx, dependencies); err != nil {
		return nil, fmt.Errorf("could not resolve headers: %s", err)
	} else if err := mapstructure.Decode(headersResolved, &request.Headers); err != nil {
		return nil, fmt.Errorf("could not decode resolved headers: %s", err)
	}

	return &request, nil
}

func (a *apiRequest) createHttpRequest() (*http.Request, error) {
	var bodyReader io.Reader
	if bodyStringified, err := json.Marshal(a.Body); err == nil {
		bodyReader = strings.NewReader(string(bodyStringified))
	} else {
		return nil, fmt.Errorf("couldn't stringify body: %s", err)
	}

	httpRequest, err := http.NewRequest(a.Method, a.URL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("could not create http request: %s", err)
	}

	for key, val := range a.Headers {
		httpRequest.Header.Add(key, val)
	}

	return httpRequest, nil
}

func (c *callData) doRequest(ctx context.Context) error {
	defer c.createLog(ctx)

	httpRequest, err := c.Request.createHttpRequest()
	if err != nil {
		return fmt.Errorf("error in creating http request: %s", err)
	}
	if c.Metadata.Timeout > 0 {
		timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(c.Metadata.Timeout)*time.Millisecond)
		httpRequest = httpRequest.WithContext(timeoutCtx)
		defer cancel()
	}

	c.Metadata.Start = time.Now()
	res, err := http.DefaultClient.Do(httpRequest)
	if err != nil {
		if httpRequest.Context().Err() == context.DeadlineExceeded {
			c.Metadata.DidTimeout = true
		} else {
			common.LogWithTracer(common.LogUser, "error in executing api call", err, true, ctx)
			c.Metadata.Error = err.Error()
			return err
		}
		return nil
	}

	localRes, err := c.createResponse(res)
	if err != nil {
		return err
	}
	c.Metadata.End = time.Now()
	c.Metadata.TimeTaken = uint64(c.Metadata.End.Sub(c.Metadata.Start).Milliseconds())
	c.Response = localRes

	return nil
}

func (c *callData) createResponse(res *http.Response) (*apiCallResponse, error) {
	var response apiCallResponse

	response.StatusCode = res.StatusCode
	response.Status = res.Status

	respHeadersMap := make(map[string][]string)
	for key, arr := range res.Header {
		respHeadersMap[key] = arr
	}
	response.Headers = respHeadersMap

	if err := response.readResponseBody(res); err != nil {
		return nil, err
	}

	return &response, nil
}

func (a *apiCallResponse) readResponseBody(res *http.Response) error {
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("could not read io/response: %s", err)
	}

	switch res.Header.Get("Content-Type") {
	case "application/json":
		if err := json.Unmarshal(body, &a.Body); err != nil {
			return fmt.Errorf("error parsing JSON: %s", err)
		}
	case "application/xml", "text/xml":
		if err := xml.Unmarshal(body, &a.Body); err != nil {
			return fmt.Errorf("error parsing XML: %s", err)
		}
	default:
		a.Body = string(body)
	}

	return nil
}

func (c *callData) createLog(ctx context.Context) {
	reqData := GetRequestData(ctx)
	callSignature := fmt.Sprintf("%s|%s|%s",
		c.Request.Method, c.Request.URL, c.Metadata.Start.Format(common.DateTimeFormat))
	reqData.ApiRes[callSignature] = structs.Map(c)

	ctxState := common.GetCtxState(ctx)
	if ctxState != nil {
		if externalExecTime, ok := ctxState.Load(common.ContextExternalExecTime); ok {
			ctxState.Store(common.ContextExternalExecTime, externalExecTime.(uint64)+c.Metadata.TimeTaken)
		}
	}
}
