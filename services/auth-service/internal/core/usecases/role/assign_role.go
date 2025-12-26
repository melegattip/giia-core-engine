package role

import (
	"context"

	"github.com/google/uuid"

	"github.com/melegattip/giia-core-engine/pkg/errors"
	"github.com/melegattip/giia-core-engine/pkg/events"
	pkgLogger "github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/auth-service/internal/core/providers"
)

type AssignRoleUseCase struct {
	roleRepo       providers.RoleRepository
	userRepo       providers.UserRepository
	cache          providers.PermissionCache
	eventPublisher providers.EventPublisher
	timeManager    providers.TimeManager
	logger         pkgLogger.Logger
}

func NewAssignRoleUseCase(
	roleRepo providers.RoleRepository,
	userRepo providers.UserRepository,
	cache providers.PermissionCache,
	eventPublisher providers.EventPublisher,
	timeManager providers.TimeManager,
	logger pkgLogger.Logger,
) *AssignRoleUseCase {
	return &AssignRoleUseCase{
		roleRepo:       roleRepo,
		userRepo:       userRepo,
		cache:          cache,
		eventPublisher: eventPublisher,
		timeManager:    timeManager,
		logger:         logger,
	}
}

func (uc *AssignRoleUseCase) Execute(ctx context.Context, userID int, roleID, assignedBy uuid.UUID) error {
	if userID == 0 {
		return errors.NewBadRequest("user ID cannot be empty")
	}

	if roleID == uuid.Nil {
		return errors.NewBadRequest("role ID cannot be empty")
	}

	if assignedBy == uuid.Nil {
		return errors.NewBadRequest("assigned by user ID cannot be empty")
	}

	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		uc.logger.Error(ctx, err, "Failed to get user", pkgLogger.Tags{
			"user_id": userID,
		})
		return errors.NewNotFound("user not found")
	}

	role, err := uc.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		uc.logger.Error(ctx, err, "Failed to get role", pkgLogger.Tags{
			"role_id": roleID.String(),
		})
		return errors.NewNotFound("role not found")
	}

	if err := uc.roleRepo.AssignRoleToUser(ctx, uuid.Nil, roleID, assignedBy); err != nil {
		uc.logger.Error(ctx, err, "Failed to assign role to user", pkgLogger.Tags{
			"user_id":     userID,
			"role_id":     roleID.String(),
			"assigned_by": assignedBy.String(),
		})
		return errors.NewInternalServerError("failed to assign role to user")
	}

	if err := uc.cache.InvalidateUserPermissions(ctx, user.IDString()); err != nil {
		uc.logger.Error(ctx, err, "Failed to invalidate user permissions cache", pkgLogger.Tags{
			"user_id": userID,
		})
	}

	uc.logger.Info(ctx, "Role assigned to user successfully", pkgLogger.Tags{
		"user_id":     userID,
		"user_email":  user.Email,
		"role_id":     roleID.String(),
		"role_name":   role.Name,
		"assigned_by": assignedBy.String(),
	})

	uc.publishRoleAssignedEvent(ctx, user.OrganizationID.String(), user.IDString(), user.Email, roleID.String(), role.Name)

	return nil
}

func (uc *AssignRoleUseCase) publishRoleAssignedEvent(ctx context.Context, orgID, userID, userEmail, roleID, roleName string) {
	event := events.NewEvent(
		"user.role.assigned",
		"auth-service",
		orgID,
		uc.timeManager.Now(),
		map[string]interface{}{
			"user_id":    userID,
			"user_email": userEmail,
			"role_id":    roleID,
			"role_name":  roleName,
		},
	)

	if err := uc.eventPublisher.PublishAsync(ctx, "auth.user.role.assigned", event); err != nil {
		uc.logger.Error(ctx, err, "Failed to publish role assigned event", pkgLogger.Tags{
			"user_id": userID,
			"role_id": roleID,
		})
	}
}
