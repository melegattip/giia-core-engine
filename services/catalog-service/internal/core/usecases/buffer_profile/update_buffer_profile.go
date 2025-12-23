package bufferProfile

import (
	"context"

	"github.com/giia/giia-core-engine/pkg/errors"
	"github.com/giia/giia-core-engine/pkg/logger"
	"github.com/giia/giia-core-engine/services/catalog-service/internal/core/domain"
	"github.com/giia/giia-core-engine/services/catalog-service/internal/core/providers"
	"github.com/google/uuid"
)

type UpdateBufferProfileRequest struct {
	ID                 uuid.UUID `json:"id" validate:"required"`
	Name               string    `json:"name" validate:"required,max=100"`
	Description        string    `json:"description"`
	LeadTimeFactor     float64   `json:"lead_time_factor" validate:"required,gt=0"`
	VariabilityFactor  float64   `json:"variability_factor" validate:"required,gt=0"`
	TargetServiceLevel int       `json:"target_service_level" validate:"gte=0,lte=100"`
}

type UpdateBufferProfileUseCase struct {
	bufferProfileRepo providers.BufferProfileRepository
	eventPublisher    providers.EventPublisher
	logger            logger.Logger
}

func NewUpdateBufferProfileUseCase(
	bufferProfileRepo providers.BufferProfileRepository,
	eventPublisher providers.EventPublisher,
	logger logger.Logger,
) *UpdateBufferProfileUseCase {
	return &UpdateBufferProfileUseCase{
		bufferProfileRepo: bufferProfileRepo,
		eventPublisher:    eventPublisher,
		logger:            logger,
	}
}

func (uc *UpdateBufferProfileUseCase) Execute(ctx context.Context, req *UpdateBufferProfileRequest) (*domain.BufferProfile, error) {
	if req == nil {
		return nil, errors.NewBadRequest("request cannot be nil")
	}

	if req.ID == uuid.Nil {
		return nil, errors.NewBadRequest("buffer profile ID is required")
	}

	orgID, ok := ctx.Value("organization_id").(uuid.UUID)
	if !ok || orgID == uuid.Nil {
		return nil, errors.NewBadRequest("organization ID is required in context")
	}

	existingProfile, err := uc.bufferProfileRepo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	existingProfile.Name = req.Name
	existingProfile.Description = req.Description
	existingProfile.LeadTimeFactor = req.LeadTimeFactor
	existingProfile.VariabilityFactor = req.VariabilityFactor
	existingProfile.TargetServiceLevel = req.TargetServiceLevel

	if err := existingProfile.Validate(); err != nil {
		uc.logger.Warn(ctx, "Buffer profile validation failed", logger.Tags{
			"buffer_profile_id": req.ID.String(),
			"error":             err.Error(),
		})
		return nil, err
	}

	if err := uc.bufferProfileRepo.Update(ctx, existingProfile); err != nil {
		uc.logger.Error(ctx, err, "Failed to update buffer profile", logger.Tags{
			"buffer_profile_id": req.ID.String(),
		})
		return nil, err
	}

	uc.logger.Info(ctx, "Buffer profile updated successfully", logger.Tags{
		"buffer_profile_id": existingProfile.ID.String(),
	})

	if err := uc.eventPublisher.PublishBufferProfileUpdated(ctx, existingProfile); err != nil {
		uc.logger.Warn(ctx, "Failed to publish buffer profile updated event", logger.Tags{
			"buffer_profile_id": existingProfile.ID.String(),
			"error":             err.Error(),
		})
	}

	return existingProfile, nil
}
