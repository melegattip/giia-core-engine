package providers

import (
	"context"
	"time"

	"github.com/melegattip/giia-core-engine/services/ddmrp-engine-service/internal/core/domain"
	"github.com/google/uuid"
)

type BufferAdjustmentRepository interface {
	Create(ctx context.Context, adjustment *domain.BufferAdjustment) error
	Update(ctx context.Context, adjustment *domain.BufferAdjustment) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.BufferAdjustment, error)
	GetActiveForDate(ctx context.Context, bufferID uuid.UUID, date time.Time) ([]domain.BufferAdjustment, error)
	ListByBuffer(ctx context.Context, bufferID uuid.UUID) ([]domain.BufferAdjustment, error)
	ListByProduct(ctx context.Context, productID, organizationID uuid.UUID) ([]domain.BufferAdjustment, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
