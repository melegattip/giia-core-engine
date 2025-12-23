package rbac

import (
	"context"

	"github.com/google/uuid"

	pkgErrors "github.com/melegattip/giia-core-engine/pkg/errors"
	pkgLogger "github.com/melegattip/giia-core-engine/pkg/logger"
)

type BatchCheckPermissionsUseCase struct {
	checkPermission *CheckPermissionUseCase
	logger          pkgLogger.Logger
}

func NewBatchCheckPermissionsUseCase(
	checkPermission *CheckPermissionUseCase,
	logger pkgLogger.Logger,
) *BatchCheckPermissionsUseCase {
	return &BatchCheckPermissionsUseCase{
		checkPermission: checkPermission,
		logger:          logger,
	}
}

func (uc *BatchCheckPermissionsUseCase) Execute(ctx context.Context, userID uuid.UUID, permissions []string) (map[string]bool, error) {
	if userID == uuid.Nil {
		return nil, pkgErrors.NewBadRequest("user ID cannot be empty")
	}

	if len(permissions) == 0 {
		return nil, pkgErrors.NewBadRequest("permissions list cannot be empty")
	}

	results := make(map[string]bool, len(permissions))

	for _, permission := range permissions {
		allowed, err := uc.checkPermission.Execute(ctx, userID, permission)
		if err != nil {
			uc.logger.Error(ctx, err, "Failed to check permission in batch", pkgLogger.Tags{
				"user_id":    userID.String(),
				"permission": permission,
			})
			return nil, err
		}
		results[permission] = allowed
	}

	uc.logger.Debug(ctx, "Batch permission check completed", pkgLogger.Tags{
		"user_id":           userID.String(),
		"permissions_count": len(permissions),
	})

	return results, nil
}
