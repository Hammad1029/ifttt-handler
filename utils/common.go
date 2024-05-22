package utils

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

func BenchmarkFn(fn func()) {
	start := time.Now()
	fn()
	fmt.Printf("execution time: %+v\n", time.Since(start))
}

func JsonStringToMap(str string) (map[string]interface{}, error) {
	if str == "" {
		str = "{}"
	}
	newMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(str), &newMap)
	return newMap, err
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

func SerializeMap(input map[string]interface{}) (map[string]interface{}, error) {
	output := make(map[string]interface{})
	for key, value := range input {
		switch v := value.(type) {
		case map[string]interface{}:
			serialized, err := SerializeMap(v)
			if err != nil {
				return nil, err
			}
			nestedMap, err := json.Marshal(serialized)
			if err != nil {
				return nil, err
			}
			output[key] = string(nestedMap)
		default:
			output[key] = value
		}
	}
	return output, nil
}
