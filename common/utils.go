package common

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/samber/lo"
)

func BenchmarkFn(fn func()) {
	start := time.Now()
	fn()
	fmt.Printf("execution time: %+v\n", time.Since(start))
}

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func GetTimeSlot(t time.Time, slot int) time.Time {
	secondsSinceStartOfDay := t.Hour()*3600 + t.Minute()*60 + t.Second()
	fullSlots := secondsSinceStartOfDay / slot
	startOfLastFullSlot := fullSlots * slot
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).Add(time.Duration(startOfLastFullSlot) * time.Second)
}

func toFloat64(v any) (float64, bool) {
	switch v := v.(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f, true
		}
	}
	return 0, false
}

func toString(v any) string {
	switch v := v.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func evaluateFloats(a, b any, evaluate func(float64, float64) any) any {
	if af, aOk := toFloat64(a); aOk {
		if bf, bOk := toFloat64(b); bOk {
			return evaluate(af, bf)
		}
	}
	return nil
}

func EqualityCheck(a, b any) bool {
	return fmt.Sprint(a) == fmt.Sprint(b)
}

func ArrayIncludes(a, b any) bool {
	var arr []any
	switch a := a.(type) {
	case []any:
		arr = a
	default:
		arr = []any{a}
	}
	return lo.ContainsBy(arr, func(x any) bool {
		return EqualityCheck(x, b)
	})
}

func RegexpArrayMatch(patterns []string, input string) (bool, error) {
	for _, p := range patterns {
		if matched, err := regexp.MatchString(p, input); err != nil {
			return false, fmt.Errorf("method RegexpArrayMatch: error in checking regexp match: %s", err)
		} else if matched {
			return true, nil
		}
	}
	return false, nil
}
