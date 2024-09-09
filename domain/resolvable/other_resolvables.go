package resolvable

import (
	"context"
)

type CallRuleResolvable struct {
	RuleId uint `json:"ruleId" mapstructure:"ruleId"`
}

func (c *CallRuleResolvable) Resolve(ctx context.Context, dependencies map[string]any) (any, error) {
	return c.RuleId, nil
}
