package auth

import (
	"context"
	"time"

	"github.com/google/uuid"

	pkgErrors "github.com/melegattip/giia-core-engine/pkg/errors"
	pkgLogger "github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/auth-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/auth-service/internal/core/providers"
)

type RequestPasswordResetUseCase struct {
	userRepo     providers.UserRepository
	tokenRepo    providers.TokenRepository
	emailService providers.EmailService
	logger       pkgLogger.Logger
}

func NewRequestPasswordResetUseCase(
	userRepo providers.UserRepository,
	tokenRepo providers.TokenRepository,
	emailService providers.EmailService,
	logger pkgLogger.Logger,
) *RequestPasswordResetUseCase {
	return &RequestPasswordResetUseCase{
		userRepo:     userRepo,
		tokenRepo:    tokenRepo,
		emailService: emailService,
		logger:       logger,
	}
}

func (uc *RequestPasswordResetUseCase) Execute(ctx context.Context, email string, organizationID uuid.UUID) error {
	if email == "" {
		return pkgErrors.NewBadRequest("email is required")
	}

	if err := validateEmail(email); err != nil {
		return err
	}

	user, err := uc.userRepo.GetByEmailAndOrg(ctx, email, organizationID)
	if err != nil {
		uc.logger.Info(ctx, "Password reset requested for non-existent email", pkgLogger.Tags{
			"email":           email,
			"organization_id": organizationID.String(),
		})
		return nil
	}

	resetToken := uuid.New().String()
	tokenHash := hashActivationToken(resetToken)

	passwordResetToken := &domain.PasswordResetToken{
		TokenHash: tokenHash,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(1 * time.Hour),
		Used:      false,
	}

	if err := uc.tokenRepo.StorePasswordResetToken(ctx, passwordResetToken); err != nil {
		uc.logger.Error(ctx, err, "Failed to store password reset token", pkgLogger.Tags{
			"user_id": user.ID.String(),
		})
		return pkgErrors.NewInternalServerError("failed to initiate password reset")
	}

	if err := uc.emailService.SendPasswordResetEmail(ctx, user.Email, resetToken, user.FirstName); err != nil {
		uc.logger.Error(ctx, err, "Failed to send password reset email", pkgLogger.Tags{
			"user_id": user.ID.String(),
			"email":   user.Email,
		})
	}

	uc.logger.Info(ctx, "Password reset requested successfully", pkgLogger.Tags{
		"user_id":         user.ID.String(),
		"email":           user.Email,
		"organization_id": user.OrganizationID.String(),
	})

	return nil
}
