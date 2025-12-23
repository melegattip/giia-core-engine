package providers

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type SalesData struct {
	OrganizationID uuid.UUID
	StartDate      time.Time
	EndDate        time.Time
	TotalValue     float64
	OrderCount     int
}

type InventorySnapshot struct {
	Date       time.Time
	TotalValue float64
	ProductID  *uuid.UUID
}

type ProductSales struct {
	ProductID     uuid.UUID
	SKU           string
	Name          string
	Sales30Days   float64
	AvgStockValue float64
}

type ExecutionServiceClient interface {
	GetSalesData(ctx context.Context, organizationID uuid.UUID, startDate, endDate time.Time) (*SalesData, error)
	GetInventorySnapshots(ctx context.Context, organizationID uuid.UUID, startDate, endDate time.Time) ([]*InventorySnapshot, error)
	GetProductSales(ctx context.Context, organizationID uuid.UUID, startDate, endDate time.Time) ([]*ProductSales, error)
}
