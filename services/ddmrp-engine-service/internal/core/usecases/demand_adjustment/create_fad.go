package demand_adjustment

import (
	"context"
	"time"

	"github.com/giia/giia-core-engine/pkg/errors"
	"github.com/giia/giia-core-engine/services/ddmrp-engine-service/internal/core/domain"
	"github.com/giia/giia-core-engine/services/ddmrp-engine-service/internal/core/providers"
	"github.com/google/uuid"
)

type CreateFADUseCase struct {
	demandAdjustmentRepo providers.DemandAdjustmentRepository
	eventPublisher       providers.EventPublisher
}

func NewCreateFADUseCase(
	repo providers.DemandAdjustmentRepository,
	publisher providers.EventPublisher,
) *CreateFADUseCase {
	return &CreateFADUseCase{
		demandAdjustmentRepo: repo,
		eventPublisher:       publisher,
	}
}

type CreateFADInput struct {
	ProductID      uuid.UUID
	OrganizationID uuid.UUID
	StartDate      time.Time
	EndDate        time.Time
	AdjustmentType domain.DemandAdjustmentType
	Factor         float64
	Reason         string
	CreatedBy      uuid.UUID
}

func (uc *CreateFADUseCase) Execute(ctx context.Context, input CreateFADInput) (*domain.DemandAdjustment, error) {
	fad, err := domain.NewDemandAdjustment(
		input.ProductID,
		input.OrganizationID,
		input.CreatedBy,
		input.StartDate,
		input.EndDate,
		input.AdjustmentType,
		input.Factor,
		input.Reason,
	)
	if err != nil {
		return nil, err
	}

	if err := uc.demandAdjustmentRepo.Create(ctx, fad); err != nil {
		return nil, errors.NewInternalServerError("failed to create demand adjustment")
	}

	if err := uc.eventPublisher.PublishFADCreated(ctx, fad); err != nil {
		return nil, errors.NewInternalServerError("failed to publish FAD created event")
	}

	return fad, nil
}
