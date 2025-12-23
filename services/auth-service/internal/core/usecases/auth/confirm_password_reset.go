package auth

import (
	"context"

	"golang.org/x/crypto/bcrypt"

	pkgErrors "github.com/giia/giia-core-engine/pkg/errors"
	pkgLogger "github.com/giia/giia-core-engine/pkg/logger"
	"github.com/giia/giia-core-engine/services/auth-service/internal/core/providers"
)

type ConfirmPasswordResetUseCase struct {
	userRepo  providers.UserRepository
	tokenRepo providers.TokenRepository
	logger    pkgLogger.Logger
}

func NewConfirmPasswordResetUseCase(
	userRepo providers.UserRepository,
	tokenRepo providers.TokenRepository,
	logger pkgLogger.Logger,
) *ConfirmPasswordResetUseCase {
	return &ConfirmPasswordResetUseCase{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		logger:    logger,
	}
}

func (uc *ConfirmPasswordResetUseCase) Execute(ctx context.Context, token, newPassword string) error {
	if token == "" {
		return pkgErrors.NewBadRequest("reset token is required")
	}

	if newPassword == "" {
		return pkgErrors.NewBadRequest("new password is required")
	}

	if err := validatePassword(newPassword); err != nil {
		return err
	}

	tokenHash := hashActivationToken(token)

	resetToken, err := uc.tokenRepo.GetPasswordResetToken(ctx, tokenHash)
	if err != nil {
		uc.logger.Warn(ctx, "Password reset token not found or expired", pkgLogger.Tags{
			"error": err.Error(),
		})
		return pkgErrors.NewBadRequest("invalid or expired reset token")
	}

	if resetToken.Used {
		uc.logger.Warn(ctx, "Attempted to use already used reset token", pkgLogger.Tags{
			"user_id": resetToken.UserID.String(),
		})
		return pkgErrors.NewBadRequest("reset token has already been used")
	}

	user, err := uc.userRepo.GetByID(ctx, resetToken.UserID)
	if err != nil {
		uc.logger.Error(ctx, err, "Failed to get user for password reset", pkgLogger.Tags{
			"user_id": resetToken.UserID.String(),
		})
		return pkgErrors.NewInternalServerError("failed to reset password")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		uc.logger.Error(ctx, err, "Failed to hash new password", nil)
		return pkgErrors.NewInternalServerError("failed to reset password")
	}

	user.Password = string(hashedPassword)
	if err := uc.userRepo.Update(ctx, user); err != nil {
		uc.logger.Error(ctx, err, "Failed to update user password", pkgLogger.Tags{
			"user_id": user.ID.String(),
		})
		return pkgErrors.NewInternalServerError("failed to reset password")
	}

	if err := uc.tokenRepo.MarkPasswordResetTokenUsed(ctx, tokenHash); err != nil {
		uc.logger.Error(ctx, err, "Failed to mark reset token as used", pkgLogger.Tags{
			"user_id": user.ID.String(),
		})
	}

	uc.logger.Info(ctx, "Password reset completed successfully", pkgLogger.Tags{
		"user_id":         user.ID.String(),
		"email":           user.Email,
		"organization_id": user.OrganizationID.String(),
	})

	return nil
}
