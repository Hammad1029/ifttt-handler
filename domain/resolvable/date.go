package resolvable

import (
	"context"
	"fmt"
	"ifttt/handler/common"

	"github.com/nleeper/goment"
)

type dateFunc struct {
	Input        dateInput         `json:"input" mapstructure:"input"`
	Manipulators []dateManipulator `json:"manipulators" mapstructure:"manipulators"`
	Format       string            `json:"format" mapstructure:"format"`
	UTC          bool              `json:"utc" mapstructure:"utc"`
}

type dateManipulator struct {
	Operator string     `json:"operator" mapstructure:"operator"`
	Operand  Resolvable `json:"operand" mapstructure:"operand"`
	Unit     string     `json:"unit" mapstructure:"unit"`
}

type dateManipulatorFunc func(input *goment.Goment) (*goment.Goment, error)

type dateInput struct {
	Input    *Resolvable `json:"input" mapstructure:"input"`
	Parse    string      `json:"parse" mapstructure:"parse"`
	Timezone string      `json:"timezone" mapstructure:"timezone"`
}

func (d *dateFunc) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	var dateInput *goment.Goment
	if inputResolved, err := d.Input.Resolve(ctx, dependencies); err != nil {
		return nil, err
	} else if casted, ok := inputResolved.(*goment.Goment); !ok {
		return nil, fmt.Errorf("date input resolution did not return *goment.Goment")
	} else {
		dateInput = casted
	}

	for _, m := range d.Manipulators {
		if mFunc, err := m.Resolve(ctx, dependencies); err != nil {
			return nil, err
		} else if mFuncCasted, ok := mFunc.(dateManipulatorFunc); !ok {
			return nil, fmt.Errorf("could not cast dateManipulatorFunc")
		} else if manipulated, err := mFuncCasted(dateInput); err != nil {
			return nil, err
		} else {
			dateInput = manipulated
		}
	}

	var dateStr string
	if d.Format != "" {
		dateStr = dateInput.Format(d.Format)
	} else if d.UTC {
		dateStr = dateInput.UTC().ToISOString()
	} else {
		dateStr = dateInput.Format(common.DateTimeFormatGeneric)
	}

	return dateStr, nil
}

func (d *dateManipulator) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	operand, err := d.Operand.Resolve(ctx, dependencies)
	if err != nil {
		return nil, err
	}
	return func(input *goment.Goment) (*goment.Goment, error) {
		if d.Operator == common.DateOperatorAdd {
			return input.Add(operand, d.Unit), nil
		} else if d.Operator == common.DateOperatorSubtract {
			return input.Subtract(operand, d.Unit), nil
		} else {
			return nil, fmt.Errorf("date manipulation operator %s not found", d.Operator)
		}
	}, nil
}

func (d *dateInput) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	var gomentDate *goment.Goment

	if d.Input == nil {
		if newDate, err := goment.New(); err != nil {
			return nil, err
		} else {
			gomentDate = newDate
		}
	} else {
		if newDate, err := d.Input.Resolve(ctx, dependencies); err != nil {
			return nil, err
		} else if dateParsed, err := goment.New(fmt.Sprint(newDate), d.Parse); err != nil {
			return nil, err
		} else {
			gomentDate = dateParsed
		}
	}

	if d.Timezone != "" {
		if err := gomentDate.SetLocale(d.Timezone); err != nil {
			return nil, err
		}
	}

	return gomentDate, nil
}
