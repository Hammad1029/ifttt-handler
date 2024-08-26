package api

import (
	"context"
	"fmt"
	"ifttt/handler/domain/resolvable"
	"strings"
)

type Dumping struct {
	Table    string                           `json:"table" mapstructure:"table"`
	Mappings map[string]resolvable.Resolvable `json:"mappings" mapstructure:"mappings"`
}

func (d *Dumping) CreateInsertQuery(ctx context.Context, dependencies map[string]any) (string, []any, error) {
	var parameters []any
	var columns []string
	var positional []string
	queryString := fmt.Sprintf("INSERT INTO %s", d.Table)

	for tableCol, accessor := range d.Mappings {
		resolved, err := accessor.Resolve(ctx, dependencies)
		if err != nil {
			return queryString, nil, fmt.Errorf("method *Dumping.CreateInsertQuery: error in resolving: %s", err)
		}

		columns = append(columns, tableCol)
		positional = append(positional, "?")
		parameters = append(parameters, resolved)
	}

	queryString = fmt.Sprintf("%s (%s) VALUES (%s)",
		queryString, strings.Join(columns, ", "), strings.Join(positional, ","))

	return queryString, parameters, nil
}
