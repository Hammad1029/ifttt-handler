package common

import (
	"fmt"
	"sync"

	"github.com/itchyny/gojq"
)

func SyncMapUnsync(s *sync.Map) map[string]any {
	m := make(map[string]any)
	s.Range(func(key, value any) bool {
		m[fmt.Sprint(key)] = value
		return true
	})
	return m
}

func SyncMapJQSet(sm *sync.Map, key string, value any) error {
	queryString := fmt.Sprintf("%s = $value", key)
	if queryString[0] != '.' {
		queryString = "." + queryString
	}
	parsed, err := gojq.Parse(queryString)
	if err != nil {
		return fmt.Errorf("invalid jq query, could not parse query %s: %v", queryString, err)
	}
	compiled, err := gojq.Compile(parsed, gojq.WithVariables([]string{"$value"}))
	if err != nil {
		return fmt.Errorf("invalid jq query, could not compile query %s: %v", queryString, err)
	}
	unsynced := SyncMapUnsync(sm)
	resultIter := compiled.Run(unsynced, value)
	for {
		v, ok := resultIter.Next()
		if !ok {
			break
		}
		if updated, ok := v.(map[string]interface{}); ok {
			for k, v := range updated {
				sm.Store(k, v)
			}
			break
		}
	}
	return nil
}

func SyncMapJQGet(sm *sync.Map, key string) (any, error) {
	parsed, err := gojq.Parse(key)
	if err != nil {
		return nil, fmt.Errorf("invalid jq query, could not parse: %v", err)
	}
	iter := parsed.Run(SyncMapUnsync(sm))
	for {
		if v, ok := iter.Next(); !ok {
			return nil, nil
		} else {
			return v, nil
		}
	}
}

func MapJqGet(m *map[string]any, key string) (any, error) {
	parsed, err := gojq.Parse(key)
	if err != nil {
		return nil, fmt.Errorf("invalid jq query, could not parse: %v", err)
	}
	iter := parsed.Run(m)
	for {
		if v, ok := iter.Next(); !ok {
			return nil, nil
		} else {
			return v, nil
		}
	}
}

func MapJQSet(m *map[string]any, key string, value any) error {
	queryString := fmt.Sprintf("%s = $value", key)
	if queryString[0] != '.' {
		queryString = "." + queryString
	}
	parsed, err := gojq.Parse(queryString)
	if err != nil {
		return fmt.Errorf("invalid jq query, could not parse query %s: %v", queryString, err)
	}
	compiled, err := gojq.Compile(parsed, gojq.WithVariables([]string{"$value"}))
	if err != nil {
		return fmt.Errorf("invalid jq query, could not compile query %s: %v", queryString, err)
	}

	resultIter := compiled.Run(*m, value)
	for {
		if v, ok := resultIter.Next(); !ok {
			break
		} else if err, ok := v.(error); ok {
			return err
		} else if updated, ok := v.(map[string]interface{}); ok {
			*m = updated
			break
		}
	}
	return nil
}
