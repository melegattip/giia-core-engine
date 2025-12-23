package providers

import (
	"context"

	"github.com/google/uuid"
)

type AuthClient interface {
	ValidateToken(ctx context.Context, token string) (*TokenValidationResult, error)
	CheckPermission(ctx context.Context, userID uuid.UUID, orgID uuid.UUID, permission string) (bool, error)
	Close() error
}

type TokenValidationResult struct {
	Valid          bool
	UserID         uuid.UUID
	OrganizationID uuid.UUID
	Email          string
	Reason         string
}
