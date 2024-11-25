package requestvalidator

import (
	"fmt"
	"regexp"
	"strconv"
	"sync"

	"github.com/mitchellh/mapstructure"
	"github.com/samber/lo"
)

func ValidateMap(schema *map[string]RequestParameter, m *map[string]any) []ValidationError {
	var wg sync.WaitGroup
	validationErrors := []ValidationError{}

	absentFromRequest, absentFromSchema := lo.Difference(lo.Keys(*schema), lo.Keys(*m))
	for _, key := range absentFromRequest {
		validationErrors = append(validationErrors, ValidationError{
			ErrorInfo: fmt.Errorf("key %s missing in request", key),
		})
	}
	for _, key := range absentFromSchema {
		validationErrors = append(validationErrors, ValidationError{
			ErrorInfo: fmt.Errorf("key %s not part of schema", key),
		})
	}
	if len(validationErrors) != 0 {
		return validationErrors
	}

	for key, val := range *m {
		wg.Add(1)
		go func(key string, val any) {
			defer wg.Done()
			if s, ok := (*schema)[key]; ok {
				if err := s.validateValue(val); err != nil {
					for _, err := range err {
						validationErrors = append(validationErrors, ValidationError{
							ErrorInfo: fmt.Errorf("validation for key %s failed: %s", key, err.ErrorInfo),
						})
					}
				}
			}
		}(key, val)
	}

	wg.Wait()
	return validationErrors
}

func (s *RequestParameter) validateValue(val any) []ValidationError {
	if val == nil {
		if s.Required {
			return []ValidationError{{ErrorInfo: fmt.Errorf("key is requires")}}
		}
		return nil
	}

	switch s.DataType {
	case dataTypeText, dataTypeNumber, dataTypeBoolean:
		switch val.(type) {
		case string:
			if s.DataType != dataTypeText {
				return []ValidationError{{ErrorInfo: fmt.Errorf("invalid datatype, requires %s", s.DataType)}}
			}
		case uint, int, float32, float64:
			if s.DataType != dataTypeNumber {
				return []ValidationError{{ErrorInfo: fmt.Errorf("invalid datatype, requires %s", s.DataType)}}
			}
		case bool:
			if s.DataType != dataTypeBoolean {
				return []ValidationError{{ErrorInfo: fmt.Errorf("invalid datatype, requires %s", s.DataType)}}
			}
		default:
			return []ValidationError{{ErrorInfo: fmt.Errorf("invalid datatype: unambigous")}}
		}
		validationRegex := regexp.MustCompile(s.Regex)
		stringifiedVal := fmt.Sprint(val)
		matches := validationRegex.FindStringSubmatch(stringifiedVal)
		if len(matches) == 0 || matches[0] != stringifiedVal {
			return []ValidationError{{ErrorInfo: fmt.Errorf("regex validation failed")}}
		}
		if s.DataType == dataTypeNumber {
			numVal, err := strconv.ParseFloat(stringifiedVal, 64)
			if err != nil {
				return []ValidationError{
					{ErrorInfo: fmt.Errorf("could not convert number to float64"), Internal: true}}
			}
			validator := numberValue{}
			if err := mapstructure.Decode(s.Config, &validator); err != nil {
				return []ValidationError{
					{ErrorInfo: fmt.Errorf("could not decode validator"), Internal: true}}
			} else if float64(validator.Minimum) > numVal || float64(validator.Maximum) < numVal {
				return []ValidationError{
					{ErrorInfo: fmt.Errorf("number not within range %d - %d", validator.Minimum, validator.Maximum)}}
			}
		}
	case dataTypeArray:
		arr, ok := val.([]any)
		if !ok {
			return []ValidationError{{ErrorInfo: fmt.Errorf("invalid datatype, requires %s", s.DataType)}}
		}
		validator := arrayValue{}
		if err := mapstructure.Decode(s.Config, &validator); err != nil {
			return []ValidationError{
				{ErrorInfo: fmt.Errorf("could not decode validator"), Internal: true}}
		}
		return validator.validate(arr)
	case dataTypeMap:
		mapVal, ok := val.(map[string]any)
		if !ok {
			return []ValidationError{{ErrorInfo: fmt.Errorf("invalid datatype, requires %s", s.DataType)}}
		}
		validator := mapValue{}
		if err := mapstructure.Decode(s.Config, &validator); err != nil {
			return []ValidationError{
				{ErrorInfo: fmt.Errorf("could not decode validator"), Internal: true}}
		}
		return ValidateMap((*map[string]RequestParameter)(&validator), &mapVal)
	}
	return nil
}

func (s *arrayValue) validate(arr []any) []ValidationError {
	var wg sync.WaitGroup
	validationErrors := []ValidationError{}

	if len(arr) < s.Minimum || len(arr) > s.Maximum {
		return []ValidationError{{ErrorInfo: fmt.Errorf("invalid length of array")}}
	}

	for i, item := range arr {
		wg.Add(1)
		go func(i int, item any) {
			defer wg.Done()
			if err := s.OfType.validateValue(item); err != nil {
				validationErrors = append(validationErrors, err...)
			}
		}(i, item)
	}

	wg.Wait()
	return validationErrors
}
