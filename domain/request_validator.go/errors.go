package requestvalidator

type ValidationError struct {
	Internal  bool
	ErrorInfo error
}
