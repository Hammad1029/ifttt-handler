package resolvable

import (
	"context"
	"fmt"
	"ifttt/handler/common"
	"ifttt/handler/domain/orm_schema"
	"sync"

	"github.com/samber/lo"
)

type ormResolvable struct {
	Query              *queryResolvable      `json:"query" mapstructure:"query"`
	Operation          string                `json:"operation" mapstructure:"operation"`
	Model              string                `json:"model" mapstructure:"model"`
	ConditionsTemplate string                `json:"conditionsTemplate" mapstructure:"conditionsTemplate"`
	ConditionsValue    []any                 `json:"conditionsValue" mapstructure:"conditionsValue"`
	Populate           []orm_schema.Populate `json:"populate" mapstructure:"populate"`
}

func (o *ormResolvable) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	ormRepo, ok := dependencies[common.DependencyOrmCacheRepo].(orm_schema.CacheRepository)
	if !ok {
		return nil, fmt.Errorf("could not cast orm repo")
	}

	if o.Query == nil {
		return nil, fmt.Errorf("query is null")
	}

	model, err := ormRepo.GetModel(o.Model, ctx)
	if err != nil {
		return nil, err
	}

	switch o.Operation {
	case common.OrmSelect:
		if results, err := o.Query.Resolve(ctx, dependencies); err != nil {
			return nil, err
		} else if queryData, ok := results.(*queryData); !ok {
			return nil, fmt.Errorf("could not cast results to query data")
		} else if transformed, err := o.transformResults(
			queryData.Results, model.Name, model, &o.Populate, ormRepo, ctx); err != nil {
			return nil, err
		} else {
			queryData.Results = &transformed
			return queryData, nil
		}
	default:
		return nil, fmt.Errorf("unsupported operation: %s", o.Operation)
	}
}

func (o *ormResolvable) transformResults(
	rawResults *[]map[string]any,
	alias string,
	currModel *orm_schema.Model,
	populate *[]orm_schema.Populate,
	ormRepo orm_schema.CacheRepository,
	ctx context.Context,
) ([]map[string]any, error) {
	pKeyAccessor := fmt.Sprintf("%s.%s", alias, currModel.PrimaryKey)
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
			select {
			case <-cancelCtx.Done():
				return
			default:
				{
					baseRow := rowGroup[0]
					transformedRow := o.projectRow(&baseRow, &currModel.Projections, alias)
					for _, p := range *populate {
						childModel, err := ormRepo.GetModel(p.Model, ctx)
						if err != nil || childModel == nil {
							cancel(err)
							return
						}
						childAlias := fmt.Sprintf("%s_%s", alias, p.As)
						childGroups, err := o.transformResults(&rowGroup, childAlias, childModel, &p.Populate, ormRepo, ctx)
						if err != nil {
							cancel(fmt.Errorf("could not transform alias %s for model %s", p.As, p.Model))
							return
						}
						transformedRow[p.As] = childGroups
					}
					transformed.Store(pKey, transformedRow)
				}
			}
			wg.Done()
		}(k, g)
	}
	wg.Wait()

	flattened := make([]map[string]any, 0, len(grouped))
	transformed.Range(func(_, value any) bool {
		if mapped, ok := value.(map[string]any); ok {
			flattened = append(flattened, mapped)
			return true
		} else {
			cancel(fmt.Errorf("could not cast sync.Map value to map[string]any"))
			return false
		}
	})

	if err := context.Cause(cancelCtx); err != nil {
		return nil, err
	}
	return flattened, nil
}

func (o *ormResolvable) projectRow(
	row *map[string]any, projections *[]orm_schema.Projection, alias string,
) map[string]any {
	projectedRow := map[string]any{}
	for _, p := range *projections {
		accessor := fmt.Sprintf("%s.%s", alias, p.Column)
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
