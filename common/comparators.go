package common

import (
	"fmt"

	"github.com/nleeper/goment"
)

func GetComparator(operator string) *anyComparator {
	if ev, ok := comparators[operator]; ok {
		return &ev
	} else {
		return nil
	}
}

var comparators map[string]anyComparator = map[string]anyComparator{
	ComparatorEquals:            compareEquals,
	ComparatorNotEquals:         compareNotEquals,
	ComparatorIn:                compareIn,
	ComparatorNotIn:             compareNotIn,
	ComparatorLessThan:          compareLessThan,
	ComparatorLessThanEquals:    compareLessThanEquals,
	ComparatorGreaterThan:       compareGreaterThan,
	ComparatorGreaterThanEquals: compareGreaterThanEquals,
}

func compareEquals(a, b any, comparisionType string) (bool, error) {
	switch comparisionType {
	case ComparisionTypeBcrypt:
		return compareBcrypt(a, b)
	default:
		return EqualityCheck(a, b), nil
	}
}

func compareNotEquals(a, b any, comparisionType string) (bool, error) {
	switch comparisionType {
	case ComparisionTypeBcrypt:
		if v, err := compareBcrypt(a, b); err != nil {
			return false, err
		} else {
			return !v, nil
		}
	default:
		return !EqualityCheck(a, b), nil
	}
}

func compareIn(a, b any, comparisionType string) (bool, error) {
	switch comparisionType {
	case ComparisionTypeString, ComparisionTypeBoolean, ComparisionTypeDate, ComparisionTypeNumber:
		return arrayIncludes(a, b), nil
	default:
		return false, fmt.Errorf("in comparator for %s not found", comparisionType)
	}
}

func compareNotIn(a, b any, comparisionType string) (bool, error) {
	switch comparisionType {
	case ComparisionTypeString, ComparisionTypeBoolean, ComparisionTypeDate, ComparisionTypeNumber:
		return !arrayIncludes(a, b), nil
	default:
		return false, fmt.Errorf("not in comparator for %s not found", comparisionType)
	}
}

func compareLessThan(a, b any, comparisionType string) (bool, error) {
	switch comparisionType {
	case ComparisionTypeString, ComparisionTypeBoolean, ComparisionTypeNumber:
		return evaluateFloats(a, b, func(aF, bF float64) bool {
			return aF < bF
		})
	case ComparisionTypeDate:
		return evaluateDates(a, b, func(dt1, dt2 *goment.Goment) bool {
			return dt1.IsBefore(dt2)
		})
	default:
		return false, fmt.Errorf("less than comparator for %s not found", comparisionType)
	}
}

func compareLessThanEquals(a, b any, comparisionType string) (bool, error) {
	switch comparisionType {
	case ComparisionTypeString, ComparisionTypeBoolean, ComparisionTypeNumber:
		return evaluateFloats(a, b, func(aF, bF float64) bool {
			return aF <= bF
		})
	case ComparisionTypeDate:
		return evaluateDates(a, b, func(dt1, dt2 *goment.Goment) bool {
			return dt1.IsSameOrBefore(dt2)
		})
	default:
		return false, fmt.Errorf("less than equals comparator for %s not found", comparisionType)
	}
}

func compareGreaterThan(a, b any, comparisionType string) (bool, error) {
	switch comparisionType {
	case ComparisionTypeString, ComparisionTypeBoolean, ComparisionTypeNumber:
		return evaluateFloats(a, b, func(aF, bF float64) bool {
			return aF > bF
		})
	case ComparisionTypeDate:
		return evaluateDates(a, b, func(dt1, dt2 *goment.Goment) bool {
			return dt1.IsAfter(dt2)
		})
	default:
		return false, fmt.Errorf("greater than comparator for %s not found", comparisionType)
	}
}

func compareGreaterThanEquals(a, b any, comparisionType string) (bool, error) {
	switch comparisionType {
	case ComparisionTypeString, ComparisionTypeBoolean, ComparisionTypeNumber:
		return evaluateFloats(a, b, func(aF, bF float64) bool {
			return aF >= bF
		})
	case ComparisionTypeDate:
		return evaluateDates(a, b, func(dt1, dt2 *goment.Goment) bool {
			return dt1.IsSameOrAfter(dt2)
		})
	default:
		return false, fmt.Errorf("greater than equals comparator for %s not found", comparisionType)
	}
}
