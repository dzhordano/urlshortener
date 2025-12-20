package errs

import (
	"errors"
	"fmt"
)

var ErrValueIsRequired = errors.New("value is required")

type ValueIsRequiredError struct {
	ParamName string
	Cause     error
}

func NewValueIsRequiredError(paramName string) *ValueIsRequiredError {
	return &ValueIsRequiredError{
		ParamName: paramName,
		Cause:     nil,
	}
}

func NewValueIsRequiredErrorWithCause(paramName string, cause error) *ValueIsRequiredError {
	return &ValueIsRequiredError{
		ParamName: paramName,
		Cause:     cause,
	}
}

func (e *ValueIsRequiredError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (cause: %s)", ErrValueIsRequired, e.ParamName, e.Cause)
	}
	return fmt.Sprintf("%s: %s", ErrValueIsRequired, e.ParamName)
}

func (e *ValueIsRequiredError) Unwrap() error {
	return ErrValueIsRequired
}
