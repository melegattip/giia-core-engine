package domain

import (
	"time"

	"github.com/google/uuid"
)

type DemandAdjustment struct {
	ID             uuid.UUID            `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ProductID      uuid.UUID            `json:"product_id" gorm:"type:uuid;not null;index:idx_demand_adj_product"`
	OrganizationID uuid.UUID            `json:"organization_id" gorm:"type:uuid;not null;index:idx_demand_adj_product,idx_demand_adj_org"`
	StartDate      time.Time            `json:"start_date" gorm:"type:date;not null;index:idx_demand_adj_dates,idx_demand_adj_active"`
	EndDate        time.Time            `json:"end_date" gorm:"type:date;not null;index:idx_demand_adj_dates,idx_demand_adj_active"`
	AdjustmentType DemandAdjustmentType `json:"adjustment_type" gorm:"type:varchar(30);not null"`
	Factor         float64              `json:"factor" gorm:"type:decimal(5,2);not null"`
	Reason         string               `json:"reason" gorm:"type:text;not null"`
	CreatedAt      time.Time            `json:"created_at" gorm:"not null;default:CURRENT_TIMESTAMP"`
	CreatedBy      uuid.UUID            `json:"created_by" gorm:"type:uuid;not null"`
}

type DemandAdjustmentType string

const (
	DemandAdjustmentFAD         DemandAdjustmentType = "fad"
	DemandAdjustmentSeasonal    DemandAdjustmentType = "seasonal"
	DemandAdjustmentNewProduct  DemandAdjustmentType = "new_product"
	DemandAdjustmentDiscontinue DemandAdjustmentType = "discontinue"
	DemandAdjustmentPromotion   DemandAdjustmentType = "promotion"
)

func (DemandAdjustment) TableName() string {
	return "demand_adjustments"
}

func (t DemandAdjustmentType) IsValid() bool {
	switch t {
	case DemandAdjustmentFAD, DemandAdjustmentSeasonal, DemandAdjustmentNewProduct,
		DemandAdjustmentDiscontinue, DemandAdjustmentPromotion:
		return true
	}
	return false
}

func NewDemandAdjustment(
	productID, orgID, createdBy uuid.UUID,
	startDate, endDate time.Time,
	adjustmentType DemandAdjustmentType,
	factor float64,
	reason string,
) (*DemandAdjustment, error) {
	if productID == uuid.Nil {
		return nil, NewValidationError("product_id is required")
	}
	if orgID == uuid.Nil {
		return nil, NewValidationError("organization_id is required")
	}
	if createdBy == uuid.Nil {
		return nil, NewValidationError("created_by is required")
	}
	if !adjustmentType.IsValid() {
		return nil, NewValidationError("invalid adjustment type")
	}
	if factor < 0 {
		return nil, NewValidationError("factor must be non-negative")
	}
	if endDate.Before(startDate) {
		return nil, NewValidationError("end_date must be >= start_date")
	}
	if reason == "" {
		return nil, NewValidationError("reason is required")
	}

	return &DemandAdjustment{
		ID:             uuid.New(),
		ProductID:      productID,
		OrganizationID: orgID,
		StartDate:      startDate,
		EndDate:        endDate,
		AdjustmentType: adjustmentType,
		Factor:         factor,
		Reason:         reason,
		CreatedAt:      time.Now(),
		CreatedBy:      createdBy,
	}, nil
}

func (da *DemandAdjustment) IsActive(date time.Time) bool {
	dateOnly := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	startOnly := time.Date(da.StartDate.Year(), da.StartDate.Month(), da.StartDate.Day(), 0, 0, 0, 0, time.UTC)
	endOnly := time.Date(da.EndDate.Year(), da.EndDate.Month(), da.EndDate.Day(), 0, 0, 0, 0, time.UTC)

	return !dateOnly.Before(startOnly) && !dateOnly.After(endOnly)
}

func (da *DemandAdjustment) Validate() error {
	if da.ProductID == uuid.Nil {
		return NewValidationError("product_id is required")
	}
	if da.OrganizationID == uuid.Nil {
		return NewValidationError("organization_id is required")
	}
	if !da.AdjustmentType.IsValid() {
		return NewValidationError("invalid adjustment type")
	}
	if da.Factor < 0 {
		return NewValidationError("factor must be non-negative")
	}
	if da.EndDate.Before(da.StartDate) {
		return NewValidationError("end_date must be >= start_date")
	}
	if da.Reason == "" {
		return NewValidationError("reason is required")
	}
	return nil
}
