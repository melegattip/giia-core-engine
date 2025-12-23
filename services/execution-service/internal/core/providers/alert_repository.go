package providers

import (
	"context"

	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/domain"
	"github.com/google/uuid"
)

type AlertRepository interface {
	Create(ctx context.Context, alert *domain.Alert) error
	GetByID(ctx context.Context, id, organizationID uuid.UUID) (*domain.Alert, error)
	Update(ctx context.Context, alert *domain.Alert) error
	List(ctx context.Context, organizationID uuid.UUID, filters map[string]interface{}, page, pageSize int) ([]*domain.Alert, int64, error)
	GetActiveAlerts(ctx context.Context, organizationID uuid.UUID) ([]*domain.Alert, error)
	GetByResourceID(ctx context.Context, resourceType string, resourceID, organizationID uuid.UUID) ([]*domain.Alert, error)
}