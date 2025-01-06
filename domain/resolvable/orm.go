package resolvable

import (
	"context"
	"fmt"
	"ifttt/handler/common"
	"ifttt/handler/domain/orm_schema"
	"sort"
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
	grouped := lo.GroupBy(*rawResults, func(row map[string]any) string {
		return fmt.Sprint(row[pKeyAccessor])
	})
	delete(grouped, fmt.Sprint(nil))

	transformed := sync.Map{}
	wg := sync.WaitGroup{}
	cancelCtx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	for k, g := range grouped {
		wg.Add(1)
		go func(pKey string, rowGroup []map[string]any) {
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
							cancel(fmt.Errorf("could not transform alias %s for model %s", p.As, p.Model))
							return
						}
						if (association.Type == common.AssociationsHasOne ||
							association.Type == common.AssociationsBelongsTo) && len(childGroups) > 0 {
							transformedRow[p.As] = childGroups[0]
						} else {
							transformedRow[p.As] = childGroups
						}
					}
					transformed.Store(pKey, transformedRow)
				}
			}
		}(k, g)
	}
	wg.Wait()

	pKeysString := []string{}
	transformed.Range(func(key, _ any) bool {
		pKeysString = append(pKeysString, fmt.Sprint(key))
		return true
	})

	pKeysSorted := make([]int, len(pKeysString))
	for i, str := range pKeysString {
		if num, err := strconv.Atoi(str); err != nil {
			return nil, fmt.Errorf("error converting pkey to int: %v", err)
		} else {
			pKeysSorted[i] = num
		}
	}
	sort.Ints(pKeysSorted)

	flattened := make([]map[string]any, 0, len(grouped))
	for _, k := range pKeysSorted {
		value, _ := transformed.Load(fmt.Sprint(k))
		if mapped, ok := value.(map[string]any); ok {
			flattened = append(flattened, mapped)
		} else {
			cancel(fmt.Errorf("could not cast sync.Map value to map[string]any"))
			break
		}
	}

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
		if colVal, ok := (*row)[accessor]; ok && colVal != nil {
			projectedRow[p.As] = colVal
		} else if p.DataType == common.DatabaseTypeString {
			projectedRow[p.As] = new(string)
		} else if p.DataType == common.DatabaseTypeNumber {
			projectedRow[p.As] = new(int)
		} else if p.DataType == common.DatabaseTypeBoolean {
			projectedRow[p.As] = new(bool)
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
