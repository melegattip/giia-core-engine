package purchase_order

import (
	"context"

	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/providers"
	"github.com/google/uuid"
)

type CancelPOUseCase struct {
	poRepo         providers.PurchaseOrderRepository
	eventPublisher providers.EventPublisher
}

func NewCancelPOUseCase(
	poRepo providers.PurchaseOrderRepository,
	publisher providers.EventPublisher,
) *CancelPOUseCase {
	return &CancelPOUseCase{
		poRepo:         poRepo,
		eventPublisher: publisher,
	}
}

type CancelPOInput struct {
	POID           uuid.UUID
	OrganizationID uuid.UUID
}

func (uc *CancelPOUseCase) Execute(ctx context.Context, input *CancelPOInput) (*domain.PurchaseOrder, error) {
	if input == nil {
		return nil, domain.NewValidationError("input cannot be nil")
	}
	if input.POID == uuid.Nil {
		return nil, domain.NewValidationError("po_id is required")
	}
	if input.OrganizationID == uuid.Nil {
		return nil, domain.NewValidationError("organization_id is required")
	}

	po, err := uc.poRepo.GetByID(ctx, input.POID, input.OrganizationID)
	if err != nil {
		return nil, err
	}

	if err := po.Cancel(); err != nil {
		return nil, err
	}

	if err := uc.poRepo.Update(ctx, po); err != nil {
		return nil, err
	}

	uc.eventPublisher.PublishPOCancelled(ctx, po)

	return po, nil
}