package providers

import (
	"context"
	"time"

	"github.com/giia/giia-core-engine/services/auth-service/internal/core/domain"
	"github.com/google/uuid"
)

type TokenRepository interface {
	// Refresh Token Operations
	StoreRefreshToken(ctx context.Context, token *domain.RefreshToken) error
	GetRefreshToken(ctx context.Context, tokenHash string) (*domain.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, tokenHash string) error
	RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error

	// Password Reset Token Operations
	StorePasswordResetToken(ctx context.Context, token *domain.PasswordResetToken) error
	GetPasswordResetToken(ctx context.Context, tokenHash string) (*domain.PasswordResetToken, error)
	MarkPasswordResetTokenUsed(ctx context.Context, tokenHash string) error

	// Activation Token Operations
	StoreActivationToken(ctx context.Context, token *domain.ActivationToken) error
	GetActivationToken(ctx context.Context, tokenHash string) (*domain.ActivationToken, error)
	MarkActivationTokenUsed(ctx context.Context, tokenHash string) error

	// Blacklist Operations (for access tokens)
	BlacklistToken(ctx context.Context, token string, ttl time.Duration) error
	IsTokenBlacklisted(ctx context.Context, token string) (bool, error)
}
