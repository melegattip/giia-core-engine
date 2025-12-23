package rbac

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	pkgErrors "github.com/melegattip/giia-core-engine/pkg/errors"
	pkgLogger "github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/auth-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/auth-service/internal/core/providers"
)

type ResolveInheritanceUseCase struct {
	roleRepo providers.RoleRepository
	permRepo providers.PermissionRepository
	logger   pkgLogger.Logger
}

func NewResolveInheritanceUseCase(
	roleRepo providers.RoleRepository,
	permRepo providers.PermissionRepository,
	logger pkgLogger.Logger,
) *ResolveInheritanceUseCase {
	return &ResolveInheritanceUseCase{
		roleRepo: roleRepo,
		permRepo: permRepo,
		logger:   logger,
	}
}

func (uc *ResolveInheritanceUseCase) Execute(ctx context.Context, roleID uuid.UUID) ([]*domain.Permission, error) {
	if roleID == uuid.Nil {
		return nil, pkgErrors.NewBadRequest("role ID cannot be empty")
	}

	visited := make(map[uuid.UUID]bool)
	permissionMap := make(map[string]*domain.Permission)

	if err := uc.collectPermissionsRecursive(ctx, roleID, visited, permissionMap); err != nil {
		return nil, err
	}

	permissions := make([]*domain.Permission, 0, len(permissionMap))
	for _, perm := range permissionMap {
		permissions = append(permissions, perm)
	}

	uc.logger.Debug(ctx, "Resolved role inheritance", pkgLogger.Tags{
		"role_id":           roleID.String(),
		"inherited_roles":   len(visited),
		"total_permissions": len(permissions),
	})

	return permissions, nil
}

func (uc *ResolveInheritanceUseCase) collectPermissionsRecursive(
	ctx context.Context,
	roleID uuid.UUID,
	visited map[uuid.UUID]bool,
	permissionMap map[string]*domain.Permission,
) error {
	if visited[roleID] {
		return pkgErrors.NewBadRequest(fmt.Sprintf("circular role hierarchy detected at role %s", roleID))
	}

	visited[roleID] = true

	role, err := uc.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		uc.logger.Error(ctx, err, "Failed to get role", pkgLogger.Tags{
			"role_id": roleID.String(),
		})
		return pkgErrors.NewInternalServerError("failed to resolve role hierarchy")
	}

	permissions, err := uc.permRepo.GetRolePermissions(ctx, roleID)
	if err != nil {
		uc.logger.Error(ctx, err, "Failed to get role permissions", pkgLogger.Tags{
			"role_id": roleID.String(),
		})
		return pkgErrors.NewInternalServerError("failed to get role permissions")
	}

	for _, perm := range permissions {
		permissionMap[perm.Code] = perm
	}

	if role.ParentRoleID != nil && *role.ParentRoleID != uuid.Nil {
		if err := uc.collectPermissionsRecursive(ctx, *role.ParentRoleID, visited, permissionMap); err != nil {
			return err
		}
	}

	return nil
}
