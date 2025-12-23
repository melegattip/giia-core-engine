package adu

import (
	"context"

	"github.com/giia/giia-core-engine/pkg/errors"
	"github.com/giia/giia-core-engine/services/ddmrp-engine-service/internal/core/domain"
	"github.com/giia/giia-core-engine/services/ddmrp-engine-service/internal/core/providers"
	"github.com/google/uuid"
)

type ListADUHistoryUseCase struct {
	aduRepo providers.ADURepository
}

func NewListADUHistoryUseCase(aduRepo providers.ADURepository) *ListADUHistoryUseCase {
	return &ListADUHistoryUseCase{
		aduRepo: aduRepo,
	}
}

type ListADUHistoryInput struct {
	ProductID      uuid.UUID
	OrganizationID uuid.UUID
	Limit          int
}

func (uc *ListADUHistoryUseCase) Execute(ctx context.Context, input ListADUHistoryInput) ([]domain.ADUCalculation, error) {
	if input.ProductID == uuid.Nil {
		return nil, errors.NewBadRequest("product_id is required")
	}
	if input.OrganizationID == uuid.Nil {
		return nil, errors.NewBadRequest("organization_id is required")
	}

	if input.Limit <= 0 {
		input.Limit = 30
	}

	adus, err := uc.aduRepo.ListHistory(ctx, input.ProductID, input.OrganizationID, input.Limit)
	if err != nil {
		return nil, errors.NewInternalServerError("failed to list ADU history")
	}

	return adus, nil
}
