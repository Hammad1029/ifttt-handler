package orm_schema

import (
	"fmt"
	"strings"
)

func (s *Schema) GetPrimaryKey() *Constraint {
	for _, c := range s.Constraints {
		if c.ConstraintType == "PRIMARY KEY" {
			return &c
		}
	}
	return nil
}

func BuildProjectionGroups(projections map[string]string) (map[string]map[string]string, error) {
	groups := map[string]map[string]string{}
	for k, v := range projections {
		split := strings.Split(k, ".")
		groupName := split[0]
		if len(split) < 2 {
			return nil, fmt.Errorf("invalid projection: %s", k)
		}
		if _, ok := groups[groupName]; !ok {
			groups[groupName] = map[string]string{}
		}
		groups[groupName][k] = v
	}
	return groups, nil
}
