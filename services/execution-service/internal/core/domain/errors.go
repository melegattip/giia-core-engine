package domain

import "errors"

var (
	ErrValidation           = errors.New("validation error")
	ErrNotFound             = errors.New("resource not found")
	ErrConflict             = errors.New("resource conflict")
	ErrUnauthorized         = errors.New("unauthorized")
	ErrInternalServerError  = errors.New("internal server error")
)

func NewValidationError(message string) error {
	return errors.New(message)
}

func NewNotFoundError(message string) error {
	return errors.New(message)
}

func NewConflictError(message string) error {
	return errors.New(message)
}