package providers

import (
	"context"

	"github.com/melegattip/giia-core-engine/services/ddmrp-engine-service/internal/core/domain"
	"github.com/google/uuid"
)

type BufferRepository interface {
	Create(ctx context.Context, buffer *domain.Buffer) error
	Save(ctx context.Context, buffer *domain.Buffer) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Buffer, error)
	GetByProduct(ctx context.Context, productID, organizationID uuid.UUID) (*domain.Buffer, error)
	List(ctx context.Context, organizationID uuid.UUID, limit, offset int) ([]domain.Buffer, error)
	ListByZone(ctx context.Context, organizationID uuid.UUID, zone domain.ZoneType) ([]domain.Buffer, error)
	ListByAlertLevel(ctx context.Context, organizationID uuid.UUID, alertLevel domain.AlertLevel) ([]domain.Buffer, error)
	ListAll(ctx context.Context, organizationID uuid.UUID) ([]domain.Buffer, error)
	UpdateNFP(ctx context.Context, bufferID uuid.UUID, onHand, onOrder, qualifiedDemand float64) error
	Delete(ctx context.Context, id uuid.UUID) error
}
