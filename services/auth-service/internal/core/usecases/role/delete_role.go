package role

import (
	"context"

	pkgErrors "github.com/giia/giia-core-engine/pkg/errors"
	pkgLogger "github.com/giia/giia-core-engine/pkg/logger"
	"github.com/giia/giia-core-engine/services/auth-service/internal/core/providers"
	"github.com/google/uuid"
)

type DeleteRoleUseCase struct {
	roleRepo providers.RoleRepository
	cache    providers.PermissionCache
	logger   pkgLogger.Logger
}

func NewDeleteRoleUseCase(
	roleRepo providers.RoleRepository,
	cache providers.PermissionCache,
	logger pkgLogger.Logger,
) *DeleteRoleUseCase {
	return &DeleteRoleUseCase{
		roleRepo: roleRepo,
		cache:    cache,
		logger:   logger,
	}
}

func (uc *DeleteRoleUseCase) Execute(ctx context.Context, roleID uuid.UUID) error {
	if roleID == uuid.Nil {
		return pkgErrors.NewBadRequest("role ID cannot be empty")
	}

	role, err := uc.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		uc.logger.Error(ctx, err, "Failed to get role", pkgLogger.Tags{
			"role_id": roleID.String(),
		})
		return pkgErrors.NewNotFound("role not found")
	}

	if role.IsSystem {
		return pkgErrors.NewBadRequest("cannot delete system roles")
	}

	userIDs, err := uc.roleRepo.GetUsersWithRole(ctx, roleID)
	if err != nil {
		uc.logger.Error(ctx, err, "Failed to get users with role", pkgLogger.Tags{
			"role_id": roleID.String(),
		})
		return pkgErrors.NewInternalServerError("failed to verify role usage")
	}

	if err := uc.roleRepo.Delete(ctx, roleID); err != nil {
		uc.logger.Error(ctx, err, "Failed to delete role", pkgLogger.Tags{
			"role_id": roleID.String(),
		})
		return pkgErrors.NewInternalServerError("failed to delete role")
	}

	if len(userIDs) > 0 {
		userIDStrings := make([]string, len(userIDs))
		for i, id := range userIDs {
			userIDStrings[i] = id.String()
		}

		if err := uc.cache.InvalidateUsersWithRole(ctx, userIDStrings); err != nil {
			uc.logger.Error(ctx, err, "Failed to invalidate cache for affected users", pkgLogger.Tags{
				"role_id":    roleID.String(),
				"user_count": len(userIDs),
			})
		}
	}

	uc.logger.Info(ctx, "Role deleted successfully", pkgLogger.Tags{
		"role_id":          roleID.String(),
		"role_name":        role.Name,
		"affected_users": len(userIDs),
	})

	return nil
}
