package api

import (
	"context"
	"fmt"
	"handler/domain/resolvable"
	"strings"
)

type Dumping struct {
	Table    string            `json:"table" mapstructure:"table"`
	Mappings map[string]string `json:"mappings" mapstructure:"mappings"`
}

func (d *Dumping) CreateInsertQuery(ctx context.Context, dependencies map[string]any) (string, []any, error) {
	var parameters []any
	var columns []string
	var positional []string
	queryString := fmt.Sprintf("INSERT INTO %s", d.Table)

	jqResolvable := &resolvable.Resolvable{
		ResolveType: "jq",
		ResolveData: map[string]any{
			"input": resolvable.Resolvable{
				ResolveType: resolvable.AccessorGetRequestResolvable,
				ResolveData: map[string]any{},
			},
		}}
	for accessor, tableCol := range d.Mappings {
		jqResolvable.ResolveData["query"] = accessor
		resolved, err := jqResolvable.Resolve(ctx, dependencies)
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
