package api

import (
	"context"
	"fmt"
	"ifttt/handler/common"
	"ifttt/handler/domain/resolvable"
	"strings"
)

type Condition struct {
	ConditionType string                `json:"conditionType" mapstructure:"conditionType"`
	Conditions    []Condition           `json:"conditions" mapstructure:"conditions"`
	Group         bool                  `json:"group" mapstructure:"group"`
	Operator1     resolvable.Resolvable `json:"op1" mapstructure:"op1"`
	Operand       string                `json:"opnd" mapstructure:"opnd"`
	Operator2     resolvable.Resolvable `json:"op2" mapstructure:"op2"`
}

func (c *Condition) EvaluateCondition(ctx context.Context, resolvableDependencies map[string]any) (bool, error) {
	if c.Group {
		return false, fmt.Errorf("method EvaluateCondition: object is a set")
	}
	evaluator := common.GetEvaluator(c.Operand)
	if evaluator == nil {
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
	ev := (*evaluator)(op1Res, op2Res)
	return ev, nil
}

func (group *Condition) EvaluateGroup(ctx context.Context, resolvableDependencies map[string]any) (bool, error) {
	if !group.Group {
		return false, fmt.Errorf("method EvaluateGroup: object is not a group")
	}
	condType := strings.ToLower(group.ConditionType)
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

		switch {
		case condType == "and" && !ev:
			return false, nil
		case condType == "or" && ev:
			return true, nil
		case condType != "and" && condType != "or":
			return false, fmt.Errorf("method EvaluateGroup: condition type not in (and,or)")
		}
	}
	return condType == "and", nil
}
