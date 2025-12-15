package providers

import (
	"context"

	"github.com/giia/giia-core-engine/services/catalog-service/internal/core/domain"
	"github.com/google/uuid"
)

type SupplierRepository interface {
	Create(ctx context.Context, supplier *domain.Supplier) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Supplier, error)
	GetByCode(ctx context.Context, code string) (*domain.Supplier, error)
	Update(ctx context.Context, supplier *domain.Supplier) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filters map[string]interface{}, page, pageSize int) ([]*domain.Supplier, int64, error)
}
