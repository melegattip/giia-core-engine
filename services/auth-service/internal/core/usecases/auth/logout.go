package auth

import (
	"context"

	pkgErrors "github.com/giia/giia-core-engine/pkg/errors"
	pkgLogger "github.com/giia/giia-core-engine/pkg/logger"
	"github.com/giia/giia-core-engine/services/auth-service/internal/core/providers"
	"github.com/giia/giia-core-engine/services/auth-service/internal/infrastructure/adapters/jwt"
	"github.com/google/uuid"
)

type LogoutUseCase struct {
	tokenRepo  providers.TokenRepository
	jwtManager *jwt.JWTManager
	logger     pkgLogger.Logger
}

func NewLogoutUseCase(
	tokenRepo providers.TokenRepository,
	jwtManager *jwt.JWTManager,
	logger pkgLogger.Logger,
) *LogoutUseCase {
	return &LogoutUseCase{
		tokenRepo:  tokenRepo,
		jwtManager: jwtManager,
		logger:     logger,
	}
}

func (uc *LogoutUseCase) Execute(ctx context.Context, accessToken string, userID uuid.UUID) error {
	if accessToken == "" {
		return pkgErrors.NewBadRequest("access token is required")
	}

	ttl := uc.jwtManager.GetAccessExpiry()
	if err := uc.tokenRepo.BlacklistToken(ctx, accessToken, ttl); err != nil {
		uc.logger.Error(ctx, err, "Failed to blacklist access token", pkgLogger.Tags{
			"user_id": userID.String(),
		})
		return pkgErrors.NewInternalServerError("failed to blacklist token")
	}

	if err := uc.tokenRepo.RevokeAllUserTokens(ctx, userID); err != nil {
		uc.logger.Error(ctx, err, "Failed to revoke refresh tokens", pkgLogger.Tags{
			"user_id": userID.String(),
		})
	}

	uc.logger.Info(ctx, "User logged out successfully", pkgLogger.Tags{
		"user_id": userID.String(),
	})

	return nil
}
