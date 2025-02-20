package resolvable

import (
	"context"
	"fmt"
	"ifttt/handler/common"
	"reflect"
	"sort"
	"sync"
)

type filterMap struct {
	Input     any           `json:"input" mapstructure:"input"`
	Do        *[]Resolvable `json:"do" mapstructure:"do"`
	Condition Condition     `json:"condition" mapstructure:"condition"`
	Async     bool          `json:"async" mapstructure:"async"`
}

type iterElement struct {
	Element any `json:"element" mapstructure:"element"`
	Index   int `json:"index" mapstructure:"index"`
}

type getIter struct {
	Index bool `json:"index" mapstructure:"index"`
}

func (f *filterMap) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	inputResolved, err := resolveMaybe(f.Input, ctx, dependencies)
	if err != nil {
		return nil, err
	} else if inputResolved == nil {
		return nil, nil
	}

	inputIndirect := reflect.Indirect(reflect.ValueOf(inputResolved))
	inputInterface := inputIndirect.Interface()
	if inputInterface == nil {
		return nil, nil
	}
	inputKind := inputIndirect.Kind()

	var inArr reflect.Value
	switch {
	case inputKind == reflect.Slice || inputKind == reflect.Array:
		inArr = inputIndirect
	case inputIndirect.CanInt():
		inArr = reflect.ValueOf(f.formNumberArray(int(inputIndirect.Int())))
	case inputIndirect.CanUint():
		inArr = reflect.ValueOf(f.formNumberArray(int(inputIndirect.Uint())))
	case inputIndirect.CanFloat():
		inArr = reflect.ValueOf(f.formNumberArray(int(inputIndirect.Float())))
	default:
		return nil, fmt.Errorf("foreach input needs int or array")
	}

	if output, err := f.runIteration(&inArr, ctx, dependencies); err != nil {
		return nil, err
	} else {
		return output, nil
	}
}

func (g *getIter) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	if element := ctx.Value(common.ContextIter); element == nil {
		return nil, fmt.Errorf("iter not available")
	} else if elementCasted, ok := element.(*iterElement); !ok {
		return nil, fmt.Errorf("iter could not be casted")
	} else if g.Index {
		return elementCasted.Index, nil
	} else {
		return elementCasted.Element, nil
	}
}

func (f *filterMap) formNumberArray(length int) []int {
	newArr := make([]int, length)
	for idx := range newArr {
		newArr[idx] = idx
	}
	return newArr
}

func (f *filterMap) runIteration(
	inputArr *reflect.Value, ctx context.Context, dependencies map[common.IntIota]any,
) ([]any, error) {
	cancelCtx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	outputArr := make([]any, 0, inputArr.Len())

	if f.Async {
		var wg sync.WaitGroup
		var outputMap sync.Map
		for idx := 0; idx < inputArr.Len(); idx++ {
			wg.Add(1)
			el := iterElement{Element: inputArr.Index(idx).Interface(), Index: idx}
			go func(fEl *iterElement) {
				defer wg.Done()
				if val, ev, err := f.singleIteration(fEl, cancelCtx, cancel, ctx, dependencies); err != nil {
					cancel(err)
					return
				} else if ev {
					outputMap.Store(fEl.Index, val)
				}
			}(&el)
		}
		wg.Wait()

		keys := make([]int, 0, inputArr.Len())
		outputMap.Range(func(key, value any) bool {
			keys = append(keys, key.(int))
			return true
		})
		sort.Ints(keys)
		for _, key := range keys {
			if val, ok := outputMap.Load(key); ok {
				outputArr = append(outputArr, val)
			} else {
				cancel(fmt.Errorf("key %d not found in outputMap", key))
				break
			}
		}
	} else {
		for idx := 0; idx < inputArr.Len(); idx++ {
			el := iterElement{Element: inputArr.Index(idx).Interface(), Index: idx}
			if val, ev, err := f.singleIteration(&el, cancelCtx, cancel, ctx, dependencies); err != nil {
				cancel(err)
				break
			} else if ev {
				outputArr = append(outputArr, val)
			}
		}
	}

	if err := context.Cause(cancelCtx); err != nil {
		return nil, err
	}
	return outputArr, nil
}

func (f *filterMap) singleIteration(
	element *iterElement,
	cancelCtx context.Context, cancelFunc context.CancelCauseFunc,
	ctx context.Context, dependencies map[common.IntIota]any,
) (any, bool, error) {
	select {
	case <-cancelCtx.Done():
		return nil, false, nil
	default:
		ctx = context.WithValue(ctx, common.ContextIter, element)
	}

	if ev, err := f.Condition.EvaluateGroup(ctx, dependencies); err != nil {
		cancelFunc(err)
		return nil, false, err
	} else if ev {
		select {
		case <-cancelCtx.Done():
			return nil, false, nil
		default:
		}

		if _, err := ResolveArrayMust(f.Do, ctx, dependencies); err != nil {
			cancelFunc(err)
			return nil, false, err
		}
		return element.Element, true, nil
	}

	return nil, false, nil
}
