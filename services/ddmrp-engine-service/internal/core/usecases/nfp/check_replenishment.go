package nfp

import (
	"context"

	"github.com/giia/giia-core-engine/pkg/errors"
	"github.com/giia/giia-core-engine/services/ddmrp-engine-service/internal/core/domain"
	"github.com/giia/giia-core-engine/services/ddmrp-engine-service/internal/core/providers"
	"github.com/google/uuid"
)

type CheckReplenishmentUseCase struct {
	bufferRepo providers.BufferRepository
}

func NewCheckReplenishmentUseCase(bufferRepo providers.BufferRepository) *CheckReplenishmentUseCase {
	return &CheckReplenishmentUseCase{
		bufferRepo: bufferRepo,
	}
}

func (uc *CheckReplenishmentUseCase) Execute(ctx context.Context, organizationID uuid.UUID) ([]domain.Buffer, error) {
	if organizationID == uuid.Nil {
		return nil, errors.NewBadRequest("organization_id is required")
	}

	replenishBuffers, err := uc.bufferRepo.ListByAlertLevel(ctx, organizationID, domain.AlertReplenish)
	if err != nil {
		return nil, errors.NewInternalServerError("failed to list replenish buffers")
	}

	criticalBuffers, err := uc.bufferRepo.ListByAlertLevel(ctx, organizationID, domain.AlertCritical)
	if err != nil {
		return nil, errors.NewInternalServerError("failed to list critical buffers")
	}

	result := append(criticalBuffers, replenishBuffers...)

	return result, nil
}
