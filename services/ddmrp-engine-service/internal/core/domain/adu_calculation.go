package domain

import (
	"time"

	"github.com/google/uuid"
)

type ADUCalculation struct {
	ID              uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ProductID       uuid.UUID `json:"product_id" gorm:"type:uuid;not null;uniqueIndex:uq_adu_product_date;index:idx_adu_product"`
	OrganizationID  uuid.UUID `json:"organization_id" gorm:"type:uuid;not null;uniqueIndex:uq_adu_product_date;index:idx_adu_product,idx_adu_org"`
	CalculationDate time.Time `json:"calculation_date" gorm:"type:date;not null;uniqueIndex:uq_adu_product_date;index:idx_adu_calc_date"`
	ADUValue        float64   `json:"adu_value" gorm:"type:decimal(15,2);not null"`
	Method          ADUMethod `json:"method" gorm:"type:varchar(20);not null"`
	PeriodDays      int       `json:"period_days" gorm:"not null;default:30"`
	CreatedAt       time.Time `json:"created_at" gorm:"not null;default:CURRENT_TIMESTAMP"`
}

type ADUMethod string

const (
	ADUMethodAverage     ADUMethod = "average"
	ADUMethodExponential ADUMethod = "exponential"
	ADUMethodWeighted    ADUMethod = "weighted"
)

func (ADUCalculation) TableName() string {
	return "adu_calculations"
}

func (m ADUMethod) IsValid() bool {
	switch m {
	case ADUMethodAverage, ADUMethodExponential, ADUMethodWeighted:
		return true
	}
	return false
}

func (ac *ADUCalculation) Validate() error {
	if ac.ProductID == uuid.Nil {
		return NewValidationError("product_id is required")
	}
	if ac.OrganizationID == uuid.Nil {
		return NewValidationError("organization_id is required")
	}
	if !ac.Method.IsValid() {
		return NewValidationError("invalid ADU method")
	}
	if ac.ADUValue < 0 {
		return NewValidationError("adu_value must be non-negative")
	}
	if ac.PeriodDays <= 0 {
		return NewValidationError("period_days must be greater than 0")
	}
	return nil
}
