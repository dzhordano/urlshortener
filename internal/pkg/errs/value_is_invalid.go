package errs

import (
	"errors"
	"fmt"
)

var ErrValueIsInvalid = errors.New("value is invalid")

type ValueIsInvalidError struct {
	ParamName string
	Cause     error
}

func NewValueIsInvalidError(paramName string) *ValueIsInvalidError {
	return &ValueIsInvalidError{
		ParamName: paramName,
		Cause:     nil,
	}
}

func NewValueIsInvalidErrorWithCause(paramName string, cause error) *ValueIsInvalidError {
	return &ValueIsInvalidError{
		ParamName: paramName,
		Cause:     cause,
	}
}

func (e *ValueIsInvalidError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (cause: %s)", ErrValueIsInvalid, e.ParamName, e.Cause)
	}
	return fmt.Sprintf("%s: %s", ErrValueIsInvalid, e.ParamName)
}

func (e *ValueIsInvalidError) Unwrap() error {
	return ErrValueIsInvalid
}
