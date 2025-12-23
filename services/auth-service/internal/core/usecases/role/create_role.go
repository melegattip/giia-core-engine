package role

import (
	"context"

	"github.com/google/uuid"

	pkgErrors "github.com/melegattip/giia-core-engine/pkg/errors"
	pkgLogger "github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/auth-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/auth-service/internal/core/providers"
)

type CreateRoleUseCase struct {
	roleRepo providers.RoleRepository
	permRepo providers.PermissionRepository
	logger   pkgLogger.Logger
}

func NewCreateRoleUseCase(
	roleRepo providers.RoleRepository,
	permRepo providers.PermissionRepository,
	logger pkgLogger.Logger,
) *CreateRoleUseCase {
	return &CreateRoleUseCase{
		roleRepo: roleRepo,
		permRepo: permRepo,
		logger:   logger,
	}
}

func (uc *CreateRoleUseCase) Execute(ctx context.Context, req *domain.CreateRoleRequest) (*domain.Role, error) {
	if req.Name == "" {
		return nil, pkgErrors.NewBadRequest("role name is required")
	}

	var orgID *uuid.UUID
	if req.OrganizationID != nil {
		parsed, err := uuid.Parse(*req.OrganizationID)
		if err != nil {
			return nil, pkgErrors.NewBadRequest("invalid organization ID format")
		}
		orgID = &parsed
	}

	existing, err := uc.roleRepo.GetByName(ctx, req.Name, orgID)
	if err == nil && existing != nil {
		return nil, pkgErrors.NewBadRequest("role with this name already exists in the organization")
	}

	var parentRoleID *uuid.UUID
	if req.ParentRoleID != nil {
		parsed, err := uuid.Parse(*req.ParentRoleID)
		if err != nil {
			return nil, pkgErrors.NewBadRequest("invalid parent role ID format")
		}

		parentRole, err := uc.roleRepo.GetByID(ctx, parsed)
		if err != nil {
			return nil, pkgErrors.NewNotFound("parent role not found")
		}

		if parentRole.IsSystem && orgID != nil {
			return nil, pkgErrors.NewBadRequest("cannot inherit from system role in organization-specific role")
		}

		parentRoleID = &parsed
	}

	role := &domain.Role{
		Name:           req.Name,
		Description:    req.Description,
		OrganizationID: orgID,
		ParentRoleID:   parentRoleID,
		IsSystem:       false,
	}

	if err := uc.roleRepo.Create(ctx, role); err != nil {
		uc.logger.Error(ctx, err, "Failed to create role", pkgLogger.Tags{
			"role_name": req.Name,
		})
		return nil, pkgErrors.NewInternalServerError("failed to create role")
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

		if err := uc.permRepo.AssignPermissionsToRole(ctx, role.ID, permissionUUIDs); err != nil {
			uc.logger.Error(ctx, err, "Failed to assign permissions to role", pkgLogger.Tags{
				"role_id": role.ID.String(),
			})
			return nil, pkgErrors.NewInternalServerError("failed to assign permissions to role")
		}
	}

	uc.logger.Info(ctx, "Role created successfully", pkgLogger.Tags{
		"role_id":           role.ID.String(),
		"role_name":         role.Name,
		"organization_id":   orgID,
		"permissions_count": len(req.PermissionIDs),
	})

	return role, nil
}
