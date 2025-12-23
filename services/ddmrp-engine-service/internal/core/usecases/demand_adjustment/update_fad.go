package demand_adjustment

import (
	"context"
	"time"

	"github.com/giia/giia-core-engine/pkg/errors"
	"github.com/giia/giia-core-engine/services/ddmrp-engine-service/internal/core/domain"
	"github.com/giia/giia-core-engine/services/ddmrp-engine-service/internal/core/providers"
	"github.com/google/uuid"
)

type UpdateFADUseCase struct {
	demandAdjustmentRepo providers.DemandAdjustmentRepository
	eventPublisher       providers.EventPublisher
}

func NewUpdateFADUseCase(
	repo providers.DemandAdjustmentRepository,
	publisher providers.EventPublisher,
) *UpdateFADUseCase {
	return &UpdateFADUseCase{
		demandAdjustmentRepo: repo,
		eventPublisher:       publisher,
	}
}

type UpdateFADInput struct {
	ID             uuid.UUID
	StartDate      time.Time
	EndDate        time.Time
	AdjustmentType domain.DemandAdjustmentType
	Factor         float64
	Reason         string
}

func (uc *UpdateFADUseCase) Execute(ctx context.Context, input UpdateFADInput) (*domain.DemandAdjustment, error) {
	if input.ID == uuid.Nil {
		return nil, errors.NewBadRequest("id is required")
	}

	fad, err := uc.demandAdjustmentRepo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, errors.NewNotFound("demand adjustment not found")
	}

	fad.StartDate = input.StartDate
	fad.EndDate = input.EndDate
	fad.AdjustmentType = input.AdjustmentType
	fad.Factor = input.Factor
	fad.Reason = input.Reason

	if err := fad.Validate(); err != nil {
		return nil, err
	}

	if err := uc.demandAdjustmentRepo.Update(ctx, fad); err != nil {
		return nil, errors.NewInternalServerError("failed to update demand adjustment")
	}

	if err := uc.eventPublisher.PublishFADUpdated(ctx, fad); err != nil {
		return nil, errors.NewInternalServerError("failed to publish FAD updated event")
	}

	return fad, nil
}
