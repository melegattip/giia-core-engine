package providers

import (
	"context"

	"github.com/google/uuid"

	"github.com/giia/giia-core-engine/services/auth-service/internal/core/domain"
)

type PermissionRepository interface {
	Create(ctx context.Context, permission *domain.Permission) error
	GetByID(ctx context.Context, permissionID uuid.UUID) (*domain.Permission, error)
	GetByCode(ctx context.Context, code string) (*domain.Permission, error)
	List(ctx context.Context) ([]*domain.Permission, error)
	GetByService(ctx context.Context, service string) ([]*domain.Permission, error)
	GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*domain.Permission, error)
	GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]*domain.Permission, error)
	AssignPermissionsToRole(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error
	RemovePermissionsFromRole(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error
	ReplaceRolePermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error
	BatchCreate(ctx context.Context, permissions []*domain.Permission) error
}
