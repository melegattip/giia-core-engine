package providers

import (
	"context"

	"github.com/giia/giia-core-engine/services/execution-service/internal/core/domain"
	"github.com/google/uuid"
)

type InventoryBalanceRepository interface {
	GetOrCreate(ctx context.Context, organizationID, productID, locationID uuid.UUID) (*domain.InventoryBalance, error)
	UpdateOnHand(ctx context.Context, organizationID, productID, locationID uuid.UUID, quantity float64) error
	UpdateReserved(ctx context.Context, organizationID, productID, locationID uuid.UUID, quantity float64) error
	GetByProduct(ctx context.Context, organizationID, productID uuid.UUID) ([]*domain.InventoryBalance, error)
	GetByLocation(ctx context.Context, organizationID, locationID uuid.UUID) ([]*domain.InventoryBalance, error)
}