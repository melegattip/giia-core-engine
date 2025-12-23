package providers

import (
	"context"

	"github.com/giia/giia-core-engine/services/execution-service/internal/core/domain"
	"github.com/google/uuid"
)

type InventoryTransactionRepository interface {
	Create(ctx context.Context, txn *domain.InventoryTransaction) error
	GetByID(ctx context.Context, id, organizationID uuid.UUID) (*domain.InventoryTransaction, error)
	List(ctx context.Context, organizationID, productID uuid.UUID, filters map[string]interface{}, page, pageSize int) ([]*domain.InventoryTransaction, int64, error)
	GetByReferenceID(ctx context.Context, referenceType string, referenceID, organizationID uuid.UUID) ([]*domain.InventoryTransaction, error)
}