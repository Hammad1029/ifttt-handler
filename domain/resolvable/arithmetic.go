package resolvable

import (
	"context"
	"fmt"
	"ifttt/handler/common"
)

type arithmetic struct {
	Group     bool         `json:"group" mapstructure:"group"`
	Operation string       `json:"operation" mapstructure:"operation"`
	Operators []arithmetic `json:"operators" mapstructure:"operators"`
	Value     Resolvable   `json:"value" mapstructure:"value"`
}

func (a *arithmetic) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	opFunc := common.GetCalculator(a.Operation)
	if opFunc == nil {
		return nil, fmt.Errorf("method Arithmetic: operation %s not found", a.Operation)
	}

	var result any

	for _, op := range a.Operators {
		var val any
		var err error
		if op.Group {
			val, err = op.Resolve(ctx, dependencies)
		} else {
			val, err = op.Value.Resolve(ctx, dependencies)
		}
		if err != nil {
			return nil, fmt.Errorf("method Arithmetic: %s", err)
		}
		if result == nil {
			result = val
		} else if r, err := (*opFunc)(result, val); err != nil {
			return nil, err
		} else {
			result = r
		}
	}

	return result, nil
}
