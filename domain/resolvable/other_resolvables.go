package resolvable

import (
	"context"
	"fmt"
	"strings"
)

type CallRuleResolvable struct {
	RuleId uint `json:"ruleId" mapstructure:"ruleId"`
}

func (c *CallRuleResolvable) Resolve(ctx context.Context, dependencies map[string]any) (any, error) {
	return c.RuleId, nil
}

type StringInterpolationResolvable struct {
	Template   string       `json:"template" mapstructure:"template"`
	Parameters []Resolvable `json:"parameters" mapstructure:"parameters"`
}

func (s *StringInterpolationResolvable) Resolve(ctx context.Context, dependencies map[string]any) (any, error) {
	resolvedString := s.Template
	for _, r := range s.Parameters {
		if val, err := r.Resolve(ctx, dependencies); err != nil {
			return nil,
				fmt.Errorf("method *StringInterpolationResolvable.Resolve: error in resolving: %s", err)
		} else {
			resolvedString = strings.Replace(resolvedString, "$param", fmt.Sprint(val), 1)
		}
	}
	return resolvedString, nil
}
