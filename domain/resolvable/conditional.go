package resolvable

import (
	"context"
	"fmt"
	"ifttt/handler/common"
	"strings"
)

type conditional struct {
	Condition Condition     `json:"condition" mapstructure:"condition"`
	True      *[]Resolvable `json:"true" mapstructure:"true"`
	False     *[]Resolvable `json:"false" mapstructure:"false"`
}

func (c *conditional) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	if ev, err := c.Condition.EvaluateGroup(ctx, dependencies); err != nil {
		return nil, err
	} else if ev {
		if _, err := ResolveArrayMust(c.True, ctx, dependencies); err != nil {
			return nil, err
		}
	} else {
		if _, err := ResolveArrayMust(c.False, ctx, dependencies); err != nil {
			return nil, err
		}
	}
	return nil, nil
}

const (
	conditionAnd = "AND"
	conditionOr  = "OR"
)

type Condition struct {
	ConditionType   string      `json:"conditionType" mapstructure:"conditionType"`
	Conditions      []Condition `json:"conditions" mapstructure:"conditions"`
	Group           bool        `json:"group" mapstructure:"group"`
	ComparisionType string      `json:"comparisionType" mapstructure:"comparisionType"`
	Operator1       Resolvable  `json:"op1" mapstructure:"op1"`
	Operand         string      `json:"opnd" mapstructure:"opnd"`
	Operator2       Resolvable  `json:"op2" mapstructure:"op2"`
}

func (group *Condition) EvaluateGroup(ctx context.Context, resolvableDependencies map[common.IntIota]any) (bool, error) {
	if !group.Group {
		return false, fmt.Errorf("method EvaluateGroup: object is not a group")
	}
	condType := strings.ToUpper(group.ConditionType)
	var ev bool
	var err error
	for _, cond := range group.Conditions {
		switch cond.Group {
		case true:
			ev, err = cond.EvaluateGroup(ctx, resolvableDependencies)
		case false:
			ev, err = cond.EvaluateCondition(ctx, resolvableDependencies)
		}

		if err != nil {
			return false, fmt.Errorf("method EvaluateGroup: %s", err)
		}

		switch strings.ToUpper(condType) {
		case conditionAnd:
			if !ev {
				return false, nil
			}
		case conditionOr:
			if ev {
				return true, nil
			}
		default:
			return false,
				fmt.Errorf("method EvaluateGroup: condition type not in (%s,%s)", conditionAnd, conditionOr)
		}
	}
	return condType == conditionAnd, nil
}

func (c *Condition) EvaluateCondition(ctx context.Context, resolvableDependencies map[common.IntIota]any) (bool, error) {
	if c.Group {
		return false, fmt.Errorf("method EvaluateCondition: object is a set")
	}
	comparator := common.GetComparator(c.Operand)
	if comparator == nil {
		return false, fmt.Errorf("method EvaluateCondition: operand not found")
	}

	op1Res, err := c.Operator1.Resolve(ctx, resolvableDependencies)
	if err != nil {
		return false, fmt.Errorf("method EvaluateCondition: %s", err)
	}
	op2Res, err := c.Operator2.Resolve(ctx, resolvableDependencies)
	if err != nil {
		return false, fmt.Errorf("method EvaluateCondition: %s", err)
	}
	if ev, err := (*comparator)(op1Res, op2Res, c.ComparisionType); err != nil {
		return false, err
	} else {
		return ev, nil
	}
}
