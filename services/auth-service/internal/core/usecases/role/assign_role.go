package role

import (
	"context"

	pkgErrors "github.com/giia/giia-core-engine/pkg/errors"
	pkgLogger "github.com/giia/giia-core-engine/pkg/logger"
	"github.com/giia/giia-core-engine/services/auth-service/internal/core/providers"
	"github.com/google/uuid"
)

type AssignRoleUseCase struct {
	roleRepo  providers.RoleRepository
	userRepo  providers.UserRepository
	cache     providers.PermissionCache
	logger    pkgLogger.Logger
}

func NewAssignRoleUseCase(
	roleRepo providers.RoleRepository,
	userRepo providers.UserRepository,
	cache providers.PermissionCache,
	logger pkgLogger.Logger,
) *AssignRoleUseCase {
	return &AssignRoleUseCase{
		roleRepo: roleRepo,
		userRepo: userRepo,
		cache:    cache,
		logger:   logger,
	}
}

func (uc *AssignRoleUseCase) Execute(ctx context.Context, userID, roleID, assignedBy uuid.UUID) error {
	if userID == uuid.Nil {
		return pkgErrors.NewBadRequest("user ID cannot be empty")
	}

	if roleID == uuid.Nil {
		return pkgErrors.NewBadRequest("role ID cannot be empty")
	}

	if assignedBy == uuid.Nil {
		return pkgErrors.NewBadRequest("assigned by user ID cannot be empty")
	}

	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		uc.logger.Error(ctx, err, "Failed to get user", pkgLogger.Tags{
			"user_id": userID.String(),
		})
		return pkgErrors.NewNotFound("user not found")
	}

	role, err := uc.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		uc.logger.Error(ctx, err, "Failed to get role", pkgLogger.Tags{
			"role_id": roleID.String(),
		})
		return pkgErrors.NewNotFound("role not found")
	}

	if err := uc.roleRepo.AssignRoleToUser(ctx, userID, roleID, assignedBy); err != nil {
		uc.logger.Error(ctx, err, "Failed to assign role to user", pkgLogger.Tags{
			"user_id":     userID.String(),
			"role_id":     roleID.String(),
			"assigned_by": assignedBy.String(),
		})
		return pkgErrors.NewInternalServerError("failed to assign role to user")
	}

	if err := uc.cache.InvalidateUserPermissions(ctx, userID.String()); err != nil {
		uc.logger.Error(ctx, err, "Failed to invalidate user permissions cache", pkgLogger.Tags{
			"user_id": userID.String(),
		})
	}

	uc.logger.Info(ctx, "Role assigned to user successfully", pkgLogger.Tags{
		"user_id":     userID.String(),
		"user_email":  user.Email,
		"role_id":     roleID.String(),
		"role_name":   role.Name,
		"assigned_by": assignedBy.String(),
	})

	return nil
}
