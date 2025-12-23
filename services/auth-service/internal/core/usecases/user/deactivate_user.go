package user

import (
	"context"
	"time"

	"github.com/google/uuid"

	pkgErrors "github.com/melegattip/giia-core-engine/pkg/errors"
	"github.com/melegattip/giia-core-engine/pkg/events"
	pkgLogger "github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/auth-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/auth-service/internal/core/providers"
)

type DeactivateUserUseCase struct {
	userRepo           providers.UserRepository
	permissionRepo     providers.PermissionRepository
	eventPublisher     providers.EventPublisher
	timeManager        providers.TimeManager
	logger             pkgLogger.Logger
	requiredPermission string
}

func NewDeactivateUserUseCase(
	userRepo providers.UserRepository,
	permissionRepo providers.PermissionRepository,
	eventPublisher providers.EventPublisher,
	timeManager providers.TimeManager,
	logger pkgLogger.Logger,
) *DeactivateUserUseCase {
	return &DeactivateUserUseCase{
		userRepo:           userRepo,
		permissionRepo:     permissionRepo,
		eventPublisher:     eventPublisher,
		timeManager:        timeManager,
		logger:             logger,
		requiredPermission: "users:deactivate",
	}
}

func (uc *DeactivateUserUseCase) Execute(ctx context.Context, adminUserID, targetUserID uuid.UUID) error {
	if adminUserID == uuid.Nil {
		return pkgErrors.NewBadRequest("admin user ID is required")
	}

	if targetUserID == uuid.Nil {
		return pkgErrors.NewBadRequest("target user ID is required")
	}

	if adminUserID == targetUserID {
		return pkgErrors.NewBadRequest("cannot deactivate your own account")
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
		uc.logger.Warn(ctx, "User attempted to deactivate account without permission", pkgLogger.Tags{
			"admin_user_id":  adminUserID.String(),
			"target_user_id": targetUserID.String(),
		})
		return pkgErrors.NewForbidden("insufficient permissions to deactivate users")
	}

	targetUser, err := uc.userRepo.GetByID(ctx, targetUserID)
	if err != nil {
		uc.logger.Error(ctx, err, "Failed to get target user", pkgLogger.Tags{
			"target_user_id": targetUserID.String(),
		})
		return pkgErrors.NewNotFound("user not found")
	}

	if targetUser.Status == domain.UserStatusInactive {
		uc.logger.Info(ctx, "User account already inactive", pkgLogger.Tags{
			"user_id": targetUser.ID.String(),
		})
		return nil
	}

	targetUser.Status = domain.UserStatusInactive

	if err := uc.userRepo.Update(ctx, targetUser); err != nil {
		uc.logger.Error(ctx, err, "Failed to update user status", pkgLogger.Tags{
			"user_id": targetUser.ID.String(),
		})
		return pkgErrors.NewInternalServerError("failed to deactivate user")
	}

	uc.publishUserDeactivatedEvent(ctx, targetUser, adminUserID)

	uc.logger.Info(ctx, "User account deactivated by admin", pkgLogger.Tags{
		"user_id":         targetUser.ID.String(),
		"email":           targetUser.Email,
		"organization_id": targetUser.OrganizationID.String(),
		"admin_user_id":   adminUserID.String(),
	})

	return nil
}

func (uc *DeactivateUserUseCase) checkAdminPermission(ctx context.Context, adminUserID uuid.UUID) (bool, error) {
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

func (uc *DeactivateUserUseCase) publishUserDeactivatedEvent(ctx context.Context, user *domain.User, adminUserID uuid.UUID) {
	event := events.NewEvent(
		"user.deactivated",
		"auth-service",
		user.OrganizationID.String(),
		uc.timeManager.Now(),
		map[string]interface{}{
			"user_id":         user.ID.String(),
			"email":           user.Email,
			"first_name":      user.FirstName,
			"last_name":       user.LastName,
			"status":          string(user.Status),
			"deactivated_by":  adminUserID.String(),
			"deactivated_at":  uc.timeManager.Now().Format(time.RFC3339),
		},
	)

	if err := uc.eventPublisher.PublishAsync(ctx, "auth.user.deactivated", event); err != nil {
		uc.logger.Error(ctx, err, "Failed to publish user deactivated event", pkgLogger.Tags{
			"user_id": user.ID.String(),
		})
	}
}
