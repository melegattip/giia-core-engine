package demand_adjustment

import (
	"context"

	"github.com/melegattip/giia-core-engine/pkg/errors"
	"github.com/melegattip/giia-core-engine/services/ddmrp-engine-service/internal/core/providers"
	"github.com/google/uuid"
)

type DeleteFADUseCase struct {
	demandAdjustmentRepo providers.DemandAdjustmentRepository
	eventPublisher       providers.EventPublisher
}

func NewDeleteFADUseCase(
	repo providers.DemandAdjustmentRepository,
	publisher providers.EventPublisher,
) *DeleteFADUseCase {
	return &DeleteFADUseCase{
		demandAdjustmentRepo: repo,
		eventPublisher:       publisher,
	}
}

func (uc *DeleteFADUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return errors.NewBadRequest("id is required")
	}

	_, err := uc.demandAdjustmentRepo.GetByID(ctx, id)
	if err != nil {
		return errors.NewNotFound("demand adjustment not found")
	}

	if err := uc.demandAdjustmentRepo.Delete(ctx, id); err != nil {
		return errors.NewInternalServerError("failed to delete demand adjustment")
	}

	if err := uc.eventPublisher.PublishFADDeleted(ctx, id.String()); err != nil {
		return errors.NewInternalServerError("failed to publish FAD deleted event")
	}

	return nil
}
