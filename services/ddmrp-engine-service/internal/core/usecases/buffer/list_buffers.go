package buffer

import (
	"context"

	"github.com/melegattip/giia-core-engine/pkg/errors"
	"github.com/melegattip/giia-core-engine/services/ddmrp-engine-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/ddmrp-engine-service/internal/core/providers"
	"github.com/google/uuid"
)

type ListBuffersUseCase struct {
	bufferRepo providers.BufferRepository
}

func NewListBuffersUseCase(bufferRepo providers.BufferRepository) *ListBuffersUseCase {
	return &ListBuffersUseCase{
		bufferRepo: bufferRepo,
	}
}

type ListBuffersInput struct {
	OrganizationID uuid.UUID
	Zone           domain.ZoneType
	AlertLevel     domain.AlertLevel
	Limit          int
	Offset         int
}

func (uc *ListBuffersUseCase) Execute(ctx context.Context, input ListBuffersInput) ([]domain.Buffer, error) {
	if input.OrganizationID == uuid.Nil {
		return nil, errors.NewBadRequest("organization_id is required")
	}

	var buffers []domain.Buffer
	var err error

	if input.Zone != "" {
		buffers, err = uc.bufferRepo.ListByZone(ctx, input.OrganizationID, input.Zone)
	} else if input.AlertLevel != "" {
		buffers, err = uc.bufferRepo.ListByAlertLevel(ctx, input.OrganizationID, input.AlertLevel)
	} else {
		buffers, err = uc.bufferRepo.List(ctx, input.OrganizationID, input.Limit, input.Offset)
	}

	if err != nil {
		return nil, errors.NewInternalServerError("failed to list buffers")
	}

	return buffers, nil
}
