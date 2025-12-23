package providers

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ProductWithInventory struct {
	ProductID        uuid.UUID
	SKU              string
	Name             string
	Category         string
	Quantity         float64
	StandardCost     float64
	LastPurchaseDate *time.Time
	LastSaleDate     *time.Time
}

type CatalogServiceClient interface {
	ListProductsWithInventory(ctx context.Context, organizationID uuid.UUID) ([]*ProductWithInventory, error)
	GetProduct(ctx context.Context, productID uuid.UUID) (*ProductWithInventory, error)
}
