package buffer

import (
	"context"

	"github.com/melegattip/giia-core-engine/pkg/errors"
	"github.com/melegattip/giia-core-engine/services/ddmrp-engine-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/ddmrp-engine-service/internal/core/providers"
	"github.com/google/uuid"
)

type GetBufferUseCase struct {
	bufferRepo providers.BufferRepository
}

func NewGetBufferUseCase(bufferRepo providers.BufferRepository) *GetBufferUseCase {
	return &GetBufferUseCase{
		bufferRepo: bufferRepo,
	}
}

func (uc *GetBufferUseCase) Execute(ctx context.Context, productID, organizationID uuid.UUID) (*domain.Buffer, error) {
	if productID == uuid.Nil {
		return nil, errors.NewBadRequest("product_id is required")
	}
	if organizationID == uuid.Nil {
		return nil, errors.NewBadRequest("organization_id is required")
	}

	buffer, err := uc.bufferRepo.GetByProduct(ctx, productID, organizationID)
	if err != nil {
		return nil, errors.NewNotFound("buffer not found")
	}

	return buffer, nil
}
