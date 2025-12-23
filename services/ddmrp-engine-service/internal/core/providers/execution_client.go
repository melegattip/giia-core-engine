package providers

import (
	"context"

	"github.com/google/uuid"
)

type InventoryLevel struct {
	ProductID      uuid.UUID
	OrganizationID uuid.UUID
	OnHand         float64
	OnOrder        float64
}

type DemandForecast struct {
	ProductID      uuid.UUID
	OrganizationID uuid.UUID
	QualifiedDemand float64
}

type ExecutionServiceClient interface {
	GetInventoryLevel(ctx context.Context, productID, organizationID uuid.UUID) (*InventoryLevel, error)
	GetDemandForecast(ctx context.Context, productID, organizationID uuid.UUID) (*DemandForecast, error)
}
