package common

type IntIota int

type anyComparator func(a, b any, comparsionType string) (bool, error)

type floatComparator func(a, b float64) bool

type anyCalculator func(any, any) (any, error)

type floatCalculator func(float64, float64) float64
