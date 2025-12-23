package providers

import (
	"context"
	"time"

	"github.com/melegattip/giia-core-engine/services/ddmrp-engine-service/internal/core/domain"
	"github.com/google/uuid"
)

type DemandAdjustmentRepository interface {
	Create(ctx context.Context, adjustment *domain.DemandAdjustment) error
	Update(ctx context.Context, adjustment *domain.DemandAdjustment) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.DemandAdjustment, error)
	GetActiveForDate(ctx context.Context, productID, organizationID uuid.UUID, date time.Time) ([]domain.DemandAdjustment, error)
	ListByProduct(ctx context.Context, productID, organizationID uuid.UUID) ([]domain.DemandAdjustment, error)
	ListByOrganization(ctx context.Context, organizationID uuid.UUID, limit, offset int) ([]domain.DemandAdjustment, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
