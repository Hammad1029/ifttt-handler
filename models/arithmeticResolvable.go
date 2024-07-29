package models

import (
	"context"
	"fmt"
	"handler/common"
)

type Arithmetic struct {
	Group     bool         `json:"group" mapstructure:"group"`
	Operation string       `json:"operation" mapstructure:"operation"`
	Operators []Arithmetic `json:"operators" mapstructure:"operators"`
	Get       Resolvable   `json:"get" mapstructure:"get"`
}

func (a *Arithmetic) Resolve(ctx context.Context) (any, error) {
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
			val, e = a.Resolve(ctx)
		} else {
			val, e = op.Get.Resolve(ctx)
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
