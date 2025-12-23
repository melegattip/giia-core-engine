package providers

import (
	"context"
	"time"

	"github.com/giia/giia-core-engine/services/ddmrp-engine-service/internal/core/domain"
	"github.com/google/uuid"
)

type BufferHistoryRepository interface {
	Create(ctx context.Context, history *domain.BufferHistory) error
	GetByBufferAndDate(ctx context.Context, bufferID uuid.UUID, date time.Time) (*domain.BufferHistory, error)
	ListByBuffer(ctx context.Context, bufferID uuid.UUID, limit int) ([]domain.BufferHistory, error)
	ListByProduct(ctx context.Context, productID, organizationID uuid.UUID, startDate, endDate time.Time) ([]domain.BufferHistory, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
