package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	pkgErrors "github.com/giia/giia-core-engine/pkg/errors"
	pkgLogger "github.com/giia/giia-core-engine/pkg/logger"
	"github.com/giia/giia-core-engine/services/auth-service/internal/core/domain"
	"github.com/giia/giia-core-engine/services/auth-service/internal/core/providers"
	"github.com/giia/giia-core-engine/services/auth-service/internal/infrastructure/adapters/jwt"
	"golang.org/x/crypto/bcrypt"
)

type LoginUseCase struct {
	userRepo    providers.UserRepository
	tokenRepo   providers.TokenRepository
	jwtManager  *jwt.JWTManager
	logger      pkgLogger.Logger
}

func NewLoginUseCase(
	userRepo providers.UserRepository,
	tokenRepo providers.TokenRepository,
	jwtManager *jwt.JWTManager,
	logger pkgLogger.Logger,
) *LoginUseCase {
	return &LoginUseCase{
		userRepo:   userRepo,
		tokenRepo:  tokenRepo,
		jwtManager: jwtManager,
		logger:     logger,
	}
}

func (uc *LoginUseCase) Execute(ctx context.Context, req *domain.LoginRequest) (*domain.LoginResponse, error) {
	if req.Email == "" {
		return nil, pkgErrors.NewBadRequest("email is required")
	}

	if req.Password == "" {
		return nil, pkgErrors.NewBadRequest("password is required")
	}

	user, err := uc.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		uc.logger.Error(ctx, err, "Failed to get user by email", pkgLogger.Tags{
			"email": req.Email,
		})
		return nil, pkgErrors.NewUnauthorized("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		uc.logger.Warn(ctx, "Failed login attempt - invalid password", pkgLogger.Tags{
			"email":   req.Email,
			"user_id": user.ID.String(),
		})
		return nil, pkgErrors.NewUnauthorized("invalid email or password")
	}

	if user.Status != domain.UserStatusActive {
		uc.logger.Warn(ctx, "Login attempt for inactive user", pkgLogger.Tags{
			"email":   req.Email,
			"user_id": user.ID.String(),
			"status":  string(user.Status),
		})
		return nil, pkgErrors.NewForbidden("account is not active")
	}

	accessToken, err := uc.jwtManager.GenerateAccessToken(
		user.ID,
		user.OrganizationID,
		user.Email,
		nil,
	)
	if err != nil {
		uc.logger.Error(ctx, err, "Failed to generate access token", pkgLogger.Tags{
			"user_id": user.ID.String(),
		})
		return nil, pkgErrors.NewInternalServerError("failed to generate access token")
	}

	refreshTokenString, err := uc.jwtManager.GenerateRefreshToken(user.ID)
	if err != nil {
		uc.logger.Error(ctx, err, "Failed to generate refresh token", pkgLogger.Tags{
			"user_id": user.ID.String(),
		})
		return nil, pkgErrors.NewInternalServerError("failed to generate refresh token")
	}

	tokenHash := hashToken(refreshTokenString)
	refreshToken := &domain.RefreshToken{
		TokenHash: tokenHash,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(uc.jwtManager.GetRefreshExpiry()),
		Revoked:   false,
	}

	if err := uc.tokenRepo.StoreRefreshToken(ctx, refreshToken); err != nil {
		uc.logger.Error(ctx, err, "Failed to store refresh token", pkgLogger.Tags{
			"user_id": user.ID.String(),
		})
		return nil, pkgErrors.NewInternalServerError("failed to store refresh token")
	}

	if err := uc.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		uc.logger.Error(ctx, err, "Failed to update last login", pkgLogger.Tags{
			"user_id": user.ID.String(),
		})
	}

	uc.logger.Info(ctx, "User logged in successfully", pkgLogger.Tags{
		"user_id":         user.ID.String(),
		"email":           user.Email,
		"organization_id": user.OrganizationID.String(),
	})

	return &domain.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenString,
		ExpiresIn:    int(uc.jwtManager.GetAccessExpiry().Seconds()),
		User:         user.ToResponse(),
	}, nil
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
