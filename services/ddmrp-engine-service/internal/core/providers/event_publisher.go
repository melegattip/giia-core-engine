package providers

import (
	"context"

	"github.com/melegattip/giia-core-engine/services/ddmrp-engine-service/internal/core/domain"
)

type EventPublisher interface {
	PublishBufferCalculated(ctx context.Context, buffer *domain.Buffer) error
	PublishBufferStatusChanged(ctx context.Context, buffer *domain.Buffer, oldZone domain.ZoneType) error
	PublishBufferAlertTriggered(ctx context.Context, buffer *domain.Buffer) error
	PublishFADCreated(ctx context.Context, fad *domain.DemandAdjustment) error
	PublishFADUpdated(ctx context.Context, fad *domain.DemandAdjustment) error
	PublishFADDeleted(ctx context.Context, fadID string) error
}
