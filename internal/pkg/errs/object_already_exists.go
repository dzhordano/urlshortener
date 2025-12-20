package errs

import (
	"errors"
	"fmt"
)

var ErrObjectAlreadyExists = errors.New("object already exists")

type ObjectAlreadyExistsError struct {
	ParamName string
	ID        any
	Cause     error
}

func NewObjectAlreadyExistsError(paramName string, id any) *ObjectAlreadyExistsError {
	return &ObjectAlreadyExistsError{
		ParamName: paramName,
		ID:        id,
		Cause:     nil,
	}
}

func NewObjectAlreadyExistsErrorWithCause(
	paramName string,
	id any,
	cause error,
) *ObjectAlreadyExistsError {
	return &ObjectAlreadyExistsError{
		ParamName: paramName,
		ID:        id,
		Cause:     cause,
	}
}

func (e *ObjectAlreadyExistsError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: param: %s, ID: %s (cause: %s)",
			ErrObjectAlreadyExists, e.ParamName, e.ID, e.Cause)
	}
	return fmt.Sprintf("%s: %s", ErrObjectAlreadyExists, e.ID)
}

func (e *ObjectAlreadyExistsError) Unwrap() error {
	return ErrObjectAlreadyExists
}
