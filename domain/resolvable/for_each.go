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
	Async bool          `json:"async" mapstructure:"async"`
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
	} else if inputResolved == nil {
		return nil, nil
	}

	inputIndirect := reflect.Indirect(reflect.ValueOf(inputResolved))
	inputInterface := inputIndirect.Interface()
	if inputInterface == nil {
		return nil, nil
	}

	var inArr reflect.Value
	switch inputIndirect.Kind() {
	case reflect.Slice, reflect.Array:
		inArr = inputIndirect
	case reflect.Int:
		length, ok := (inputInterface).(int)
		if !ok {
			return nil, fmt.Errorf("could not cast range to number")
		}
		newArr := make([]int, length)
		for idx := range newArr {
			newArr[idx] = idx
		}
		inArr = reflect.ValueOf(newArr)
	default:
		return nil, fmt.Errorf("foreach input needs int or array")
	}

	cancelCtx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	forEachFunc := func(idx int) error {
		select {
		case <-cancelCtx.Done():
			return nil
		default:
			ctx = context.WithValue(ctx, common.ContextIter,
				forEachElement{Element: inArr.Index(idx).Interface(), Index: idx})
			if _, err := ResolveArrayMust(f.Do, ctx, dependencies); err != nil {
				cancel(err)
				return err
			}
		}
		return nil
	}

	if f.Async {
		var wg sync.WaitGroup
		for i := 0; i < inputIndirect.Len(); i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				forEachFunc(idx)
			}(i)
		}
		wg.Wait()
	} else {
		for i := 0; i < inputIndirect.Len(); i++ {
			if err := forEachFunc(i); err != nil {
				break
			}
		}
	}

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
