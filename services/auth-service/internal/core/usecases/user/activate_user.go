package user

import (
	"context"
	"time"

	"github.com/google/uuid"

	pkgErrors "github.com/giia/giia-core-engine/pkg/errors"
	"github.com/giia/giia-core-engine/pkg/events"
	pkgLogger "github.com/giia/giia-core-engine/pkg/logger"
	"github.com/giia/giia-core-engine/services/auth-service/internal/core/domain"
	"github.com/giia/giia-core-engine/services/auth-service/internal/core/providers"
)

type ActivateUserUseCase struct {
	userRepo           providers.UserRepository
	permissionRepo     providers.PermissionRepository
	eventPublisher     providers.EventPublisher
	timeManager        providers.TimeManager
	logger             pkgLogger.Logger
	requiredPermission string
}

func NewActivateUserUseCase(
	userRepo providers.UserRepository,
	permissionRepo providers.PermissionRepository,
	eventPublisher providers.EventPublisher,
	timeManager providers.TimeManager,
	logger pkgLogger.Logger,
) *ActivateUserUseCase {
	return &ActivateUserUseCase{
		userRepo:           userRepo,
		permissionRepo:     permissionRepo,
		eventPublisher:     eventPublisher,
		timeManager:        timeManager,
		logger:             logger,
		requiredPermission: "users:activate",
	}
}

func (uc *ActivateUserUseCase) Execute(ctx context.Context, adminUserID, targetUserID uuid.UUID) error {
	if adminUserID == uuid.Nil {
		return pkgErrors.NewBadRequest("admin user ID is required")
	}

	if targetUserID == uuid.Nil {
		return pkgErrors.NewBadRequest("target user ID is required")
	}

	hasPermission, err := uc.checkAdminPermission(ctx, adminUserID)
	if err != nil {
		uc.logger.Error(ctx, err, "Failed to check admin permissions", pkgLogger.Tags{
			"admin_user_id":  adminUserID.String(),
			"target_user_id": targetUserID.String(),
		})
		return pkgErrors.NewInternalServerError("failed to verify permissions")
	}

	if !hasPermission {
		uc.logger.Warn(ctx, "User attempted to activate account without permission", pkgLogger.Tags{
			"admin_user_id":  adminUserID.String(),
			"target_user_id": targetUserID.String(),
		})
		return pkgErrors.NewForbidden("insufficient permissions to activate users")
	}

	targetUser, err := uc.userRepo.GetByID(ctx, targetUserID)
	if err != nil {
		uc.logger.Error(ctx, err, "Failed to get target user", pkgLogger.Tags{
			"target_user_id": targetUserID.String(),
		})
		return pkgErrors.NewNotFound("user not found")
	}

	if targetUser.Status == domain.UserStatusActive {
		uc.logger.Info(ctx, "User account already active", pkgLogger.Tags{
			"user_id": targetUser.ID.String(),
		})
		return nil
	}

	targetUser.Status = domain.UserStatusActive
	now := uc.timeManager.Now()
	targetUser.VerifiedAt = &now

	if err := uc.userRepo.Update(ctx, targetUser); err != nil {
		uc.logger.Error(ctx, err, "Failed to update user status", pkgLogger.Tags{
			"user_id": targetUser.ID.String(),
		})
		return pkgErrors.NewInternalServerError("failed to activate user")
	}

	uc.publishUserActivatedEvent(ctx, targetUser, adminUserID)

	uc.logger.Info(ctx, "User account activated by admin", pkgLogger.Tags{
		"user_id":         targetUser.ID.String(),
		"email":           targetUser.Email,
		"organization_id": targetUser.OrganizationID.String(),
		"admin_user_id":   adminUserID.String(),
	})

	return nil
}

func (uc *ActivateUserUseCase) checkAdminPermission(ctx context.Context, adminUserID uuid.UUID) (bool, error) {
	permissions, err := uc.permissionRepo.GetUserPermissions(ctx, adminUserID)
	if err != nil {
		return false, err
	}

	for _, permission := range permissions {
		if permission.Code == uc.requiredPermission {
			return true, nil
		}
	}

	return false, nil
}

func (uc *ActivateUserUseCase) publishUserActivatedEvent(ctx context.Context, user *domain.User, adminUserID uuid.UUID) {
	event := events.NewEvent(
		"user.activated",
		"auth-service",
		user.OrganizationID.String(),
		uc.timeManager.Now(),
		map[string]interface{}{
			"user_id":       user.ID.String(),
			"email":         user.Email,
			"first_name":    user.FirstName,
			"last_name":     user.LastName,
			"status":        string(user.Status),
			"activated_by":  adminUserID.String(),
			"activated_at":  uc.timeManager.Now().Format(time.RFC3339),
		},
	)

	if err := uc.eventPublisher.PublishAsync(ctx, "auth.user.activated", event); err != nil {
		uc.logger.Error(ctx, err, "Failed to publish user activated event", pkgLogger.Tags{
			"user_id": user.ID.String(),
		})
	}
}
