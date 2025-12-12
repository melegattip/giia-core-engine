package rbac

import (
	"context"
	"strings"

	"github.com/google/uuid"

	pkgErrors "github.com/giia/giia-core-engine/pkg/errors"
	pkgLogger "github.com/giia/giia-core-engine/pkg/logger"
)

type CheckPermissionUseCase struct {
	getUserPermissions *GetUserPermissionsUseCase
	logger             pkgLogger.Logger
}

func NewCheckPermissionUseCase(
	getUserPermissions *GetUserPermissionsUseCase,
	logger pkgLogger.Logger,
) *CheckPermissionUseCase {
	return &CheckPermissionUseCase{
		getUserPermissions: getUserPermissions,
		logger:             logger,
	}
}

func (uc *CheckPermissionUseCase) Execute(ctx context.Context, userID uuid.UUID, requiredPermission string) (bool, error) {
	if userID == uuid.Nil {
		return false, pkgErrors.NewBadRequest("user ID cannot be empty")
	}

	if requiredPermission == "" {
		return false, pkgErrors.NewBadRequest("permission cannot be empty")
	}

	userPermissions, err := uc.getUserPermissions.Execute(ctx, userID)
	if err != nil {
		return false, err
	}

	allowed := uc.checkPermissionMatch(userPermissions, requiredPermission)

	uc.logger.Debug(ctx, "Permission check completed", pkgLogger.Tags{
		"user_id":    userID.String(),
		"permission": requiredPermission,
		"allowed":    allowed,
	})

	return allowed, nil
}

func (uc *CheckPermissionUseCase) checkPermissionMatch(userPermissions []string, requiredPermission string) bool {
	for _, userPerm := range userPermissions {
		if userPerm == "*:*:*" {
			return true
		}

		if userPerm == requiredPermission {
			return true
		}

		if uc.matchesWildcard(userPerm, requiredPermission) {
			return true
		}
	}

	return false
}

func (uc *CheckPermissionUseCase) matchesWildcard(pattern, permission string) bool {
	patternParts := strings.Split(pattern, ":")
	permissionParts := strings.Split(permission, ":")

	if len(patternParts) != 3 || len(permissionParts) != 3 {
		return false
	}

	for i := 0; i < 3; i++ {
		if patternParts[i] == "*" {
			continue
		}
		if patternParts[i] != permissionParts[i] {
			return false
		}
	}

	return true
}
