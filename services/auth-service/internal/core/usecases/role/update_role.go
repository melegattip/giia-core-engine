package role

import (
	"context"

	pkgErrors "github.com/giia/giia-core-engine/pkg/errors"
	pkgLogger "github.com/giia/giia-core-engine/pkg/logger"
	"github.com/giia/giia-core-engine/services/auth-service/internal/core/domain"
	"github.com/giia/giia-core-engine/services/auth-service/internal/core/providers"
	"github.com/google/uuid"
)

type UpdateRoleUseCase struct {
	roleRepo providers.RoleRepository
	permRepo providers.PermissionRepository
	cache    providers.PermissionCache
	logger   pkgLogger.Logger
}

func NewUpdateRoleUseCase(
	roleRepo providers.RoleRepository,
	permRepo providers.PermissionRepository,
	cache providers.PermissionCache,
	logger pkgLogger.Logger,
) *UpdateRoleUseCase {
	return &UpdateRoleUseCase{
		roleRepo: roleRepo,
		permRepo: permRepo,
		cache:    cache,
		logger:   logger,
	}
}

func (uc *UpdateRoleUseCase) Execute(ctx context.Context, roleID uuid.UUID, req *domain.UpdateRoleRequest) (*domain.Role, error) {
	if roleID == uuid.Nil {
		return nil, pkgErrors.NewBadRequest("role ID cannot be empty")
	}

	role, err := uc.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		uc.logger.Error(ctx, err, "Failed to get role", pkgLogger.Tags{
			"role_id": roleID.String(),
		})
		return nil, pkgErrors.NewNotFound("role not found")
	}

	if role.IsSystem {
		return nil, pkgErrors.NewBadRequest("cannot update system roles")
	}

	if req.Name != "" {
		role.Name = req.Name
	}

	if req.Description != "" {
		role.Description = req.Description
	}

	if req.ParentRoleID != nil {
		parsed, err := uuid.Parse(*req.ParentRoleID)
		if err != nil {
			return nil, pkgErrors.NewBadRequest("invalid parent role ID format")
		}

		parentRole, err := uc.roleRepo.GetByID(ctx, parsed)
		if err != nil {
			return nil, pkgErrors.NewNotFound("parent role not found")
		}

		if parentRole.IsSystem && role.OrganizationID != nil {
			return nil, pkgErrors.NewBadRequest("cannot inherit from system role in organization-specific role")
		}

		role.ParentRoleID = &parsed
	}

	if err := uc.roleRepo.Update(ctx, role); err != nil {
		uc.logger.Error(ctx, err, "Failed to update role", pkgLogger.Tags{
			"role_id": roleID.String(),
		})
		return nil, pkgErrors.NewInternalServerError("failed to update role")
	}

	if len(req.PermissionIDs) > 0 {
		permissionUUIDs := make([]uuid.UUID, len(req.PermissionIDs))
		for i, permID := range req.PermissionIDs {
			parsed, err := uuid.Parse(permID)
			if err != nil {
				return nil, pkgErrors.NewBadRequest("invalid permission ID format")
			}
			permissionUUIDs[i] = parsed
		}

		if err := uc.permRepo.ReplaceRolePermissions(ctx, roleID, permissionUUIDs); err != nil {
			uc.logger.Error(ctx, err, "Failed to update role permissions", pkgLogger.Tags{
				"role_id": roleID.String(),
			})
			return nil, pkgErrors.NewInternalServerError("failed to update role permissions")
		}

		userIDs, err := uc.roleRepo.GetUsersWithRole(ctx, roleID)
		if err != nil {
			uc.logger.Error(ctx, err, "Failed to get users with role", pkgLogger.Tags{
				"role_id": roleID.String(),
			})
		} else if len(userIDs) > 0 {
			userIDStrings := make([]string, len(userIDs))
			for i, id := range userIDs {
				userIDStrings[i] = id.String()
			}

			if err := uc.cache.InvalidateUsersWithRole(ctx, userIDStrings); err != nil {
				uc.logger.Error(ctx, err, "Failed to invalidate cache for users with role", pkgLogger.Tags{
					"role_id":    roleID.String(),
					"user_count": len(userIDs),
				})
			}
		}
	}

	uc.logger.Info(ctx, "Role updated successfully", pkgLogger.Tags{
		"role_id":   roleID.String(),
		"role_name": role.Name,
	})

	return role, nil
}
