package requestvalidator

import "fmt"

type ValidationError struct {
	Internal  bool
	ErrorInfo error
}

func (v *ValidationError) Error() string {
	if v.Internal {
		return fmt.Sprintf("%s | %s", "internal", v.ErrorInfo.Error())
	} else {
		return fmt.Sprintf("%s | %s", "not internal", v.ErrorInfo.Error())
	}
}

func Normalize(vErr []ValidationError) []error {
	e := make([]error, 0, len(vErr))
	for _, v := range vErr {
		e = append(e, v.ErrorInfo)
	}
	return e
}
