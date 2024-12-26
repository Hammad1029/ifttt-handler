package infrastructure

import (
	"context"
	"fmt"
	"ifttt/handler/domain/orm_schema"
	"strings"
)

type MySqlOrmQueryRepository struct {
	*MySqlBaseRepository
}

func NewMySqlOrmQueryRepository(base *MySqlBaseRepository) *MySqlOrmQueryRepository {
	return &MySqlOrmQueryRepository{MySqlBaseRepository: base}
}

func (m *MySqlOrmQueryRepository) ExecuteSelect(
	tableName string,
	projections map[string]string,
	populate []orm_schema.Populate,
	conditionsTemplate string,
	conditionsValue []any,
	schemaRepo orm_schema.CacheRepository,
	ctx context.Context,
) ([]map[string]any, error) {
	projectionGroups, err := orm_schema.BuildProjectionGroups(projections)
	if err != nil {
		return nil, err
	}
	selectClauses, joinClauses, err := buildQuery(tableName, projectionGroups, populate, schemaRepo, ctx)
	if err != nil {
		return nil, err
	}
	query := fmt.Sprintf("SELECT %s FROM `%s` AS `%s` %s WHERE %s",
		strings.Join(selectClauses, ","), tableName, tableName, strings.Join(joinClauses, " "), conditionsTemplate)
	var results []map[string]any
	if err := m.client.Raw(query, conditionsValue...).Scan(&results).Error; err != nil {
		return nil, err
	}
	return results, err
}

func buildQuery(
	tableName string, projectionGroups map[string]map[string]string, populate []orm_schema.Populate, schemaRepo orm_schema.CacheRepository, ctx context.Context,
) ([]string, []string, error) {
	var selectClauses []string
	var joinClauses []string

	rootSchema, err := schemaRepo.GetSchema(tableName, ctx)
	if err != nil {
		return nil, nil, err
	}
	selectClauses = append(selectClauses,
		buildSelectClauses(projectionGroups[rootSchema.TableName], rootSchema.TableName, rootSchema)...)
	if selectParts, joinParts, err :=
		buildJoins(projectionGroups, &populate, rootSchema, rootSchema.TableName, schemaRepo, ctx); err != nil {
		return nil, nil, err
	} else {
		selectClauses = append(selectClauses, selectParts...)
		joinClauses = append(joinClauses, joinParts...)
	}

	return selectClauses, joinClauses, nil
}

func buildSelectClauses(projections map[string]string, alias string, schema *orm_schema.Schema) []string {
	if projections == nil {
		return []string{}
	}

	var clauses []string
	selectAll := len(projections) == 0
	for _, col := range schema.Columns {
		if selectAll {
			clauses =
				append(clauses, fmt.Sprintf("`%s`.`%s` AS `%s.%s`", alias, col.ColumnName, alias, col.ColumnName))
		} else if as, ok := projections[fmt.Sprintf("%s.%s", alias, col.ColumnName)]; ok {
			if as == "" {
				clauses =
					append(clauses, fmt.Sprintf("`%s`.`%s` AS `%s.%s`", alias, col.ColumnName, alias, col.ColumnName))
			} else {
				clauses = append(clauses,
					fmt.Sprintf("`%s`.`%s` AS `%s.%s`", alias, col.ColumnName, alias, as))
			}
		}
	}
	return clauses
}

func buildJoins(
	projectionGroups map[string]map[string]string, populate *[]orm_schema.Populate, parentSchema *orm_schema.Schema,
	parentAlias string, schemaRepo orm_schema.CacheRepository, ctx context.Context,
) ([]string, []string, error) {
	selectClauses := []string{}
	joinClauses := []string{}
	for _, p := range *populate {
		var foundConstraint *orm_schema.Constraint
		childSchema, err := schemaRepo.GetSchema(p.Table, ctx)
		if err != nil {
			return nil, nil, err
		}
		for _, c := range childSchema.Constraints {
			if c.ReferencesTable == parentSchema.TableName && c.ColumnName == p.Column {
				foundConstraint = &c
				break
			}
		}
		if foundConstraint == nil {
			return nil, nil, fmt.Errorf("invalid relation between '%s' to '%s'", p.Table, parentSchema.TableName)
		}

		childAlias := fmt.Sprintf("%s_%s", parentAlias, p.Alias)
		selectClauses = append(selectClauses,
			buildSelectClauses(projectionGroups[p.Alias], childAlias, childSchema)...)
		joinClauses = append(joinClauses,
			fmt.Sprintf("LEFT OUTER JOIN `%s` AS `%s` ON `%s`.`%s` = `%s`.`%s`",
				childSchema.TableName, childAlias, childAlias, p.Column, parentAlias, foundConstraint.ReferencesField,
			))

		if recursiveSelect, recursiveJoins, err :=
			buildJoins(projectionGroups, &p.Populate, childSchema, childAlias, schemaRepo, ctx); err != nil {
			return nil, nil, err
		} else {
			selectClauses = append(selectClauses, recursiveSelect...)
			joinClauses = append(joinClauses, recursiveJoins...)
		}
	}

	return selectClauses, joinClauses, nil
}
