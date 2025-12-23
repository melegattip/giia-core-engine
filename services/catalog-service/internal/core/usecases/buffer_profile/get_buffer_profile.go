package bufferProfile

import (
	"context"

	"github.com/melegattip/giia-core-engine/pkg/errors"
	"github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/catalog-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/catalog-service/internal/core/providers"
	"github.com/google/uuid"
)

type GetBufferProfileUseCase struct {
	bufferProfileRepo providers.BufferProfileRepository
	logger            logger.Logger
}

func NewGetBufferProfileUseCase(
	bufferProfileRepo providers.BufferProfileRepository,
	logger logger.Logger,
) *GetBufferProfileUseCase {
	return &GetBufferProfileUseCase{
		bufferProfileRepo: bufferProfileRepo,
		logger:            logger,
	}
}

func (uc *GetBufferProfileUseCase) Execute(ctx context.Context, id uuid.UUID) (*domain.BufferProfile, error) {
	if id == uuid.Nil {
		return nil, errors.NewBadRequest("buffer profile ID is required")
	}

	orgID, ok := ctx.Value("organization_id").(uuid.UUID)
	if !ok || orgID == uuid.Nil {
		return nil, errors.NewBadRequest("organization ID is required in context")
	}

	bufferProfile, err := uc.bufferProfileRepo.GetByID(ctx, id)
	if err != nil {
		uc.logger.Warn(ctx, "Buffer profile not found", logger.Tags{
			"buffer_profile_id": id.String(),
		})
		return nil, err
	}

	return bufferProfile, nil
}
