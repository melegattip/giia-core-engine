package domain

import (
	"errors"
	"fmt"
)

var (
	ErrValidation = errors.New("validation error")
	ErrNotFound   = errors.New("not found")
	ErrConflict   = errors.New("conflict")
)

func NewValidationError(message string) error {
	return fmt.Errorf("%w: %s", ErrValidation, message)
}

func NewNotFoundError(resource string) error {
	return fmt.Errorf("%w: %s not found", ErrNotFound, resource)
}

func NewConflictError(message string) error {
	return fmt.Errorf("%w: %s", ErrConflict, message)
}
