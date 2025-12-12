package providers

import (
	"context"

	"github.com/google/uuid"

	"github.com/giia/giia-core-engine/services/auth-service/internal/core/domain"
)

type RoleRepository interface {
	Create(ctx context.Context, role *domain.Role) error
	GetByID(ctx context.Context, roleID uuid.UUID) (*domain.Role, error)
	GetByName(ctx context.Context, name string, orgID *uuid.UUID) (*domain.Role, error)
	GetWithPermissions(ctx context.Context, roleID uuid.UUID) (*domain.Role, error)
	Update(ctx context.Context, role *domain.Role) error
	Delete(ctx context.Context, roleID uuid.UUID) error
	List(ctx context.Context, orgID *uuid.UUID) ([]*domain.Role, error)
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error)
	AssignRoleToUser(ctx context.Context, userID, roleID, assignedBy uuid.UUID) error
	RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error
	GetUsersWithRole(ctx context.Context, roleID uuid.UUID) ([]uuid.UUID, error)
}
