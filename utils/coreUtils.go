package utils

func GetEvaluators() map[string]func(a, b interface{}) bool {
	return map[string]func(a, b interface{}) bool{
		"eq": func(a, b interface{}) bool {
			return EqualityCheck(a, b)
		},
		"ne": func(a, b interface{}) bool {
			return !EqualityCheck(a, b)
		},
		"in": func(a, b interface{}) bool {
			return ArrayIncludes(a, b)
		},
		"notIn": func(a, b interface{}) bool {
			return !ArrayIncludes(a, b)
		},
		"lt": func(a, b interface{}) bool {
			return evaluateFloats(a, b, func(aFloat, bFloat float64) interface{} {
				return aFloat < bFloat
			}).(bool)
		},
		"lte": func(a, b interface{}) bool {
			return evaluateFloats(a, b, func(aFloat, bFloat float64) interface{} {
				return aFloat <= bFloat
			}).(bool)
		},
		"gt": func(a, b interface{}) bool {
			return evaluateFloats(a, b, func(aFloat, bFloat float64) interface{} {
				return aFloat > bFloat
			}).(bool)
		},
		"gte": func(a, b interface{}) bool {
			return evaluateFloats(a, b, func(aFloat, bFloat float64) interface{} {
				return aFloat >= bFloat
			}).(bool)
		},
	}
}
func GetArithmeticOperators() map[string]func(a, b interface{}) interface{} {
	return map[string]func(a, b interface{}) interface{}{
		"+": func(a, b interface{}) interface{} {
			switch a := a.(type) {
			case string:
				return a + toString(b)
			case float64:
				if strB, ok := b.(string); ok {
					if floatB, ok := toFloat64(strB); ok {
						return a + floatB
					}
				} else if floatB, ok := toFloat64(b); ok {
					return a + floatB
				}
			case int:
				if strB, ok := b.(string); ok {
					if floatB, ok := toFloat64(strB); ok {
						return float64(a) + floatB
					}
				} else if floatB, ok := toFloat64(b); ok {
					return float64(a) + floatB
				}
			default:
				return toString(a) + toString(b)
			}
			return toString(a) + toString(b)
		},
		"-": func(a, b interface{}) interface{} {
			return evaluateFloats(a, b, func(aFloat, bFloat float64) interface{} {
				return aFloat - bFloat
			}).(float64)
		},
		"*": func(a, b interface{}) interface{} {
			return evaluateFloats(a, b, func(aFloat, bFloat float64) interface{} {
				return aFloat * bFloat
			}).(float64)
		},
		"/": func(a, b interface{}) interface{} {
			return evaluateFloats(a, b, func(aFloat, bFloat float64) interface{} {
				return aFloat / bFloat
			}).(float64)
		},
		"%": func(a, b interface{}) interface{} {
			return evaluateFloats(a, b, func(aFloat, bFloat float64) interface{} {
				return int(aFloat) % int(bFloat)
			}).(int)
		},
	}

}
