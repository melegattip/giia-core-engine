package providers

import (
	"context"
	"time"

	"github.com/giia/giia-core-engine/services/ddmrp-engine-service/internal/core/domain"
	"github.com/google/uuid"
)

type ADURepository interface {
	Create(ctx context.Context, adu *domain.ADUCalculation) error
	GetLatest(ctx context.Context, productID, organizationID uuid.UUID) (*domain.ADUCalculation, error)
	GetByDate(ctx context.Context, productID, organizationID uuid.UUID, date time.Time) (*domain.ADUCalculation, error)
	ListHistory(ctx context.Context, productID, organizationID uuid.UUID, limit int) ([]domain.ADUCalculation, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
