package rbac

import (
	"context"
	"time"

	"github.com/google/uuid"

	pkgErrors "github.com/giia/giia-core-engine/pkg/errors"
	pkgLogger "github.com/giia/giia-core-engine/pkg/logger"
	"github.com/giia/giia-core-engine/services/auth-service/internal/core/providers"
)

const (
	permissionCacheTTL = 5 * time.Minute
)

type GetUserPermissionsUseCase struct {
	roleRepo           providers.RoleRepository
	resolveInheritance *ResolveInheritanceUseCase
	cache              providers.PermissionCache
	logger             pkgLogger.Logger
}

func NewGetUserPermissionsUseCase(
	roleRepo providers.RoleRepository,
	resolveInheritance *ResolveInheritanceUseCase,
	cache providers.PermissionCache,
	logger pkgLogger.Logger,
) *GetUserPermissionsUseCase {
	return &GetUserPermissionsUseCase{
		roleRepo:           roleRepo,
		resolveInheritance: resolveInheritance,
		cache:              cache,
		logger:             logger,
	}
}

func (uc *GetUserPermissionsUseCase) Execute(ctx context.Context, userID uuid.UUID) ([]string, error) {
	if userID == uuid.Nil {
		return nil, pkgErrors.NewBadRequest("user ID cannot be empty")
	}

	cached, err := uc.cache.GetUserPermissions(ctx, userID.String())
	if err == nil && cached != nil {
		uc.logger.Debug(ctx, "Cache hit for user permissions", pkgLogger.Tags{
			"user_id":           userID.String(),
			"permissions_count": len(cached),
		})
		return cached, nil
	}

	uc.logger.Debug(ctx, "Cache miss for user permissions", pkgLogger.Tags{
		"user_id": userID.String(),
	})

	roles, err := uc.roleRepo.GetUserRoles(ctx, userID)
	if err != nil {
		uc.logger.Error(ctx, err, "Failed to get user roles", pkgLogger.Tags{
			"user_id": userID.String(),
		})
		return nil, pkgErrors.NewInternalServerError("failed to get user roles")
	}

	if len(roles) == 0 {
		uc.logger.Warn(ctx, "User has no roles assigned", pkgLogger.Tags{
			"user_id": userID.String(),
		})
		return []string{}, nil
	}

	permissionMap := make(map[string]bool)
	hasWildcard := false

	for _, role := range roles {
		permissions, err := uc.resolveInheritance.Execute(ctx, role.ID)
		if err != nil {
			uc.logger.Error(ctx, err, "Failed to resolve role inheritance", pkgLogger.Tags{
				"user_id": userID.String(),
				"role_id": role.ID.String(),
			})
			return nil, pkgErrors.NewInternalServerError("failed to resolve permissions")
		}

		for _, perm := range permissions {
			if perm.Code == "*:*:*" {
				hasWildcard = true
				break
			}
			permissionMap[perm.Code] = true
		}

		if hasWildcard {
			break
		}
	}

	permissionCodes := make([]string, 0, len(permissionMap))
	if hasWildcard {
		permissionCodes = append(permissionCodes, "*:*:*")
	} else {
		for code := range permissionMap {
			permissionCodes = append(permissionCodes, code)
		}
	}

	if err := uc.cache.SetUserPermissions(ctx, userID.String(), permissionCodes, permissionCacheTTL); err != nil {
		uc.logger.Error(ctx, err, "Failed to cache user permissions", pkgLogger.Tags{
			"user_id": userID.String(),
		})
	}

	uc.logger.Info(ctx, "Retrieved user permissions", pkgLogger.Tags{
		"user_id":           userID.String(),
		"roles_count":       len(roles),
		"permissions_count": len(permissionCodes),
		"has_wildcard":      hasWildcard,
	})

	return permissionCodes, nil
}
