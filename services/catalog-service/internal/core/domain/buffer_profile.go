package domain

import (
	"time"

	"github.com/giia/giia-core-engine/pkg/errors"
	"github.com/google/uuid"
)

type BufferProfile struct {
	ID                 uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name               string    `json:"name" gorm:"type:varchar(100);not null;index:idx_buffer_profiles_name"`
	Description        string    `json:"description" gorm:"type:text"`
	LeadTimeFactor     float64   `json:"lead_time_factor" gorm:"type:decimal(5,2);not null;default:1.0"`
	VariabilityFactor  float64   `json:"variability_factor" gorm:"type:decimal(5,2);not null;default:1.0"`
	TargetServiceLevel int       `json:"target_service_level" gorm:"not null;default:95"`
	OrganizationID     uuid.UUID `json:"organization_id" gorm:"type:uuid;not null;index:idx_buffer_profiles_organization_id"`
	CreatedAt          time.Time `json:"created_at" gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt          time.Time `json:"updated_at" gorm:"not null;default:CURRENT_TIMESTAMP"`
}

func (BufferProfile) TableName() string {
	return "buffer_profiles"
}

func (bp *BufferProfile) Validate() error {
	if bp.Name == "" {
		return errors.NewBadRequest("buffer profile name is required")
	}
	if len(bp.Name) > 100 {
		return errors.NewBadRequest("buffer profile name must be 100 characters or less")
	}
	if bp.LeadTimeFactor <= 0 {
		return errors.NewBadRequest("lead time factor must be greater than 0")
	}
	if bp.VariabilityFactor <= 0 {
		return errors.NewBadRequest("variability factor must be greater than 0")
	}
	if bp.TargetServiceLevel < 0 || bp.TargetServiceLevel > 100 {
		return errors.NewBadRequest("target service level must be between 0 and 100")
	}
	if bp.OrganizationID == uuid.Nil {
		return errors.NewBadRequest("organization ID is required")
	}
	return nil
}
