package requestvalidator

import (
	"fmt"
	"regexp"
	"strconv"
	"sync"

	"github.com/mitchellh/mapstructure"
	"github.com/samber/lo"
)

func ValidateMap(schema *map[string]RequestParameter, m *map[string]any,
	scan func(tagName string, value any) error) []ValidationError {
	var wg sync.WaitGroup
	validationErrors := []ValidationError{}

	absentFromRequest, _ := lo.Difference(lo.Keys(*schema), lo.Keys(*m))
	for _, key := range absentFromRequest {
		validationErrors = append(validationErrors, ValidationError{
			ErrorInfo: fmt.Errorf("key %s missing in request", key),
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
				if err := s.validateValue(val, scan); err != nil {
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

func (s *RequestParameter) validateValue(val any, scan func(tagName string, value any) error) []ValidationError {
	if val == nil {
		if s.Required {
			return []ValidationError{{ErrorInfo: fmt.Errorf("key is required")}}
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
		if err := scan(s.InternalTag, val); err != nil {
			return []ValidationError{{ErrorInfo: fmt.Errorf("could not scan request parameter: %s", err)}}
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
		if err := scan(s.InternalTag, val); err != nil {
			return []ValidationError{{ErrorInfo: fmt.Errorf("could not scan request parameter: %s", err)}}
		}
		return validator.validate(arr, scan)
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
		if err := scan(s.InternalTag, val); err != nil {
			return []ValidationError{{ErrorInfo: fmt.Errorf("could not scan request parameter: %s", err)}}
		}
		return ValidateMap((*map[string]RequestParameter)(&validator), &mapVal, scan)
	}
	return nil
}

func (s *arrayValue) validate(arr []any, scan func(tagName string, value any) error) []ValidationError {
	var wg sync.WaitGroup
	validationErrors := []ValidationError{}

	if s.Maximum > 0 && len(arr) < s.Maximum {
		return []ValidationError{{ErrorInfo: fmt.Errorf("invalid length of array")}}
	} else if s.Minimum > 0 && len(arr) < s.Minimum {
		return []ValidationError{{ErrorInfo: fmt.Errorf("invalid length of array")}}
	}

	for i, item := range arr {
		wg.Add(1)
		go func(i int, item any) {
			defer wg.Done()
			if err := s.OfType.validateValue(item, scan); err != nil {
				validationErrors = append(validationErrors, err...)
			}
		}(i, item)
	}

	wg.Wait()
	return validationErrors
}
