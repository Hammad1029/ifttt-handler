package resolvable

import (
	"context"
	"fmt"
	"ifttt/handler/common"
	"ifttt/handler/domain/orm_schema"
	"strconv"
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
	if o.Query == nil {
		return nil, fmt.Errorf("query resolvable is null")
	}

	ormRepo, ok := dependencies[common.DependencyOrmCacheRepo].(orm_schema.CacheRepository)
	if !ok {
		return nil, fmt.Errorf("could not cast orm repo")
	}

	var (
		namedResolved      map[string]any
		positionalResolved []any
		err                error
	)

	model, err := ormRepo.GetModel(o.Model, ctx)
	if err != nil {
		return nil, err
	}

	switch o.Operation {
	case common.OrmSelect:
		positionalResolved, err = resolveArrayMustParallel(&o.Query.PositionalParameters, ctx, dependencies)
	case common.OrmInsert:
		namedResolved, err = resolveMapMustParallel(&o.Query.NamedParameters, ctx, dependencies)
	default:
		return nil, fmt.Errorf("unsupported operation: %s", o.Operation)
	}

	if err != nil {
		return nil, err
	} else if err := o.verifyParameters(&positionalResolved, &namedResolved, model, ctx); err != nil {
		return nil, err
	} else if queryData, err := o.Query.init(positionalResolved, namedResolved, ctx, dependencies); err != nil {
		return nil, err
	} else if transformed, err := o.transformResults(
		queryData.Results, model.Name, model, o.Project, o.Populate, ormRepo, ctx); err != nil {
		return nil, err
	} else {
		queryData.Results = &transformed
		return queryData, nil
	}
}

func (o *orm) verifyParameters(
	positional *[]any, named *map[string]any, model *orm_schema.Model, ctx context.Context,
) error {
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	wg := sync.WaitGroup{}
	safeMap := sync.Map{}
	available := make(map[string]orm_schema.Projection, len(model.Projections))
	for _, p := range model.Projections {
		available[p.Column] = p
	}

	if positional != nil {
	}
	if named != nil {
		for key, param := range *named {
			wg.Add(1)
			go func(k string, v any) {
				defer wg.Done()
				projection, ok := available[k]
				if !ok {
					cancel(fmt.Errorf("config for column %s not found", k))
				}
				select {
				case <-ctx.Done():
					return
				default:
					if sanitized, err := o.sanitize(v, true, &projection); err != nil {
						cancel(err)
					} else {
						select {
						case <-ctx.Done():
							return
						default:
							safeMap.Store(k, sanitized)
						}
					}
				}

			}(key, param)
		}
		wg.Wait()

		if err := context.Cause(ctx); err != nil {
			return err
		}
		safeMap.Range(func(key, value any) bool {
			(*named)[fmt.Sprint(key)] = value
			return true
		})
	}

	return nil
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
					var (
						transformedRow map[string]any
						err            error
					)
					if len(*customProjections) > 0 {
						transformedRow, err = o.projectRow(&baseRow, customProjections, alias, ctx)
					} else {
						transformedRow, err = o.projectRow(&baseRow, &currModel.Projections, alias, ctx)
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
	row *map[string]any, projections *[]orm_schema.Projection, alias string, ctx context.Context,
) (map[string]any, error) {
	safeRow := sync.Map{}
	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	for _, projection := range *projections {
		wg.Add(1)
		proj := projection
		go func(p *orm_schema.Projection) {
			defer wg.Done()
			var accessor string
			if o.Operation == common.OrmSelect {
				accessor = fmt.Sprintf("%s.%s", alias, p.Column)
			} else {
				accessor = p.Column
			}
			colVal, exists := (*row)[accessor]
			if sanitized, err := o.sanitize(colVal, exists, p); err != nil {
				cancel(err)
			} else {
				select {
				case <-ctx.Done():
					return
				default:
					safeRow.Store(p.As, sanitized)
				}
			}
		}(&proj)
	}
	wg.Wait()

	if err := context.Cause(ctx); err != nil {
		return nil, err
	}

	projectedRow := make(map[string]any, len(*projections))
	safeRow.Range(func(key, value any) bool {
		projectedRow[fmt.Sprint(key)] = value
		return true
	})

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

func (o *orm) sanitize(val any, exists bool, projection *orm_schema.Projection) (any, error) {
	switch projection.ModelType {
	case common.DatabaseTypeString:
		if exists && val != nil {
			return fmt.Sprint(val), nil
		} else if projection.NotNull {
			return "", nil
		} else {
			return nil, nil
		}
	case common.DatabaseTypeBoolean:
		if exists && projection.ModelType == projection.SchemaType {
			return val, nil
		} else if exists {
			if val, err := strconv.ParseBool(fmt.Sprint(val)); err != nil {
				return nil, fmt.Errorf("cast to %s failed: %s", projection.ModelType, err)
			} else {
				return val, nil
			}
		} else {
			return false, nil
		}

	case common.DatabaseTypeNumber:
		if exists && val != nil && projection.ModelType == projection.SchemaType {
			return val, nil
		} else if exists && val != nil && projection.SchemaType == common.DatabaseTypeString {
			if val, err := strconv.ParseFloat(fmt.Sprint(val), 64); err != nil {
				return nil, fmt.Errorf("cast to %s failed: %s", projection.ModelType, err)
			} else {
				return val, nil
			}
		} else if exists && val != nil && projection.SchemaType == common.DatabaseTypeBoolean {
			if val, err := strconv.ParseBool(fmt.Sprint(val)); err != nil {
				return nil, fmt.Errorf("cast to %s failed: %s", projection.ModelType, err)
			} else {
				return val, nil
			}
		} else if projection.NotNull {
			return 0, nil
		} else {
			return nil, nil
		}
	default:
		return val, nil
	}
}
