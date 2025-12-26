package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type Organization struct {
	ID        uuid.UUID          `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name      string             `json:"name" gorm:"type:varchar(255);not null"`
	Slug      string             `json:"slug" gorm:"type:varchar(100);unique;not null"`
	Status    OrganizationStatus `json:"status" gorm:"type:varchar(20);not null;default:'active'"`
	Settings  datatypes.JSON     `json:"settings" gorm:"type:jsonb;default:'{}'"`
	CreatedAt time.Time          `json:"created_at" gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time          `json:"updated_at" gorm:"not null;default:CURRENT_TIMESTAMP"`
}

type OrganizationStatus string

const (
	OrganizationStatusActive    OrganizationStatus = "active"
	OrganizationStatusSuspended OrganizationStatus = "suspended"
	OrganizationStatusInactive  OrganizationStatus = "inactive"
)

func (Organization) TableName() string {
	return "organizations"
}
