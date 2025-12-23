package providers

import (
	"context"

	"github.com/giia/giia-core-engine/services/execution-service/internal/core/domain"
	"github.com/google/uuid"
)

type SalesOrderRepository interface {
	Create(ctx context.Context, so *domain.SalesOrder) error
	GetByID(ctx context.Context, id, organizationID uuid.UUID) (*domain.SalesOrder, error)
	GetBySONumber(ctx context.Context, soNumber string, organizationID uuid.UUID) (*domain.SalesOrder, error)
	Update(ctx context.Context, so *domain.SalesOrder) error
	Delete(ctx context.Context, id, organizationID uuid.UUID) error
	List(ctx context.Context, organizationID uuid.UUID, filters map[string]interface{}, page, pageSize int) ([]*domain.SalesOrder, int64, error)
	GetQualifiedDemand(ctx context.Context, organizationID, productID uuid.UUID) (float64, error)
}