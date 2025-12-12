package repositories

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/giia/giia-core-engine/services/auth-service/internal/core/domain"
	"github.com/giia/giia-core-engine/services/auth-service/internal/core/providers"
)

type permissionRepository struct {
	db *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) providers.PermissionRepository {
	return &permissionRepository{db: db}
}

func (r *permissionRepository) Create(ctx context.Context, permission *domain.Permission) error {
	return r.db.WithContext(ctx).Create(permission).Error
}

func (r *permissionRepository) GetByID(ctx context.Context, permissionID uuid.UUID) (*domain.Permission, error) {
	var permission domain.Permission
	err := r.db.WithContext(ctx).
		Where("id = ?", permissionID).
		First(&permission).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

func (r *permissionRepository) GetByCode(ctx context.Context, code string) (*domain.Permission, error) {
	var permission domain.Permission
	err := r.db.WithContext(ctx).
		Where("code = ?", code).
		First(&permission).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

func (r *permissionRepository) List(ctx context.Context) ([]*domain.Permission, error) {
	var permissions []*domain.Permission
	err := r.db.WithContext(ctx).
		Order("service ASC, resource ASC, action ASC").
		Find(&permissions).Error
	if err != nil {
		return nil, err
	}
	return permissions, nil
}

func (r *permissionRepository) GetByService(ctx context.Context, service string) ([]*domain.Permission, error) {
	var permissions []*domain.Permission
	err := r.db.WithContext(ctx).
		Where("service = ?", service).
		Order("resource ASC, action ASC").
		Find(&permissions).Error
	if err != nil {
		return nil, err
	}
	return permissions, nil
}

func (r *permissionRepository) GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*domain.Permission, error) {
	var permissions []*domain.Permission
	err := r.db.WithContext(ctx).
		Table("permissions").
		Joins("INNER JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Where("role_permissions.role_id = ?", roleID).
		Find(&permissions).Error
	if err != nil {
		return nil, err
	}
	return permissions, nil
}

func (r *permissionRepository) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]*domain.Permission, error) {
	var permissions []*domain.Permission
	err := r.db.WithContext(ctx).
		Table("permissions").
		Joins("INNER JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Joins("INNER JOIN user_roles ON user_roles.role_id = role_permissions.role_id").
		Where("user_roles.user_id = ?", userID).
		Distinct("permissions.id", "permissions.code", "permissions.description", "permissions.service", "permissions.resource", "permissions.action", "permissions.created_at").
		Find(&permissions).Error
	if err != nil {
		return nil, err
	}
	return permissions, nil
}

func (r *permissionRepository) AssignPermissionsToRole(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	if len(permissionIDs) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, permissionID := range permissionIDs {
			result := tx.Exec(
				"INSERT INTO role_permissions (role_id, permission_id) VALUES (?, ?) ON CONFLICT (role_id, permission_id) DO NOTHING",
				roleID, permissionID,
			)
			if result.Error != nil {
				return result.Error
			}
		}
		return nil
	})
}

func (r *permissionRepository) RemovePermissionsFromRole(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	if len(permissionIDs) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).
		Where("role_id = ? AND permission_id IN ?", roleID, permissionIDs).
		Delete(&struct {
			RoleID       uuid.UUID `gorm:"column:role_id"`
			PermissionID uuid.UUID `gorm:"column:permission_id"`
		}{}).Error
}

func (r *permissionRepository) ReplaceRolePermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM role_permissions WHERE role_id = ?", roleID).Error; err != nil {
			return err
		}

		if len(permissionIDs) == 0 {
			return nil
		}

		for _, permissionID := range permissionIDs {
			if err := tx.Exec(
				"INSERT INTO role_permissions (role_id, permission_id) VALUES (?, ?)",
				roleID, permissionID,
			).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *permissionRepository) BatchCreate(ctx context.Context, permissions []*domain.Permission) error {
	if len(permissions) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).CreateInBatches(permissions, 100).Error
}
