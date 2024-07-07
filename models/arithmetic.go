package models

import (
	"context"
	"fmt"
	"handler/utils"
)

type Arithmetic struct {
	Group     bool         `json:"group" mapstructure:"group"`
	Operation string       `json:"operation" mapstructure:"operation"`
	Operators []Arithmetic `json:"operators" mapstructure:"operators"`
	Get       Resolvable   `json:"get" mapstructure:"get"`
}

func (a *Arithmetic) Arithmetic(ctx context.Context) (interface{}, error) {
	opFuncs := utils.GetArithmeticOperators()
	currFunc, ok := opFuncs[a.Operation]

	if !ok {
		return nil, fmt.Errorf("method Arithmetic: operation %s not found", a.Operation)
	}

	var err error
	var result interface{}

	for _, op := range a.Operators {
		var val interface{}
		var e error
		if op.Group {
			val, e = a.Arithmetic(ctx)
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
