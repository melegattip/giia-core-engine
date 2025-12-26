package auth

import (
	"context"
	"strconv"

	pkgErrors "github.com/melegattip/giia-core-engine/pkg/errors"
	pkgLogger "github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/auth-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/auth-service/internal/core/providers"
)

type RefreshTokenUseCase struct {
	userRepo   providers.UserRepository
	tokenRepo  providers.TokenRepository
	jwtManager providers.JWTManager
	logger     pkgLogger.Logger
}

func NewRefreshTokenUseCase(
	userRepo providers.UserRepository,
	tokenRepo providers.TokenRepository,
	jwtManager providers.JWTManager,
	logger pkgLogger.Logger,
) *RefreshTokenUseCase {
	return &RefreshTokenUseCase{
		userRepo:   userRepo,
		tokenRepo:  tokenRepo,
		jwtManager: jwtManager,
		logger:     logger,
	}
}

func (uc *RefreshTokenUseCase) Execute(ctx context.Context, refreshTokenString string) (string, error) {
	if refreshTokenString == "" {
		return "", pkgErrors.NewBadRequest("refresh token is required")
	}

	claims, err := uc.jwtManager.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		uc.logger.Warn(ctx, "Invalid refresh token", pkgLogger.Tags{
			"error": err.Error(),
		})
		return "", pkgErrors.NewUnauthorized("invalid or expired refresh token")
	}

	tokenHash := hashToken(refreshTokenString)
	storedToken, err := uc.tokenRepo.GetRefreshToken(ctx, tokenHash)
	if err != nil {
		uc.logger.Warn(ctx, "Refresh token not found or expired", pkgLogger.Tags{
			"error": err.Error(),
		})
		return "", pkgErrors.NewUnauthorized("refresh token not found or expired")
	}

	if storedToken.Revoked {
		uc.logger.Warn(ctx, "Attempted to use revoked refresh token", pkgLogger.Tags{
			"user_id": strconv.Itoa(storedToken.UserID),
		})
		return "", pkgErrors.NewUnauthorized("refresh token has been revoked")
	}

	userID, err := strconv.Atoi(claims.Subject)
	if err != nil {
		return "", pkgErrors.NewUnauthorized("invalid user ID in token")
	}

	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		uc.logger.Error(ctx, err, "Failed to get user", pkgLogger.Tags{
			"user_id": strconv.Itoa(userID),
		})
		return "", pkgErrors.NewUnauthorized("user not found")
	}

	if user.Status != domain.UserStatusActive {
		uc.logger.Warn(ctx, "Refresh attempt for inactive user", pkgLogger.Tags{
			"user_id": user.IDString(),
			"status":  string(user.Status),
		})
		return "", pkgErrors.NewForbidden("account is not active")
	}

	accessToken, err := uc.jwtManager.GenerateAccessToken(
		user.ID,
		user.OrganizationID,
		user.Email,
		nil,
	)
	if err != nil {
		uc.logger.Error(ctx, err, "Failed to generate access token", pkgLogger.Tags{
			"user_id": user.IDString(),
		})
		return "", pkgErrors.NewInternalServerError("failed to generate access token")
	}

	uc.logger.Info(ctx, "Access token refreshed successfully", pkgLogger.Tags{
		"user_id":         user.IDString(),
		"organization_id": user.OrganizationID.String(),
	})

	return accessToken, nil
}
