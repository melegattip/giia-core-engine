package domain

import "github.com/giia/giia-core-engine/pkg/errors"

func NewValidationError(message string) error {
	return errors.NewBadRequest(message)
}

func NewNotFoundError(message string) error {
	return errors.NewNotFound(message)
}

func NewInternalError(message string) error {
	return errors.NewInternalServerError(message)
}
