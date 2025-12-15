package domain

import (
	"time"

	"github.com/giia/giia-core-engine/pkg/errors"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type SupplierStatus string

const (
	SupplierStatusActive   SupplierStatus = "active"
	SupplierStatusInactive SupplierStatus = "inactive"
)

type Supplier struct {
	ID                uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Code              string         `json:"code" gorm:"type:varchar(100);not null;index:idx_suppliers_code"`
	Name              string         `json:"name" gorm:"type:varchar(255);not null"`
	LeadTimeDays      int            `json:"lead_time_days" gorm:"not null;default:0"`
	ReliabilityRating int            `json:"reliability_rating" gorm:"default:80"`
	ContactInfo       datatypes.JSON `json:"contact_info,omitempty" gorm:"type:jsonb"`
	Status            SupplierStatus `json:"status" gorm:"type:varchar(20);not null;default:'active';index:idx_suppliers_status"`
	OrganizationID    uuid.UUID      `json:"organization_id" gorm:"type:uuid;not null;index:idx_suppliers_organization_id"`
	CreatedAt         time.Time      `json:"created_at" gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt         time.Time      `json:"updated_at" gorm:"not null;default:CURRENT_TIMESTAMP"`
}

func (Supplier) TableName() string {
	return "suppliers"
}

func (s *Supplier) Validate() error {
	if s.Code == "" {
		return errors.NewBadRequest("supplier code is required")
	}
	if len(s.Code) > 100 {
		return errors.NewBadRequest("supplier code must be 100 characters or less")
	}
	if s.Name == "" {
		return errors.NewBadRequest("supplier name is required")
	}
	if len(s.Name) > 255 {
		return errors.NewBadRequest("supplier name must be 255 characters or less")
	}
	if s.LeadTimeDays < 0 {
		return errors.NewBadRequest("lead time days cannot be negative")
	}
	if s.ReliabilityRating < 0 || s.ReliabilityRating > 100 {
		return errors.NewBadRequest("reliability rating must be between 0 and 100")
	}
	if s.OrganizationID == uuid.Nil {
		return errors.NewBadRequest("organization ID is required")
	}
	if !s.IsValidStatus() {
		return errors.NewBadRequest("invalid supplier status")
	}
	return nil
}

func (s *Supplier) IsActive() bool {
	return s.Status == SupplierStatusActive
}

func (s *Supplier) IsValidStatus() bool {
	switch s.Status {
	case SupplierStatusActive, SupplierStatusInactive:
		return true
	default:
		return false
	}
}

func (s *Supplier) Deactivate() {
	s.Status = SupplierStatusInactive
}

func (s *Supplier) Activate() {
	s.Status = SupplierStatusActive
}
