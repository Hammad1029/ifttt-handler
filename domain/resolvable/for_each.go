package resolvable

import (
	"context"
	"fmt"
	"ifttt/handler/common"
	"reflect"
	"sync"
)

type forEach struct {
	Input any           `json:"input" mapstructure:"input"`
	Do    *[]Resolvable `json:"do" mapstructure:"do"`
}

type forEachElement struct {
	Element any `json:"element" mapstructure:"element"`
	Index   int `json:"index" mapstructure:"index"`
}

type getIter struct {
}

func (f *forEach) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	inputResolved, err := resolveMaybe(f.Input, ctx, dependencies)
	if err != nil {
		return nil, err
	}

	inputIndirect := reflect.Indirect(reflect.ValueOf(inputResolved))
	if inputIndirect.Interface() == nil {
		return nil, nil
	} else if inputKind := inputIndirect.Kind(); inputKind != reflect.Slice && inputKind != reflect.Array {
		return nil, fmt.Errorf("array required as input for foreach")
	}

	var wg sync.WaitGroup
	cancelCtx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	for i := 0; i < inputIndirect.Len(); i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			select {
			case <-cancelCtx.Done():
				return
			default:
				ctx = context.WithValue(ctx, common.ContextIter,
					forEachElement{Element: inputIndirect.Index(idx).Interface(), Index: idx})
				if _, err := ResolveArrayMust(f.Do, ctx, dependencies); err != nil {
					cancel(err)
				}
			}
		}(i)
	}
	wg.Wait()

	if err := context.Cause(cancelCtx); err != nil {
		return nil, err
	}
	return nil, nil
}

func (g *getIter) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	if element := ctx.Value(common.ContextIter); element != nil {
		return element, nil
	}
	return nil, fmt.Errorf("iter not available")
}
