package domain

import "github.com/giia/giia-core-engine/pkg/errors"

func NewValidationError(message string) error {
	return errors.NewBadRequest(message)
}
