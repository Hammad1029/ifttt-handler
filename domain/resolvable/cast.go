package resolvable

import (
	"context"
	"encoding/json"
	"fmt"
	"ifttt/handler/common"
	"strconv"
)

type cast struct {
	Input any    `json:"input" mapstructure:"input"`
	To    string `json:"to" mapstructure:"to"`
}

func (c *cast) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	if c.To != common.CastToString && c.To != common.CastToNumber && c.To != common.CastToBoolean {
		return nil, fmt.Errorf("cast to %s failed: datatype not found", c.To)
	}
	resolved, err := resolveMaybe(c.Input, ctx, dependencies)
	if err != nil {
		return nil, err
	}
	switch c.To {
	case common.CastToString:
		if marshalled, err := json.Marshal(&resolved); err != nil {
			return nil, fmt.Errorf("cast to %s failed: %s", c.To, err)
		} else {
			return string(marshalled), nil
		}
	case common.CastToNumber:
		if val, err := strconv.ParseFloat(fmt.Sprint(resolved), 64); err != nil {
			return nil, fmt.Errorf("cast to %s failed: %s", c.To, err)
		} else {
			return val, nil
		}
	case common.CastToBoolean:
		if val, err := strconv.ParseBool(fmt.Sprint(resolved)); err != nil {
			return nil, fmt.Errorf("cast to %s failed: %s", c.To, err)
		} else {
			return val, nil
		}
	}
	return nil, nil
}
