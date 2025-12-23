package bufferProfile

import (
	"context"

	"github.com/melegattip/giia-core-engine/pkg/errors"
	"github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/catalog-service/internal/core/providers"
	"github.com/google/uuid"
)

type DeleteBufferProfileUseCase struct {
	bufferProfileRepo providers.BufferProfileRepository
	eventPublisher    providers.EventPublisher
	logger            logger.Logger
}

func NewDeleteBufferProfileUseCase(
	bufferProfileRepo providers.BufferProfileRepository,
	eventPublisher providers.EventPublisher,
	logger logger.Logger,
) *DeleteBufferProfileUseCase {
	return &DeleteBufferProfileUseCase{
		bufferProfileRepo: bufferProfileRepo,
		eventPublisher:    eventPublisher,
		logger:            logger,
	}
}

func (uc *DeleteBufferProfileUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return errors.NewBadRequest("buffer profile ID is required")
	}

	orgID, ok := ctx.Value("organization_id").(uuid.UUID)
	if !ok || orgID == uuid.Nil {
		return errors.NewBadRequest("organization ID is required in context")
	}

	bufferProfile, err := uc.bufferProfileRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := uc.bufferProfileRepo.Delete(ctx, id); err != nil {
		uc.logger.Error(ctx, err, "Failed to delete buffer profile", logger.Tags{
			"buffer_profile_id": id.String(),
		})
		return err
	}

	uc.logger.Info(ctx, "Buffer profile deleted successfully", logger.Tags{
		"buffer_profile_id": id.String(),
	})

	if err := uc.eventPublisher.PublishBufferProfileDeleted(ctx, bufferProfile); err != nil {
		uc.logger.Warn(ctx, "Failed to publish buffer profile deleted event", logger.Tags{
			"buffer_profile_id": id.String(),
			"error":             err.Error(),
		})
	}

	return nil
}
