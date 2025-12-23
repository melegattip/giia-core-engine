package providers

import (
	"context"

	"github.com/giia/giia-core-engine/services/execution-service/internal/core/domain"
)

type EventPublisher interface {
	PublishPOCreated(ctx context.Context, po *domain.PurchaseOrder) error
	PublishPOUpdated(ctx context.Context, po *domain.PurchaseOrder) error
	PublishPOReceived(ctx context.Context, po *domain.PurchaseOrder) error
	PublishPOCancelled(ctx context.Context, po *domain.PurchaseOrder) error
	PublishSOCreated(ctx context.Context, so *domain.SalesOrder) error
	PublishSOUpdated(ctx context.Context, so *domain.SalesOrder) error
	PublishSOCancelled(ctx context.Context, so *domain.SalesOrder) error
	PublishDeliveryNoteIssued(ctx context.Context, so *domain.SalesOrder) error
	PublishInventoryUpdated(ctx context.Context, txn *domain.InventoryTransaction) error
	PublishAlertCreated(ctx context.Context, alert *domain.Alert) error
}