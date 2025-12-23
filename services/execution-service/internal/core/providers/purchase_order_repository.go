package providers

import (
	"context"

	"github.com/giia/giia-core-engine/services/execution-service/internal/core/domain"
	"github.com/google/uuid"
)

type PurchaseOrderRepository interface {
	Create(ctx context.Context, po *domain.PurchaseOrder) error
	GetByID(ctx context.Context, id, organizationID uuid.UUID) (*domain.PurchaseOrder, error)
	GetByPONumber(ctx context.Context, poNumber string, organizationID uuid.UUID) (*domain.PurchaseOrder, error)
	Update(ctx context.Context, po *domain.PurchaseOrder) error
	Delete(ctx context.Context, id, organizationID uuid.UUID) error
	List(ctx context.Context, organizationID uuid.UUID, filters map[string]interface{}, page, pageSize int) ([]*domain.PurchaseOrder, int64, error)
	GetDelayedOrders(ctx context.Context, organizationID uuid.UUID) ([]*domain.PurchaseOrder, error)
}