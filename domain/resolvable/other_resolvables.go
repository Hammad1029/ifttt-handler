package resolvable

import (
	"context"
	"fmt"
)

type CallRuleResolvable struct {
	RuleId string `json:"ruleId" mapstructure:"ruleId"`
}

func (c *CallRuleResolvable) Resolve(ctx context.Context, dependencies map[string]any) (any, error) {
	return nil, fmt.Errorf("method *CallRuleResolvable.Resolve: wrong execution of rule")
}
