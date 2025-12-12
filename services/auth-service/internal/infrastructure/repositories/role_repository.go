package repositories

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/giia/giia-core-engine/services/auth-service/internal/core/domain"
	"github.com/giia/giia-core-engine/services/auth-service/internal/core/providers"
)

type roleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) providers.RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) Create(ctx context.Context, role *domain.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *roleRepository) GetByID(ctx context.Context, roleID uuid.UUID) (*domain.Role, error) {
	var role domain.Role
	err := r.db.WithContext(ctx).
		Where("id = ?", roleID).
		First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) GetByName(ctx context.Context, name string, orgID *uuid.UUID) (*domain.Role, error) {
	var role domain.Role
	query := r.db.WithContext(ctx).Where("name = ?", name)

	if orgID != nil {
		query = query.Where("organization_id = ?", *orgID)
	} else {
		query = query.Where("organization_id IS NULL")
	}

	err := query.First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) GetWithPermissions(ctx context.Context, roleID uuid.UUID) (*domain.Role, error) {
	var role domain.Role
	err := r.db.WithContext(ctx).
		Preload("Permissions").
		Where("id = ?", roleID).
		First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) Update(ctx context.Context, role *domain.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

func (r *roleRepository) Delete(ctx context.Context, roleID uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Role{}, "id = ?", roleID).Error
}

func (r *roleRepository) List(ctx context.Context, orgID *uuid.UUID) ([]*domain.Role, error) {
	var roles []*domain.Role
	query := r.db.WithContext(ctx)

	if orgID != nil {
		query = query.Where("organization_id = ? OR is_system = ?", *orgID, true)
	} else {
		query = query.Where("is_system = ?", true)
	}

	err := query.Order("name ASC").Find(&roles).Error
	if err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *roleRepository) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error) {
	var roles []*domain.Role
	err := r.db.WithContext(ctx).
		Table("roles").
		Joins("INNER JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", userID).
		Preload("Permissions").
		Find(&roles).Error
	if err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *roleRepository) AssignRoleToUser(ctx context.Context, userID, roleID, assignedBy uuid.UUID) error {
	userRole := &domain.UserRole{
		UserID:     userID,
		RoleID:     roleID,
		AssignedBy: &assignedBy,
	}
	return r.db.WithContext(ctx).Create(userRole).Error
}

func (r *roleRepository) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND role_id = ?", userID, roleID).
		Delete(&domain.UserRole{}).Error
}

func (r *roleRepository) GetUsersWithRole(ctx context.Context, roleID uuid.UUID) ([]uuid.UUID, error) {
	var userIDs []uuid.UUID
	err := r.db.WithContext(ctx).
		Table("user_roles").
		Where("role_id = ?", roleID).
		Pluck("user_id", &userIDs).Error
	if err != nil {
		return nil, err
	}
	return userIDs, nil
}
