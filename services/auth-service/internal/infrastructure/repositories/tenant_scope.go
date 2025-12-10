package repositories

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func TenantScope(orgID uuid.UUID) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if orgID != uuid.Nil {
			return db.Where("organization_id = ?", orgID)
		}
		return db
	}
}

func WithTenantScope(db *gorm.DB, orgID uuid.UUID) *gorm.DB {
	if orgID != uuid.Nil {
		return db.Scopes(TenantScope(orgID))
	}
	return db
}
