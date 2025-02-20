package resolvable

import (
	"context"
	"fmt"
	"ifttt/handler/common"
	"strconv"

	"github.com/nleeper/goment"
)

type dateFunc struct {
	Input        dateInput        `json:"input" mapstructure:"input"`
	Manipulators []dateArithmetic `json:"manipulators" mapstructure:"manipulators"`
	Format       string           `json:"format" mapstructure:"format"`
	UTC          bool             `json:"utc" mapstructure:"utc"`
}

type dateArithmetic struct {
	Operator string     `json:"operator" mapstructure:"operator"`
	Operand  Resolvable `json:"operand" mapstructure:"operand"`
	Unit     string     `json:"unit" mapstructure:"unit"`
}

type dateArithmeticFunc func(input *goment.Goment) (*goment.Goment, error)

type dateInput struct {
	Input    *Resolvable `json:"input" mapstructure:"input"`
	Parse    string      `json:"parse" mapstructure:"parse"`
	Timezone string      `json:"timezone" mapstructure:"timezone"`
}

type dateIntervals struct {
	Start  dateInput `json:"start" mapstructure:"start"`
	End    dateInput `json:"end" mapstructure:"end"`
	Unit   string    `json:"unit" mapstructure:"unit"`
	Format string    `json:"format" mapstructure:"format"`
}

func (d *dateFunc) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	dateInput, err := d.Input.get(ctx, dependencies)
	if err != nil {
		return nil, err
	}

	for _, m := range d.Manipulators {
		if mFunc, err := m.getFunc(ctx, dependencies); err != nil {
			return nil, err
		} else if manipulated, err := mFunc(dateInput); err != nil {
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

func (d *dateArithmetic) getFunc(ctx context.Context, dependencies map[common.IntIota]any) (dateArithmeticFunc, error) {
	operand, err := d.Operand.Resolve(ctx, dependencies)
	if err != nil {
		return nil, err
	} else if d.Operator != common.DateOperatorAdd && d.Operator != common.DateOperatorSubtract {
		return nil, fmt.Errorf("date manipulation operator %s not found", d.Operator)
	}

	operandNumber, ok := operand.(int)
	if !ok {
		if opN, err := strconv.Atoi(fmt.Sprint(operand)); err != nil {
			return nil, fmt.Errorf("failed to convert operand to number")
		} else {
			operandNumber = opN
		}
	}

	return func(input *goment.Goment) (*goment.Goment, error) {
		if d.Operator == common.DateOperatorSubtract {
			return input.Subtract(operandNumber, d.Unit), nil
		} else {
			return input.Add(operandNumber, d.Unit), nil
		}
	}, nil
}

func (d *dateInput) get(ctx context.Context, dependencies map[common.IntIota]any) (*goment.Goment, error) {
	var gomentDate *goment.Goment

	if d.Input == nil {
		if newDate, err := goment.New(); err != nil {
			return nil, err
		} else {
			gomentDate = newDate
		}
	} else {
		newDate, err := d.Input.Resolve(ctx, dependencies)
		if err != nil {
			return nil, err
		}

		if d.Parse == "" {
			d.Parse = common.DateTimeFormatGeneric
		}

		if strDate := fmt.Sprint(newDate); len(d.Parse) > len(strDate) {
			return nil, fmt.Errorf("parser length greater than date")
		} else if dateParsed, err := goment.New(strDate, d.Parse); err != nil {
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

func (d *dateIntervals) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	start, err := d.Start.get(ctx, dependencies)
	if err != nil {
		return nil, err
	}

	end, err := d.End.get(ctx, dependencies)
	if err != nil {
		return nil, err
	}

	formatter := d.Format
	if formatter == "" {
		formatter = common.DateTimeFormatGeneric
	}

	var dates []string
	for start.IsBefore(end, d.Unit) || start.IsSame(end, d.Unit) {
		dates = append(dates, start.Format(formatter))
		start.Add(1, d.Unit)
	}

	return dates, nil
}
