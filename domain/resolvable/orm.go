package resolvable

import (
	"context"
	"fmt"
	"ifttt/handler/common"
	"ifttt/handler/domain/orm_schema"
	"sync"

	"github.com/samber/lo"
)

type orm struct {
	Query     *query                   `json:"query" mapstructure:"query"`
	Operation string                   `json:"operation" mapstructure:"operation"`
	Model     string                   `json:"model" mapstructure:"model"`
	Project   *[]orm_schema.Projection `json:"project" mapstructure:"project"`
	Columns   *map[string]any          `json:"columns" mapstructure:"columns"`
	Populate  *[]orm_schema.Populate   `json:"populate" mapstructure:"populate"`
	Where     *orm_schema.Where        `json:"where" mapstructure:"where"`
	OrderBy   string                   `json:"orderBy" mapstructure:"orderBy"`
	Limit     int                      `json:"limit" mapstructure:"limit"`
}

func (o *orm) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	ormRepo, ok := dependencies[common.DependencyOrmCacheRepo].(orm_schema.CacheRepository)
	if !ok {
		return nil, fmt.Errorf("could not cast orm repo")
	}

	if o.Query == nil {
		return nil, fmt.Errorf("query resolvable is null")
	}

	model, err := ormRepo.GetModel(o.Model, ctx)
	if err != nil {
		return nil, err
	}

	switch o.Operation {
	case common.OrmSelect, common.OrmInsert:
		if results, err := o.Query.Resolve(ctx, dependencies); err != nil {
			return nil, err
		} else if queryData, ok := results.(*queryData); !ok {
			return nil, fmt.Errorf("could not cast results to query data")
		} else if transformed, err := o.transformResults(
			queryData.Results, model.Name, model, o.Project, o.Populate, ormRepo, ctx); err != nil {
			return nil, err
		} else {
			queryData.Results = &transformed
			return queryData, nil
		}
	default:
		return nil, fmt.Errorf("unsupported operation: %s", o.Operation)
	}
}

func (o *orm) transformResults(
	rawResults *[]map[string]any,
	alias string,
	currModel *orm_schema.Model,
	customProjections *[]orm_schema.Projection,
	populate *[]orm_schema.Populate,
	ormRepo orm_schema.CacheRepository,
	ctx context.Context,
) ([]map[string]any, error) {
	if currModel.PrimaryKey == "" {
		return nil, fmt.Errorf("primary key not found in model %s", currModel.Name)
	}

	var pKeyAccessor string
	if o.Operation == common.OrmSelect {
		pKeyAccessor = fmt.Sprintf("%s.%s", alias, currModel.PrimaryKey)
	} else {
		pKeyAccessor = currModel.PrimaryKey
	}
	pKeyOrder := lo.Uniq(lo.FilterMap(*rawResults, func(row map[string]any, _ int) (any, bool) {
		return row[pKeyAccessor], row[pKeyAccessor] != nil
	}))
	grouped := lo.GroupBy(*rawResults, func(row map[string]any) any {
		return row[pKeyAccessor]
	})
	delete(grouped, nil)

	flattened := make([]map[string]any, len(pKeyOrder))
	wg := sync.WaitGroup{}
	cancelCtx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	for idx, key := range pKeyOrder {
		rowGroup, ok := grouped[key]
		if !ok {
			return nil, fmt.Errorf("group for key %s not found", key)
		}

		wg.Add(1)
		go func(idx int, rowGroup []map[string]any) {
			defer wg.Done()
			select {
			case <-cancelCtx.Done():
				return
			default:
				{
					baseRow := rowGroup[0]
					var transformedRow map[string]any
					if len(*customProjections) > 0 {
						transformedRow = o.projectRow(&baseRow, customProjections, alias)
					} else {
						transformedRow = o.projectRow(&baseRow, &currModel.Projections, alias)
					}
					for _, p := range *populate {

						childModel, err := ormRepo.GetModel(p.Model, ctx)
						if err != nil || childModel == nil {
							cancel(err)
							return
						}
						association := o.findAssociation(currModel, childModel)
						if association == nil {
							cancel(fmt.Errorf("association not found between %s and %s", currModel.Name, childModel.Name))
							return
						}
						childAlias := fmt.Sprintf("%s_%s", alias, p.As)
						childGroups, err := o.transformResults(&rowGroup, childAlias, childModel, &p.Project, &p.Populate, ormRepo, ctx)
						if err != nil {
							cancel(fmt.Errorf("could not transform alias %s for model %s: %s", childAlias, p.Model, err))
							return
						}
						if association.Type == common.AssociationsHasOne ||
							association.Type == common.AssociationsBelongsTo {
							if len(childGroups) > 0 {
								transformedRow[p.As] = childGroups[0]
							} else {
								transformedRow[p.As] = nil
							}
						} else {
							transformedRow[p.As] = childGroups
						}
					}
					flattened[idx] = transformedRow
				}
			}
		}(idx, rowGroup)
	}
	wg.Wait()

	if err := context.Cause(cancelCtx); err != nil {
		return nil, err
	}
	return flattened, nil
}

func (o *orm) projectRow(
	row *map[string]any, projections *[]orm_schema.Projection, alias string,
) map[string]any {
	projectedRow := map[string]any{}
	var accessor string
	for _, p := range *projections {
		if o.Operation == common.OrmSelect {
			accessor = fmt.Sprintf("%s.%s", alias, p.Column)
		} else {
			accessor = p.Column
		}

		if colVal, ok := (*row)[accessor]; !ok || colVal == nil {
			switch p.DataType {
			case common.DatabaseTypeString:
				projectedRow[p.As] = new(string)
			case common.DatabaseTypeNumber:
				projectedRow[p.As] = new(int)
			case common.DatabaseTypeBoolean:
				projectedRow[p.As] = new(bool)
			default:
				projectedRow[p.As] = nil
			}
		} else if p.DataType == common.DatabaseTypeString {
			projectedRow[p.As] = fmt.Sprint(colVal)
		} else {
			projectedRow[p.As] = colVal
		}

	}
	return projectedRow
}

func (o *orm) findAssociation(
	parent *orm_schema.Model, join *orm_schema.Model) *orm_schema.ModelAssociation {
	for _, a := range parent.OwningAssociations {
		if a.ReferencesModel.Name == join.Name {
			return &a
		}
	}
	for _, a := range parent.ReferencedAssociations {
		if a.OwningModel.Name == join.Name {
			return &a
		}
	}
	return nil
}
