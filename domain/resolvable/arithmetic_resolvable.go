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

func (a *arithmetic) Resolve(ctx context.Context, dependencies map[string]any) (any, error) {
	opFunc := common.GetArithmeticOperator(a.Operation)
	if opFunc == nil {
		return nil, fmt.Errorf("method Arithmetic: operation %s not found", a.Operation)
	}

	var err error
	var result any

	for _, op := range a.Operators {
		var val any
		var e error
		if op.Group {
			val, e = op.Resolve(ctx, dependencies)
		} else {
			val, e = op.Value.Resolve(ctx, dependencies)
		}
		if e != nil {
			err = fmt.Errorf("method Arithmetic: %s", e)
			break
		}
		if result == nil {
			result = val
		} else {
			result = (*opFunc)(result, val)
		}
	}

	return result, err
}
