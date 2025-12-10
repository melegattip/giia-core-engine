package auth

import (
	"context"

	pkgErrors "github.com/giia/giia-core-engine/pkg/errors"
	pkgLogger "github.com/giia/giia-core-engine/pkg/logger"
	"github.com/giia/giia-core-engine/services/auth-service/internal/core/domain"
	"github.com/giia/giia-core-engine/services/auth-service/internal/core/providers"
)

type ActivateAccountUseCase struct {
	userRepo     providers.UserRepository
	tokenRepo    providers.TokenRepository
	emailService providers.EmailService
	logger       pkgLogger.Logger
}

func NewActivateAccountUseCase(
	userRepo providers.UserRepository,
	tokenRepo providers.TokenRepository,
	emailService providers.EmailService,
	logger pkgLogger.Logger,
) *ActivateAccountUseCase {
	return &ActivateAccountUseCase{
		userRepo:     userRepo,
		tokenRepo:    tokenRepo,
		emailService: emailService,
		logger:       logger,
	}
}

func (uc *ActivateAccountUseCase) Execute(ctx context.Context, token string) error {
	if token == "" {
		return pkgErrors.NewBadRequest("activation token is required")
	}

	tokenHash := hashActivationToken(token)

	activationToken, err := uc.tokenRepo.GetActivationToken(ctx, tokenHash)
	if err != nil {
		uc.logger.Warn(ctx, "Activation token not found or expired", pkgLogger.Tags{
			"error": err.Error(),
		})
		return pkgErrors.NewBadRequest("invalid or expired activation token")
	}

	if activationToken.Used {
		uc.logger.Warn(ctx, "Attempted to use already used activation token", pkgLogger.Tags{
			"user_id": activationToken.UserID.String(),
		})
		return pkgErrors.NewBadRequest("activation token has already been used")
	}

	user, err := uc.userRepo.GetByID(ctx, activationToken.UserID)
	if err != nil {
		uc.logger.Error(ctx, err, "Failed to get user for activation", pkgLogger.Tags{
			"user_id": activationToken.UserID.String(),
		})
		return pkgErrors.NewInternalServerError("failed to activate account")
	}

	if user.Status == domain.UserStatusActive {
		uc.logger.Info(ctx, "User account already active", pkgLogger.Tags{
			"user_id": user.ID.String(),
		})
		return nil
	}

	user.Status = domain.UserStatusActive
	if err := uc.userRepo.Update(ctx, user); err != nil {
		uc.logger.Error(ctx, err, "Failed to update user status", pkgLogger.Tags{
			"user_id": user.ID.String(),
		})
		return pkgErrors.NewInternalServerError("failed to activate account")
	}

	if err := uc.tokenRepo.MarkActivationTokenUsed(ctx, tokenHash); err != nil {
		uc.logger.Error(ctx, err, "Failed to mark activation token as used", pkgLogger.Tags{
			"user_id": user.ID.String(),
		})
	}

	if err := uc.emailService.SendWelcomeEmail(ctx, user.Email, user.FirstName); err != nil {
		uc.logger.Error(ctx, err, "Failed to send welcome email", pkgLogger.Tags{
			"user_id": user.ID.String(),
			"email":   user.Email,
		})
	}

	uc.logger.Info(ctx, "User account activated successfully", pkgLogger.Tags{
		"user_id":         user.ID.String(),
		"email":           user.Email,
		"organization_id": user.OrganizationID.String(),
	})

	return nil
}
