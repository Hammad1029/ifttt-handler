package resolvable

import (
	"context"
	"fmt"
	"ifttt/handler/common"
	"ifttt/handler/domain/orm_schema"
	"strings"
)

type ormResolvable struct {
	Operation          string                `json:"operation" mapstructure:"operation"`
	Table              string                `json:"table" mapstructure:"table"`
	Projections        map[string]string     `json:"projections" mapstructure:"projections"`
	ConditionsTemplate string                `json:"conditionsTemplate" mapstructure:"conditionsTemplate"`
	ConditionsValue    []any                 `json:"conditionsValue" mapstructure:"conditionsValue"`
	Columns            map[string]any        `json:"columns" mapstructure:"columns"`
	Populate           []orm_schema.Populate `json:"populate" mapstructure:"populate"`
}

type OrmQueryBuilderRepository interface {
	ExecuteSelect(
		tableName string,
		projections map[string]string,
		populate []orm_schema.Populate,
		conditionsTemplate string,
		conditionsValue []any,
		schemaRepo orm_schema.CacheRepository,
		ctx context.Context,
	) ([]map[string]any, error)
}

func (o *ormResolvable) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	schemaRepo, ok := dependencies[common.DependencyOrmSchemaRepo].(orm_schema.CacheRepository)
	if !ok {
		return nil, fmt.Errorf("could not cast schema repo")
	}

	qb, ok := dependencies[common.DependencyOrmQueryRepo].(OrmQueryBuilderRepository)
	if !ok {
		return nil, fmt.Errorf("dependency orm query repo not found")
	}

	schema, err := schemaRepo.GetSchema(o.Table, ctx)
	if err != nil {
		return nil, err
	}

	switch o.Operation {
	case common.OrmSelect:
		conditionsResolved, err := resolveSlice(&o.ConditionsValue, ctx, dependencies)
		if err != nil {
			return nil, err
		}
		res, err := qb.ExecuteSelect(
			o.Table, o.Projections, o.Populate, o.ConditionsTemplate, conditionsResolved, schemaRepo, ctx,
		)
		if err != nil {
			return nil, err
		}
		projectionGroups, err := orm_schema.BuildProjectionGroups(o.Projections)
		if err != nil {
			return nil, err
		}
		primaryKey := schema.GetPrimaryKey()
		if primaryKey == nil {
			return nil, fmt.Errorf("primary key not found")
		}
		results, err := o.transformResults(
			&res, schema.TableName, &o.Populate, &projectionGroups, primaryKey, schemaRepo, ctx,
		)
		if err != nil {
			return nil, err
		}
		return results, nil
	default:
		return nil, fmt.Errorf("unsupported operation: %s", o.Operation)
	}
}

func (o *ormResolvable) transformResults(
	rawResults *[]map[string]any, alias string, populate *[]orm_schema.Populate,
	projectionGroups *map[string]map[string]string, primaryKey *orm_schema.Constraint,
	schemaRepo orm_schema.CacheRepository, ctx context.Context,
) ([]map[string]any, error) {
	groupedResults := map[any]map[string]any{}

	for _, row := range *rawResults {
		pKeyValue, ok := row[fmt.Sprintf("%s.%s", alias, primaryKey.ColumnName)]
		if !ok {
			return nil, fmt.Errorf("primary key not found in recordset: %s.%s",
				primaryKey.TableName, primaryKey.ColumnName)
		} else if _, ok := groupedResults[primaryKey.ColumnName]; !ok {
			groupedResults[pKeyValue] = map[string]any{}
			for _, p := range *populate {
				arr, ok := groupedResults[pKeyValue][p.Alias]
				if !ok {
					arr = []map[string]any{}
				}
				arrCasted, ok := arr.([]map[string]any)
				if !ok {
					return nil, fmt.Errorf("could not cast %s.%s to array", pKeyValue, p.Alias)
				}
				childSchema, err := schemaRepo.GetSchema(p.Table, ctx)
				if err != nil {
					return nil, fmt.Errorf("could not get schema %s: %s", p.Table, err)
				}
				if childPrimaryKey := childSchema.GetPrimaryKey(); childPrimaryKey == nil {
					return nil, fmt.Errorf("primary key not found")
				} else if pVal, ok := row[fmt.Sprintf("%s.%s", p.Alias, childPrimaryKey.ColumnName)]; ok && pVal != nil {
					if childResults, err := o.transformResults(
						rawResults, p.Alias, &p.Populate, projectionGroups, primaryKey, schemaRepo, ctx,
					); err != nil {
						return nil, err
					} else {
						arrCasted = append(arrCasted, childResults...)
					}
				}
				groupedResults[pKeyValue][p.Alias] = arrCasted
			}
			if pGroup, ok := (*projectionGroups)[alias]; !ok {
				return nil, fmt.Errorf("projection group not found for %s", alias)
			} else {
				for k, v := range pGroup {
					if v == "" {
						split := strings.Split(k, ".")
						if len(split) < 2 {
							return nil, fmt.Errorf("invalid projection %s", k)
						}
						groupedResults[pKeyValue][split[len(split)-1]] = row[k]
					} else if val, ok := row[v]; ok {
						groupedResults[pKeyValue][v] = val
					} else {
						return nil, fmt.Errorf("invalid projection %s", k)
					}
				}
			}
		}

	}

	transformedResults := []map[string]any{}
	for _, row := range groupedResults {
		transformedResults = append(transformedResults, row)
	}

	return transformedResults, nil
}
