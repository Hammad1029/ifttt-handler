package common

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/nleeper/goment"
	"github.com/samber/lo"
	"golang.org/x/crypto/bcrypt"
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

func evaluateFloats(a, b any, evaluate floatComparator) (bool, error) {
	if af, aOk := toFloat64(a); aOk {
		if bf, bOk := toFloat64(b); bOk {
			return evaluate(af, bf), nil
		}
	}
	return false, fmt.Errorf("could not convert values to floats")
}

func calculateFloats(a, b any, evaluate floatCalculator) (float64, error) {
	if af, aOk := toFloat64(a); aOk {
		if bf, bOk := toFloat64(b); bOk {
			return evaluate(af, bf), nil
		}
	}
	return 0, fmt.Errorf("could not convert values to floats")
}

func evaluateDates(a, b any, evaluate func(dt1, dt2 *goment.Goment) bool) (bool, error) {
	if date1, err := goment.New(a); err != nil {
		return false, fmt.Errorf("could not parse date1")
	} else if date2, err := goment.New(b); err != nil {
		return false, fmt.Errorf("could not parse date2")
	} else {
		return evaluate(date1, date2), nil
	}
}

func EqualityCheck(a, b any) bool {
	return fmt.Sprint(a) == fmt.Sprint(b)
}

func arrayIncludes(is, in any) bool {
	var arr []any
	switch in := in.(type) {
	case []any:
		arr = in
	default:
		arr = []any{in}
	}
	return lo.ContainsBy(arr, func(x any) bool {
		return EqualityCheck(x, is)
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

func compareBcrypt(a, b any) (bool, error) {
	aArr := []byte(fmt.Sprint(a))
	bArr := []byte(fmt.Sprint(b))
	if err := bcrypt.CompareHashAndPassword(aArr, bArr); err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return false, nil
		}
		return false, err
	} else {
		return true, nil
	}
}
