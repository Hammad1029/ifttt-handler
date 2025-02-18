package common

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sync"

	"github.com/gofiber/fiber/v2"
)

func GetCtxState(ctx context.Context) *sync.Map {
	if state, ok := ctx.Value(ContextState).(*sync.Map); ok {
		return state
	}
	return nil
}

func GetResponseSent(ctx context.Context) bool {
	requestState := GetCtxState(ctx)
	if v, ok := requestState.Load(ContextResponseSent); ok {
		return v.(bool)
	}
	return false
}

func SetResponseSent(ctx context.Context) bool {
	requestState := GetCtxState(ctx)
	v, ok := requestState.Load(ContextResponseSent)
	if !ok {
		return false
	}

	if v.(bool) {
		return false
	}
	requestState.Store(ContextResponseSent, true)
	return true
}

func BodyParser(c *fiber.Ctx) (*map[string]any, error) {
	output := make(map[string]any)
	bodyBytes := c.Body()
	bodyStr := string(bodyBytes)
	contentType := c.Get("Content-Type")
	switch contentType {
	case "application/json", "application/json; charset=UTF-8":
		if err := json.Unmarshal(bodyBytes, &output); err != nil {
			return nil, err
		}
	case "application/x-www-form-urlencoded":
		if values, err := url.ParseQuery(bodyStr); err != nil {
			return nil, err
		} else {
			for k, v := range values {
				if len(v) > 0 {
					output[k] = v[0]
				}
			}
		}
	default:
		return nil, fmt.Errorf("no parser for %s", contentType)
	}
	return &output, nil
}
