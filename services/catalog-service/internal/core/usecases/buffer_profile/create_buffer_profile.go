package bufferProfile

import (
	"context"

	"github.com/melegattip/giia-core-engine/pkg/errors"
	"github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/catalog-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/catalog-service/internal/core/providers"
	"github.com/google/uuid"
)

type CreateBufferProfileRequest struct {
	Name               string  `json:"name" validate:"required,max=100"`
	Description        string  `json:"description"`
	LeadTimeFactor     float64 `json:"lead_time_factor" validate:"required,gt=0"`
	VariabilityFactor  float64 `json:"variability_factor" validate:"required,gt=0"`
	TargetServiceLevel int     `json:"target_service_level" validate:"gte=0,lte=100"`
}

type CreateBufferProfileUseCase struct {
	bufferProfileRepo providers.BufferProfileRepository
	eventPublisher    providers.EventPublisher
	logger            logger.Logger
}

func NewCreateBufferProfileUseCase(
	bufferProfileRepo providers.BufferProfileRepository,
	eventPublisher providers.EventPublisher,
	logger logger.Logger,
) *CreateBufferProfileUseCase {
	return &CreateBufferProfileUseCase{
		bufferProfileRepo: bufferProfileRepo,
		eventPublisher:    eventPublisher,
		logger:            logger,
	}
}

func (uc *CreateBufferProfileUseCase) Execute(ctx context.Context, req *CreateBufferProfileRequest) (*domain.BufferProfile, error) {
	if req == nil {
		return nil, errors.NewBadRequest("request cannot be nil")
	}

	orgID, ok := ctx.Value("organization_id").(uuid.UUID)
	if !ok || orgID == uuid.Nil {
		return nil, errors.NewBadRequest("organization ID is required in context")
	}

	if req.TargetServiceLevel == 0 {
		req.TargetServiceLevel = 95
	}

	bufferProfile := &domain.BufferProfile{
		Name:               req.Name,
		Description:        req.Description,
		LeadTimeFactor:     req.LeadTimeFactor,
		VariabilityFactor:  req.VariabilityFactor,
		TargetServiceLevel: req.TargetServiceLevel,
		OrganizationID:     orgID,
	}

	if err := bufferProfile.Validate(); err != nil {
		uc.logger.Warn(ctx, "Buffer profile validation failed", logger.Tags{
			"name":  req.Name,
			"error": err.Error(),
		})
		return nil, err
	}

	if err := uc.bufferProfileRepo.Create(ctx, bufferProfile); err != nil {
		uc.logger.Error(ctx, err, "Failed to create buffer profile", logger.Tags{
			"name": req.Name,
		})
		return nil, err
	}

	uc.logger.Info(ctx, "Buffer profile created successfully", logger.Tags{
		"buffer_profile_id": bufferProfile.ID.String(),
		"name":              bufferProfile.Name,
	})

	if err := uc.eventPublisher.PublishBufferProfileCreated(ctx, bufferProfile); err != nil {
		uc.logger.Warn(ctx, "Failed to publish buffer profile created event", logger.Tags{
			"buffer_profile_id": bufferProfile.ID.String(),
			"error":             err.Error(),
		})
	}

	return bufferProfile, nil
}
