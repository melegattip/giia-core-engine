package providers

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID         string   `json:"user_id"`
	Email          string   `json:"email"`
	OrganizationID string   `json:"organization_id"`
	Roles          []string `json:"roles,omitempty"`
	jwt.RegisteredClaims
}

type JWTManager interface {
	GenerateAccessToken(userID, orgID uuid.UUID, email string, roles []string) (string, error)
	GenerateRefreshToken(userID uuid.UUID) (string, error)
	ValidateAccessToken(tokenString string) (*Claims, error)
	ValidateRefreshToken(tokenString string) (*jwt.RegisteredClaims, error)
	GetAccessExpiry() time.Duration
	GetRefreshExpiry() time.Duration
}
