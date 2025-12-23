package providers

import (
	"context"

	"github.com/melegattip/giia-core-engine/services/catalog-service/internal/core/domain"
	"github.com/google/uuid"
)

type ProductRepository interface {
	Create(ctx context.Context, product *domain.Product) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Product, error)
	GetByIDWithSuppliers(ctx context.Context, id uuid.UUID) (*domain.Product, error)
	GetBySKU(ctx context.Context, sku string) (*domain.Product, error)
	Update(ctx context.Context, product *domain.Product) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filters map[string]interface{}, page, pageSize int) ([]*domain.Product, int64, error)
	Search(ctx context.Context, query string, filters map[string]interface{}, page, pageSize int) ([]*domain.Product, int64, error)
	AssociateSupplier(ctx context.Context, productSupplier *domain.ProductSupplier) error
	RemoveSupplier(ctx context.Context, productID, supplierID uuid.UUID) error
	GetProductSuppliers(ctx context.Context, productID uuid.UUID) ([]*domain.ProductSupplier, error)
}
