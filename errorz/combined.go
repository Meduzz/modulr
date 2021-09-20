package errorz

import (
	"errors"
	"fmt"
)

// CombinedError - struct that combines a bunch of errors to one
type CombinedError struct {
	errors []error
}

// NewError - returns a new CombinedError
func NewError(err error) *CombinedError {
	erz := make([]error, 0)

	created := &CombinedError{
		errors: erz,
	}

	if err != nil {
		created.Append(err)
	}

	return created
}

// Append - append an error, ignored if nil
func (e *CombinedError) Append(err error) {
	if err != nil {
		e.errors = append(e.errors, err)
	}
}

// Error - compile the sum of the appended errors
func (e *CombinedError) Error() error {
	if len(e.errors) == 1 {
		return e.errors[0]
	}

	main := ""
	for _, err := range e.errors {
		main = fmt.Sprintf("%s\n%s", main, err.Error())
	}

	if main == "" {
		return nil
	}

	return errors.New(main)
}
