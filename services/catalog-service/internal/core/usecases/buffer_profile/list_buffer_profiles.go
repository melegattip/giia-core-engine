package bufferProfile

import (
	"context"

	"github.com/melegattip/giia-core-engine/pkg/errors"
	"github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/catalog-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/catalog-service/internal/core/providers"
	"github.com/google/uuid"
)

type ListBufferProfilesRequest struct {
	Page     int `json:"page" validate:"gte=1"`
	PageSize int `json:"page_size" validate:"gte=1,lte=100"`
}

type ListBufferProfilesResponse struct {
	BufferProfiles []*domain.BufferProfile `json:"buffer_profiles"`
	Total          int64                   `json:"total"`
	Page           int                     `json:"page"`
	PageSize       int                     `json:"page_size"`
	TotalPages     int                     `json:"total_pages"`
}

type ListBufferProfilesUseCase struct {
	bufferProfileRepo providers.BufferProfileRepository
	logger            logger.Logger
}

func NewListBufferProfilesUseCase(
	bufferProfileRepo providers.BufferProfileRepository,
	logger logger.Logger,
) *ListBufferProfilesUseCase {
	return &ListBufferProfilesUseCase{
		bufferProfileRepo: bufferProfileRepo,
		logger:            logger,
	}
}

func (uc *ListBufferProfilesUseCase) Execute(ctx context.Context, req *ListBufferProfilesRequest) (*ListBufferProfilesResponse, error) {
	if req == nil {
		return nil, errors.NewBadRequest("request cannot be nil")
	}

	orgID, ok := ctx.Value("organization_id").(uuid.UUID)
	if !ok || orgID == uuid.Nil {
		return nil, errors.NewBadRequest("organization ID is required in context")
	}

	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	bufferProfiles, total, err := uc.bufferProfileRepo.List(ctx, make(map[string]interface{}), req.Page, req.PageSize)
	if err != nil {
		uc.logger.Error(ctx, err, "Failed to list buffer profiles", nil)
		return nil, err
	}

	totalPages := int(total) / req.PageSize
	if int(total)%req.PageSize != 0 {
		totalPages++
	}

	return &ListBufferProfilesResponse{
		BufferProfiles: bufferProfiles,
		Total:          total,
		Page:           req.Page,
		PageSize:       req.PageSize,
		TotalPages:     totalPages,
	}, nil
}
