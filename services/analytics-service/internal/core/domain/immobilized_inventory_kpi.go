package domain

import (
	"time"

	"github.com/google/uuid"
)

type ImmobilizedInventoryKPI struct {
	ID                    uuid.UUID
	OrganizationID        uuid.UUID
	SnapshotDate          time.Time
	ThresholdYears        int
	ImmobilizedCount      int
	ImmobilizedValue      float64
	TotalStockValue       float64
	ImmobilizedPercentage float64
	CreatedAt             time.Time
}

func NewImmobilizedInventoryKPI(
	organizationID uuid.UUID,
	snapshotDate time.Time,
	thresholdYears int,
	immobilizedCount int,
	immobilizedValue float64,
	totalStockValue float64,
) (*ImmobilizedInventoryKPI, error) {
	if organizationID == uuid.Nil {
		return nil, NewValidationError("organization_id is required")
	}
	if snapshotDate.IsZero() {
		return nil, NewValidationError("snapshot_date is required")
	}
	if thresholdYears <= 0 {
		return nil, NewValidationError("threshold_years must be positive")
	}
	if immobilizedCount < 0 {
		return nil, NewValidationError("immobilized_count cannot be negative")
	}
	if immobilizedValue < 0 {
		return nil, NewValidationError("immobilized_value cannot be negative")
	}
	if totalStockValue < 0 {
		return nil, NewValidationError("total_stock_value cannot be negative")
	}

	immobilizedPercentage := 0.0
	if totalStockValue > 0 {
		immobilizedPercentage = (immobilizedValue / totalStockValue) * 100
	}

	return &ImmobilizedInventoryKPI{
		ID:                    uuid.New(),
		OrganizationID:        organizationID,
		SnapshotDate:          snapshotDate,
		ThresholdYears:        thresholdYears,
		ImmobilizedCount:      immobilizedCount,
		ImmobilizedValue:      immobilizedValue,
		TotalStockValue:       totalStockValue,
		ImmobilizedPercentage: immobilizedPercentage,
		CreatedAt:             time.Now().UTC(),
	}, nil
}

type ImmobilizedProduct struct {
	ProductID    uuid.UUID
	SKU          string
	Name         string
	Category     string
	Quantity     float64
	PurchaseDate time.Time
	YearsInStock float64
	UnitCost     float64
	TotalValue   float64
	LastSaleDate *time.Time
}

func CalculateYearsInStock(purchaseDate time.Time, currentDate time.Time) float64 {
	daysDiff := currentDate.Sub(purchaseDate).Hours() / 24
	return daysDiff / 365.0
}

func IsImmobilized(purchaseDate time.Time, currentDate time.Time, thresholdYears int) bool {
	yearsInStock := CalculateYearsInStock(purchaseDate, currentDate)
	return yearsInStock >= float64(thresholdYears)
}
