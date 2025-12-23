package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/melegattip/giia-core-engine/pkg/errors"
	"github.com/melegattip/giia-core-engine/pkg/events"
	pkgLogger "github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/auth-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/auth-service/internal/core/providers"
)

type LoginUseCase struct {
	userRepo       providers.UserRepository
	tokenRepo      providers.TokenRepository
	jwtManager     providers.JWTManager
	eventPublisher providers.EventPublisher
	timeManager    providers.TimeManager
	logger         pkgLogger.Logger
}

func NewLoginUseCase(
	userRepo providers.UserRepository,
	tokenRepo providers.TokenRepository,
	jwtManager providers.JWTManager,
	eventPublisher providers.EventPublisher,
	timeManager providers.TimeManager,
	logger pkgLogger.Logger,
) *LoginUseCase {
	return &LoginUseCase{
		userRepo:       userRepo,
		tokenRepo:      tokenRepo,
		jwtManager:     jwtManager,
		eventPublisher: eventPublisher,
		timeManager:    timeManager,
		logger:         logger,
	}
}

func (uc *LoginUseCase) Execute(ctx context.Context, req *domain.LoginRequest) (*domain.LoginResponse, error) {
	if req.Email == "" {
		return nil, errors.NewBadRequest("email is required")
	}

	if req.Password == "" {
		return nil, errors.NewBadRequest("password is required")
	}

	user, err := uc.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		uc.logger.Error(ctx, err, "Failed to get user by email", pkgLogger.Tags{
			"email": req.Email,
		})
		uc.publishLoginFailedEvent(ctx, req.Email, "", "user_not_found")
		return nil, errors.NewUnauthorized("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		uc.logger.Warn(ctx, "Failed login attempt - invalid password", pkgLogger.Tags{
			"email":   req.Email,
			"user_id": user.ID.String(),
		})
		uc.publishLoginFailedEvent(ctx, req.Email, user.OrganizationID.String(), "invalid_password")
		return nil, errors.NewUnauthorized("invalid email or password")
	}

	if user.Status != domain.UserStatusActive {
		uc.logger.Warn(ctx, "Login attempt for inactive user", pkgLogger.Tags{
			"email":   req.Email,
			"user_id": user.ID.String(),
			"status":  string(user.Status),
		})
		uc.publishLoginFailedEvent(ctx, req.Email, user.OrganizationID.String(), "inactive_account")
		return nil, errors.NewForbidden("account is not active")
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
		return nil, errors.NewInternalServerError("failed to generate access token")
	}

	refreshTokenString, err := uc.jwtManager.GenerateRefreshToken(user.ID)
	if err != nil {
		uc.logger.Error(ctx, err, "Failed to generate refresh token", pkgLogger.Tags{
			"user_id": user.ID.String(),
		})
		return nil, errors.NewInternalServerError("failed to generate refresh token")
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
		return nil, errors.NewInternalServerError("failed to store refresh token")
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

	uc.publishLoginSucceededEvent(ctx, user)

	return &domain.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenString,
		ExpiresIn:    int(uc.jwtManager.GetAccessExpiry().Seconds()),
		User:         user.ToResponse(),
	}, nil
}

func (uc *LoginUseCase) publishLoginSucceededEvent(ctx context.Context, user *domain.User) {
	event := events.NewEvent(
		"user.login.succeeded",
		"auth-service",
		user.OrganizationID.String(),
		uc.timeManager.Now(),
		map[string]interface{}{
			"user_id": user.ID.String(),
			"email":   user.Email,
		},
	)

	if err := uc.eventPublisher.PublishAsync(ctx, "auth.user.login.succeeded", event); err != nil {
		uc.logger.Error(ctx, err, "Failed to publish login succeeded event", pkgLogger.Tags{
			"user_id": user.ID.String(),
		})
	}
}

func (uc *LoginUseCase) publishLoginFailedEvent(ctx context.Context, email, organizationID, reason string) {
	if organizationID == "" {
		organizationID = "unknown"
	}

	event := events.NewEvent(
		"user.login.failed",
		"auth-service",
		organizationID,
		uc.timeManager.Now(),
		map[string]interface{}{
			"email":  email,
			"reason": reason,
		},
	)

	if err := uc.eventPublisher.PublishAsync(ctx, "auth.user.login.failed", event); err != nil {
		uc.logger.Error(ctx, err, "Failed to publish login failed event", pkgLogger.Tags{
			"email": email,
		})
	}
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
