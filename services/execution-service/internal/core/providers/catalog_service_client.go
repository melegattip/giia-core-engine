package providers

import (
	"context"

	"github.com/google/uuid"
)

type Product struct {
	ID             uuid.UUID
	SKU            string
	Name           string
	UnitOfMeasure  string
	OrganizationID uuid.UUID
}

type Supplier struct {
	ID             uuid.UUID
	Name           string
	Code           string
	OrganizationID uuid.UUID
}

type CatalogServiceClient interface {
	GetProduct(ctx context.Context, productID uuid.UUID) (*Product, error)
	GetSupplier(ctx context.Context, supplierID uuid.UUID) (*Supplier, error)
	GetProductsByIDs(ctx context.Context, productIDs []uuid.UUID) ([]*Product, error)
}