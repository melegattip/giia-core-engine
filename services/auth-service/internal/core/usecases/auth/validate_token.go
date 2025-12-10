package auth

import (
	"context"

	pkgErrors "github.com/giia/giia-core-engine/pkg/errors"
	pkgLogger "github.com/giia/giia-core-engine/pkg/logger"
	"github.com/giia/giia-core-engine/services/auth-service/internal/core/domain"
	"github.com/giia/giia-core-engine/services/auth-service/internal/core/providers"
	"github.com/giia/giia-core-engine/services/auth-service/internal/infrastructure/adapters/jwt"
	"github.com/google/uuid"
)

type ValidateTokenUseCase struct {
	userRepo   providers.UserRepository
	jwtManager *jwt.JWTManager
	logger     pkgLogger.Logger
}

type TokenValidationResult struct {
	Valid          bool
	UserID         uuid.UUID
	Email          string
	OrganizationID uuid.UUID
	Roles          []string
	ExpiresAt      int64
}

func NewValidateTokenUseCase(
	userRepo providers.UserRepository,
	jwtManager *jwt.JWTManager,
	logger pkgLogger.Logger,
) *ValidateTokenUseCase {
	return &ValidateTokenUseCase{
		userRepo:   userRepo,
		jwtManager: jwtManager,
		logger:     logger,
	}
}

func (uc *ValidateTokenUseCase) Execute(ctx context.Context, tokenString string) (*TokenValidationResult, error) {
	if tokenString == "" {
		return nil, pkgErrors.NewBadRequest("token is required")
	}

	claims, err := uc.jwtManager.ValidateAccessToken(tokenString)
	if err != nil {
		uc.logger.Warn(ctx, "Invalid access token", pkgLogger.Tags{
			"error": err.Error(),
		})
		return &TokenValidationResult{
			Valid: false,
		}, nil
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		uc.logger.Error(ctx, err, "Invalid user ID in token claims", pkgLogger.Tags{
			"user_id": claims.UserID,
		})
		return &TokenValidationResult{
			Valid: false,
		}, nil
	}

	orgID, err := uuid.Parse(claims.OrganizationID)
	if err != nil {
		uc.logger.Error(ctx, err, "Invalid organization ID in token claims", pkgLogger.Tags{
			"organization_id": claims.OrganizationID,
		})
		return &TokenValidationResult{
			Valid: false,
		}, nil
	}

	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		uc.logger.Warn(ctx, "User not found for valid token", pkgLogger.Tags{
			"user_id": userID.String(),
		})
		return &TokenValidationResult{
			Valid: false,
		}, nil
	}

	if user.Status != domain.UserStatusActive {
		uc.logger.Warn(ctx, "Token validation failed - user not active", pkgLogger.Tags{
			"user_id": user.ID.String(),
			"status":  string(user.Status),
		})
		return &TokenValidationResult{
			Valid: false,
		}, nil
	}

	expiresAt := claims.ExpiresAt.Unix()

	uc.logger.Info(ctx, "Token validated successfully", pkgLogger.Tags{
		"user_id":         userID.String(),
		"organization_id": orgID.String(),
	})

	return &TokenValidationResult{
		Valid:          true,
		UserID:         userID,
		Email:          claims.Email,
		OrganizationID: orgID,
		Roles:          claims.Roles,
		ExpiresAt:      expiresAt,
	}, nil
}
