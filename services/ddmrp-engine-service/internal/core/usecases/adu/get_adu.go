package adu

import (
	"context"

	"github.com/melegattip/giia-core-engine/pkg/errors"
	"github.com/melegattip/giia-core-engine/services/ddmrp-engine-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/ddmrp-engine-service/internal/core/providers"
	"github.com/google/uuid"
)

type GetADUUseCase struct {
	aduRepo providers.ADURepository
}

func NewGetADUUseCase(aduRepo providers.ADURepository) *GetADUUseCase {
	return &GetADUUseCase{
		aduRepo: aduRepo,
	}
}

func (uc *GetADUUseCase) Execute(ctx context.Context, productID, organizationID uuid.UUID) (*domain.ADUCalculation, error) {
	if productID == uuid.Nil {
		return nil, errors.NewBadRequest("product_id is required")
	}
	if organizationID == uuid.Nil {
		return nil, errors.NewBadRequest("organization_id is required")
	}

	adu, err := uc.aduRepo.GetLatest(ctx, productID, organizationID)
	if err != nil {
		return nil, errors.NewNotFound("ADU calculation not found")
	}

	return adu, nil
}
