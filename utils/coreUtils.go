package utils

func GetEvaluators() map[string]func(a, b interface{}) bool {
	return map[string]func(a, b interface{}) bool{
		"eq": func(a, b interface{}) bool {
			if aStr, aOk := a.(string); aOk {
				if bStr, bOk := b.(string); bOk {
					return aStr == bStr
				}
			}

			if af, aOk := toFloat64(a); aOk {
				if bf, bOk := toFloat64(b); bOk {
					return af == bf
				}
			}
			return false
		},
		"ne": func(a, b interface{}) bool {

			if aStr, aOk := a.(string); aOk {
				if bStr, bOk := b.(string); bOk {
					return aStr != bStr
				}
			}

			if af, aOk := toFloat64(a); aOk {
				if bf, bOk := toFloat64(b); bOk {
					return af != bf
				}
			}
			return false
		},
		"lt": func(a, b interface{}) bool {
			if af, aOk := toFloat64(a); aOk {
				if bf, bOk := toFloat64(b); bOk {
					return af < bf
				}
			}
			return false
		},
		"lte": func(a, b interface{}) bool {
			if af, aOk := toFloat64(a); aOk {
				if bf, bOk := toFloat64(b); bOk {
					return af <= bf
				}
			}
			return false
		},
		"gt": func(a, b interface{}) bool {
			if af, aOk := toFloat64(a); aOk {
				if bf, bOk := toFloat64(b); bOk {
					return af > bf
				}
			}
			return false
		},
		"gte": func(a, b interface{}) bool {
			if af, aOk := toFloat64(a); aOk {
				if bf, bOk := toFloat64(b); bOk {
					return af >= bf
				}
			}
			return false
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
			floatA, okA := toFloat64(a)
			floatB, okB := toFloat64(b)
			if okA && okB {
				return floatA - floatB
			}
			return nil
		},
		"*": func(a, b interface{}) interface{} {
			floatA, okA := toFloat64(a)
			floatB, okB := toFloat64(b)
			if okA && okB {
				return floatA * floatB
			}
			return nil
		},
		"/": func(a, b interface{}) interface{} {
			floatA, okA := toFloat64(a)
			floatB, okB := toFloat64(b)
			if okA && okB {
				return floatA / floatB
			}
			return nil
		},
		"%": func(a, b interface{}) interface{} {
			intA, okA := toInt(a)
			intB, okB := toInt(b)
			if okA && okB {
				return intA % intB
			}
			return nil
		},
	}

}
