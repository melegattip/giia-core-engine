package domain

import (
	"time"

	"github.com/google/uuid"
)

type InventoryRotationKPI struct {
	ID                   uuid.UUID
	OrganizationID       uuid.UUID
	SnapshotDate         time.Time
	SalesLast30Days      float64
	AvgMonthlyStock      float64
	RotationRatio        float64
	TopRotatingProducts  []RotatingProduct
	SlowRotatingProducts []RotatingProduct
	CreatedAt            time.Time
}

func NewInventoryRotationKPI(
	organizationID uuid.UUID,
	snapshotDate time.Time,
	salesLast30Days float64,
	avgMonthlyStock float64,
	topRotatingProducts []RotatingProduct,
	slowRotatingProducts []RotatingProduct,
) (*InventoryRotationKPI, error) {
	if organizationID == uuid.Nil {
		return nil, NewValidationError("organization_id is required")
	}
	if snapshotDate.IsZero() {
		return nil, NewValidationError("snapshot_date is required")
	}
	if salesLast30Days < 0 {
		return nil, NewValidationError("sales_last_30_days cannot be negative")
	}
	if avgMonthlyStock < 0 {
		return nil, NewValidationError("avg_monthly_stock cannot be negative")
	}

	rotationRatio := 0.0
	if avgMonthlyStock > 0 {
		rotationRatio = salesLast30Days / avgMonthlyStock
	}

	if topRotatingProducts == nil {
		topRotatingProducts = []RotatingProduct{}
	}
	if slowRotatingProducts == nil {
		slowRotatingProducts = []RotatingProduct{}
	}

	return &InventoryRotationKPI{
		ID:                   uuid.New(),
		OrganizationID:       organizationID,
		SnapshotDate:         snapshotDate,
		SalesLast30Days:      salesLast30Days,
		AvgMonthlyStock:      avgMonthlyStock,
		RotationRatio:        rotationRatio,
		TopRotatingProducts:  topRotatingProducts,
		SlowRotatingProducts: slowRotatingProducts,
		CreatedAt:            time.Now().UTC(),
	}, nil
}

type RotatingProduct struct {
	ProductID     uuid.UUID
	SKU           string
	Name          string
	Sales30Days   float64
	AvgStockValue float64
	RotationRatio float64
}

func CalculateProductRotation(sales30Days float64, avgStockValue float64) float64 {
	if avgStockValue <= 0 {
		return 0.0
	}
	return sales30Days / avgStockValue
}

func NewRotatingProduct(
	productID uuid.UUID,
	sku string,
	name string,
	sales30Days float64,
	avgStockValue float64,
) RotatingProduct {
	return RotatingProduct{
		ProductID:     productID,
		SKU:           sku,
		Name:          name,
		Sales30Days:   sales30Days,
		AvgStockValue: avgStockValue,
		RotationRatio: CalculateProductRotation(sales30Days, avgStockValue),
	}
}
