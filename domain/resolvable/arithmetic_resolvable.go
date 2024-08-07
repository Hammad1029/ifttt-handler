package resolvable

import (
	"context"
	"fmt"
	"handler/common"
)

type Arithmetic struct {
	Group     bool         `json:"group" mapstructure:"group"`
	Operation string       `json:"operation" mapstructure:"operation"`
	Operators []Arithmetic `json:"operators" mapstructure:"operators"`
	Value     Resolvable   `json:"value" mapstructure:"value"`
}

func (a *Arithmetic) Resolve(ctx context.Context, optional ...any) (any, error) {
	opFuncs := common.GetArithmeticOperators()
	currFunc, ok := opFuncs[a.Operation]

	if !ok {
		return nil, fmt.Errorf("method Arithmetic: operation %s not found", a.Operation)
	}

	var err error
	var result any

	for _, op := range a.Operators {
		var val any
		var e error
		if op.Group {
			val, e = a.Resolve(ctx, optional)
		} else {
			val, e = op.Value.Resolve(ctx, optional)
		}
		if e != nil {
			err = fmt.Errorf("method Arithmetic: %s", e)
			break
		}
		if result == nil {
			result = val
		} else {
			result = currFunc(result, val)
		}
	}

	return result, err
}
