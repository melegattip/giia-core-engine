package domain

import (
	"time"

	"github.com/google/uuid"
)

type DaysInInventoryKPI struct {
	ID                uuid.UUID
	OrganizationID    uuid.UUID
	SnapshotDate      time.Time
	TotalValuedDays   float64
	AverageValuedDays float64
	TotalProducts     int
	CreatedAt         time.Time
}

func NewDaysInInventoryKPI(
	organizationID uuid.UUID,
	snapshotDate time.Time,
	totalValuedDays float64,
	averageValuedDays float64,
	totalProducts int,
) (*DaysInInventoryKPI, error) {
	if organizationID == uuid.Nil {
		return nil, NewValidationError("organization_id is required")
	}
	if snapshotDate.IsZero() {
		return nil, NewValidationError("snapshot_date is required")
	}
	if totalValuedDays < 0 {
		return nil, NewValidationError("total_valued_days cannot be negative")
	}
	if averageValuedDays < 0 {
		return nil, NewValidationError("average_valued_days cannot be negative")
	}
	if totalProducts < 0 {
		return nil, NewValidationError("total_products cannot be negative")
	}

	return &DaysInInventoryKPI{
		ID:                uuid.New(),
		OrganizationID:    organizationID,
		SnapshotDate:      snapshotDate,
		TotalValuedDays:   totalValuedDays,
		AverageValuedDays: averageValuedDays,
		TotalProducts:     totalProducts,
		CreatedAt:         time.Now().UTC(),
	}, nil
}

type ProductInventoryAge struct {
	ProductID       uuid.UUID
	SKU             string
	Name            string
	Quantity        float64
	PurchaseDate    time.Time
	DaysInInventory int
	UnitCost        float64
	TotalValue      float64
	ValuedDays      float64
}

func CalculateValuedDays(product ProductInventoryAge, currentDate time.Time) float64 {
	daysInStock := int(currentDate.Sub(product.PurchaseDate).Hours() / 24)
	if daysInStock < 0 {
		daysInStock = 0
	}
	totalValue := product.Quantity * product.UnitCost
	return float64(daysInStock) * totalValue
}
