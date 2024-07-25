package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/itchyny/gojq"
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

func toFloat64(v interface{}) (float64, bool) {
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

func toString(v interface{}) string {
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

func evaluateFloats(a, b interface{}, evaluate func(float64, float64) interface{}) interface{} {
	if af, aOk := toFloat64(a); aOk {
		if bf, bOk := toFloat64(b); bOk {
			return evaluate(af, bf)
		}
	}
	return nil
}

func EqualityCheck(a, b interface{}) bool {
	return fmt.Sprint(a) == fmt.Sprint(b)
}

func ArrayIncludes(a, b interface{}) bool {
	var arr []interface{}
	switch a := a.(type) {
	case []interface{}:
		arr = a
	default:
		arr = []interface{}{a}
	}
	return lo.ContainsBy(arr, func(x interface{}) bool {
		return EqualityCheck(x, b)
	})
}

func RunJQQuery(queryString string, input interface{}) (interface{}, error) {
	query, err := gojq.Parse(queryString)
	if err != nil {
		return nil, fmt.Errorf("method runJQQuery: could not parse gojq query: %s", err)
	}

	var resultVals []interface{}
	resultIter := query.Run(input)

	for {
		v, ok := resultIter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			if err, ok := err.(*gojq.HaltError); ok && err.Value() == nil {
				break
			}
			return nil, fmt.Errorf("method runJQQuery: error in running gojq iter: %s", err)
		}
		resultVals = append(resultVals, v)
	}

	switch len(resultVals) {
	case 0:
		return nil, nil
	case 1:
		return resultVals[0], nil
	default:
		return resultVals, nil
	}
}
