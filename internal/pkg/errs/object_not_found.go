package errs

import (
	"errors"
	"fmt"
)

var ErrObjectNotFound = errors.New("object not found")

type ObjectNotFoundError struct {
	ParamName string
	ID        any
	Cause     error
}

func NewObjectNotFoundError(paramName string, id any) *ObjectNotFoundError {
	return &ObjectNotFoundError{
		ParamName: paramName,
		ID:        id,
		Cause:     nil,
	}
}

func NewObjectNotFoundErrorWithCause(paramName string, id any, cause error) *ObjectNotFoundError {
	return &ObjectNotFoundError{
		ParamName: paramName,
		ID:        id,
		Cause:     cause,
	}
}

func (e *ObjectNotFoundError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: param: %s, ID: %s (cause: %s)",
			ErrObjectNotFound, e.ParamName, e.ID, e.Cause)
	}
	return fmt.Sprintf("%s: %s", ErrObjectNotFound, e.ID)
}

func (e *ObjectNotFoundError) Unwrap() error {
	return ErrObjectNotFound
}
