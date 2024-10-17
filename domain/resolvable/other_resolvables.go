package resolvable

import (
	"context"
	"fmt"
	"strings"
)

type stringInterpolationResolvable struct {
	Template   string       `json:"template" mapstructure:"template"`
	Parameters []Resolvable `json:"parameters" mapstructure:"parameters"`
}

func (s *stringInterpolationResolvable) Resolve(ctx context.Context, dependencies map[string]any) (any, error) {
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
