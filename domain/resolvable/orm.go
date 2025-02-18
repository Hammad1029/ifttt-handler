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
	Query           *query                   `json:"query" mapstructure:"query"`
	SuccessiveQuery *query                   `json:"successiveQuery" mapstructure:"successiveQuery"`
	Operation       string                   `json:"operation" mapstructure:"operation"`
	Model           string                   `json:"model" mapstructure:"model"`
	Project         *[]orm_schema.Projection `json:"project" mapstructure:"project"`
	Columns         *map[string]any          `json:"columns" mapstructure:"columns"`
	Populate        *[]orm_schema.Populate   `json:"populate" mapstructure:"populate"`
	Where           *orm_schema.Where        `json:"where" mapstructure:"where"`
	OrderBy         string                   `json:"orderBy" mapstructure:"orderBy"`
	Limit           int                      `json:"limit" mapstructure:"limit"`
	ModelsInUse     *[]string                `json:"modelsInUse" mapstructure:"modelsInUse"`
}

func (o *orm) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	if o.Query == nil {
		return nil, fmt.Errorf("query resolvable is null")
	}

	ormRepo, ok := dependencies[common.DependencyOrmCacheRepo].(orm_schema.CacheRepository)
	if !ok {
		return nil, fmt.Errorf("could not cast orm repo")
	}

	rawQueryRepo, ok := dependencies[common.DependencyRawQueryRepo].(RawQueryRepository)
	if !ok {
		return nil, fmt.Errorf("method *QueryResolvable: could not cast raw query repo")
	}

	switch o.Operation {
	case common.OrmSelect, common.OrmInsert, common.OrmUpdate, common.OrmDelete:
	default:
		return nil, fmt.Errorf("unsupported operation: %s", o.Operation)
	}

	modelsInUse := make(map[string]*orm_schema.Model)
	for _, m := range *o.ModelsInUse {
		if childModel, err := ormRepo.GetModel(m, ctx); err != nil {
			return nil, err
		} else {
			modelsInUse[childModel.Name] = childModel
		}
	}

	mainModel, ok := modelsInUse[o.Model]
	if !ok {
		return nil, fmt.Errorf("main model %s not found", o.Model)
	}

	resolved, err := resolveArrayMustParallel(&o.Query.Parameters, ctx, dependencies)
	if err != nil {
		return nil, err
	}

	tx, err := rawQueryRepo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not begin tx: %s", err)
	}

	var queryData *queryData
	queryData, err = o.Query.init(tx, resolved, ctx, dependencies)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return nil, fmt.Errorf("attempted rollback on error: %s. rollback failed %s", err, rollbackErr)
		}
		return nil, fmt.Errorf("rolled back. error: %s", err)
	}

	if o.SuccessiveQuery == nil {
		if transformed, err := o.transformResults(
			queryData.Results, mainModel.Name, mainModel, o.Project, o.Populate, &modelsInUse, ctx,
		); err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return nil, fmt.Errorf("attempted rollback on error: %s. rollback failed %s", err, rollbackErr)
			}
			return nil, fmt.Errorf("rolled back. error: %s", err)
		} else if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("commit failed: %s", err)
		} else {
			queryData.Results = &transformed
			return queryData, nil
		}
	}

	queryData, err = o.SuccessiveQuery.init(tx, resolved, ctx, dependencies)
	if err != nil {
		return nil, err
	}

	if transformed, err := o.transformResults(
		queryData.Results, mainModel.Name, mainModel, nil, nil, &modelsInUse, ctx,
	); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return nil, fmt.Errorf("attempted rollback on error: %s. rollback failed %s", err, rollbackErr)
		}
		return nil, fmt.Errorf("rolled back. error: %s", err)
	} else if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit failed: %s", err)
	} else {
		queryData.Results = &transformed
		return queryData, nil
	}
}

func (o *orm) transformResults(
	rawResults *[]map[string]any,
	alias string,
	currModel *orm_schema.Model,
	customProjections *[]orm_schema.Projection,
	populate *[]orm_schema.Populate,
	modelsInUse *map[string]*orm_schema.Model,
	ctx context.Context,
) ([]map[string]any, error) {
	if rawResults == nil {
		return nil, nil
	} else if currModel.PrimaryKey == "" {
		return nil, fmt.Errorf("primary key not found in model %s", currModel.Name)
	}

	pKeyAccessor := fmt.Sprintf("%s.%s", alias, currModel.PrimaryKey)
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
					var (
						transformedRow map[string]any
						err            error
					)
					if customProjections != nil && len(*customProjections) > 0 {
						transformedRow, err = o.projectRow(&baseRow, customProjections, alias)
					} else {
						transformedRow, err = o.projectRow(&baseRow, &currModel.Projections, alias)
					}
					if err != nil {
						cancel(err)
						return
					}
					select {
					case <-ctx.Done():
						return
					default:
						{
							if populate != nil && len(*populate) != 0 {
								for _, p := range *populate {

									childModel, ok := (*modelsInUse)[p.Model]
									if !ok || childModel == nil {
										cancel(fmt.Errorf("model %s not found", p.Model))
										return
									}
									association := o.findAssociation(currModel, childModel)
									if association == nil {
										cancel(fmt.Errorf("association not found between %s and %s", currModel.Name, childModel.Name))
										return
									}
									childAlias := fmt.Sprintf("%s_%s", alias, p.As)
									childGroups, err := o.transformResults(&rowGroup, childAlias, childModel, &p.Project, &p.Populate, modelsInUse, ctx)
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
							}
							flattened[idx] = transformedRow
						}
					}
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
) (map[string]any, error) {
	projectedRow := make(map[string]any, len(*projections))
	prefix := alias + "."

	for _, projection := range *projections {
		p := &projection
		accessor := prefix + p.Column

		colVal, exists := (*row)[accessor]
		if sanitized, err := p.SanitizeValue(colVal, exists); err != nil {
			return nil, err
		} else {
			projectedRow[p.As] = sanitized
		}
	}

	return projectedRow, nil
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
