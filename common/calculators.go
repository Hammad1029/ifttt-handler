package common

func GetCalculator(operator string) *anyCalculator {
	if op, ok := calculators[operator]; ok {
		return &op
	} else {
		return nil
	}
}

var calculators map[string]anyCalculator = map[string]anyCalculator{
	CalculatorAdd: func(a, b any) (any, error) {
		switch a := a.(type) {
		case string:
			return a + toString(b), nil
		case float64:
			if strB, ok := b.(string); ok {
				if floatB, ok := toFloat64(strB); ok {
					return a + floatB, nil
				}
			} else if floatB, ok := toFloat64(b); ok {
				return a + floatB, nil
			}
		case int:
			if strB, ok := b.(string); ok {
				if floatB, ok := toFloat64(strB); ok {
					return float64(a) + floatB, nil
				}
			} else if floatB, ok := toFloat64(b); ok {
				return float64(a) + floatB, nil
			}
		default:
			return toString(a) + toString(b), nil
		}
		return toString(a) + toString(b), nil
	},
	CalculatorSubtract: func(a, b any) (any, error) {
		return calculateFloats(a, b, func(aF, bF float64) float64 {
			return aF - bF
		})
	},
	CalculatorMultiply: func(a, b any) (any, error) {
		return calculateFloats(a, b, func(aF, bF float64) float64 {
			return aF * bF
		})
	},
	CalculatorDivide: func(a, b any) (any, error) {
		return calculateFloats(a, b, func(aF, bF float64) float64 {
			return aF / bF
		})
	},
	CalculatorModulus: func(a, b any) (any, error) {
		return calculateFloats(a, b, func(aF, bF float64) float64 {
			return float64(uint(aF) % uint(bF))
		})
	},
}
