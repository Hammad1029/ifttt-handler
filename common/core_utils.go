package common

import (
	"context"
	"sync"
)

var evaluators map[string]func(a, b any) bool = map[string]func(a, b any) bool{
	"eq": func(a, b any) bool {
		return EqualityCheck(a, b)
	},
	"ne": func(a, b any) bool {
		return !EqualityCheck(a, b)
	},
	"in": func(a, b any) bool {
		return ArrayIncludes(a, b)
	},
	"notIn": func(a, b any) bool {
		return !ArrayIncludes(a, b)
	},
	"lt": func(a, b any) bool {
		return evaluateFloats(a, b, func(aFloat, bFloat float64) any {
			return aFloat < bFloat
		}).(bool)
	},
	"lte": func(a, b any) bool {
		return evaluateFloats(a, b, func(aFloat, bFloat float64) any {
			return aFloat <= bFloat
		}).(bool)
	},
	"gt": func(a, b any) bool {
		return evaluateFloats(a, b, func(aFloat, bFloat float64) any {
			return aFloat > bFloat
		}).(bool)
	},
	"gte": func(a, b any) bool {
		return evaluateFloats(a, b, func(aFloat, bFloat float64) any {
			return aFloat >= bFloat
		}).(bool)
	},
}

func GetEvaluator(operator string) *func(a, b any) bool {
	if ev, ok := evaluators[operator]; ok {
		return &ev
	} else {
		return nil
	}
}

var arithmeticOperators map[string]func(a, b any) any = map[string]func(a, b any) any{
	"+": func(a, b any) any {
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
	"-": func(a, b any) any {
		return evaluateFloats(a, b, func(aFloat, bFloat float64) any {
			return aFloat - bFloat
		}).(float64)
	},
	"*": func(a, b any) any {
		return evaluateFloats(a, b, func(aFloat, bFloat float64) any {
			return aFloat * bFloat
		}).(float64)
	},
	"/": func(a, b any) any {
		return evaluateFloats(a, b, func(aFloat, bFloat float64) any {
			return aFloat / bFloat
		}).(float64)
	},
	"%": func(a, b any) any {
		return evaluateFloats(a, b, func(aFloat, bFloat float64) any {
			return int(aFloat) % int(bFloat)
		}).(int)
	},
}

func GetArithmeticOperator(operator string) *func(a, b any) any {
	if op, ok := arithmeticOperators[operator]; ok {
		return &op
	} else {
		return nil
	}
}

func GetCtxState(ctx context.Context) *sync.Map {
	return ctx.Value(ContextState).(*sync.Map)
}

func GetResponseSent(ctx context.Context) bool {
	requestState := GetCtxState(ctx)
	if v, ok := requestState.Load(ContextResponseSent); ok {
		return v.(bool)
	}
	return false
}

func SetResponseSent(ctx context.Context) bool {
	requestState := GetCtxState(ctx)
	v, ok := requestState.Load(ContextResponseSent)
	if !ok {
		return false
	}

	if v.(bool) {
		return false
	}
	requestState.Store(ContextResponseSent, true)
	return true
}
