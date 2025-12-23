package domain

import (
	"time"

	"github.com/google/uuid"
)

type BufferAdjustment struct {
	ID             uuid.UUID            `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	BufferID       uuid.UUID            `json:"buffer_id" gorm:"type:uuid;not null;index:idx_buffer_adj_buffer,idx_buffer_adj_active"`
	ProductID      uuid.UUID            `json:"product_id" gorm:"type:uuid;not null;index:idx_buffer_adj_product"`
	OrganizationID uuid.UUID            `json:"organization_id" gorm:"type:uuid;not null;index:idx_buffer_adj_product"`
	AdjustmentType BufferAdjustmentType `json:"adjustment_type" gorm:"type:varchar(30);not null"`
	TargetZone     ZoneType             `json:"target_zone" gorm:"type:varchar(20);not null"`
	Factor         float64              `json:"factor" gorm:"type:decimal(5,2);not null"`
	StartDate      time.Time            `json:"start_date" gorm:"type:date;not null;index:idx_buffer_adj_dates,idx_buffer_adj_active"`
	EndDate        time.Time            `json:"end_date" gorm:"type:date;not null;index:idx_buffer_adj_dates,idx_buffer_adj_active"`
	Reason         string               `json:"reason" gorm:"type:text;not null"`
	CreatedAt      time.Time            `json:"created_at" gorm:"not null;default:CURRENT_TIMESTAMP"`
	CreatedBy      uuid.UUID            `json:"created_by" gorm:"type:uuid;not null"`
}

type BufferAdjustmentType string

const (
	BufferAdjustmentZoneFactor      BufferAdjustmentType = "zone_factor"
	BufferAdjustmentPlannedEvent    BufferAdjustmentType = "planned_event"
	BufferAdjustmentSpikeManagement BufferAdjustmentType = "spike_management"
	BufferAdjustmentSeasonalPrepare BufferAdjustmentType = "seasonal_prepare"
)

func (BufferAdjustment) TableName() string {
	return "buffer_adjustments"
}

func (t BufferAdjustmentType) IsValid() bool {
	switch t {
	case BufferAdjustmentZoneFactor, BufferAdjustmentPlannedEvent,
		BufferAdjustmentSpikeManagement, BufferAdjustmentSeasonalPrepare:
		return true
	}
	return false
}

func (ba *BufferAdjustment) IsActive(date time.Time) bool {
	dateOnly := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	startOnly := time.Date(ba.StartDate.Year(), ba.StartDate.Month(), ba.StartDate.Day(), 0, 0, 0, 0, time.UTC)
	endOnly := time.Date(ba.EndDate.Year(), ba.EndDate.Month(), ba.EndDate.Day(), 0, 0, 0, 0, time.UTC)

	return !dateOnly.Before(startOnly) && !dateOnly.After(endOnly)
}

func (ba *BufferAdjustment) Validate() error {
	if ba.BufferID == uuid.Nil {
		return NewValidationError("buffer_id is required")
	}
	if ba.ProductID == uuid.Nil {
		return NewValidationError("product_id is required")
	}
	if ba.OrganizationID == uuid.Nil {
		return NewValidationError("organization_id is required")
	}
	if !ba.AdjustmentType.IsValid() {
		return NewValidationError("invalid adjustment type")
	}
	if ba.Factor <= 0 {
		return NewValidationError("factor must be greater than 0")
	}
	if ba.EndDate.Before(ba.StartDate) {
		return NewValidationError("end_date must be >= start_date")
	}
	if ba.Reason == "" {
		return NewValidationError("reason is required")
	}
	return nil
}
