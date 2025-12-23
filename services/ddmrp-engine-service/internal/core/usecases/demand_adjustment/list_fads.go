package demand_adjustment

import (
	"context"

	"github.com/giia/giia-core-engine/pkg/errors"
	"github.com/giia/giia-core-engine/services/ddmrp-engine-service/internal/core/domain"
	"github.com/giia/giia-core-engine/services/ddmrp-engine-service/internal/core/providers"
	"github.com/google/uuid"
)

type ListFADsUseCase struct {
	demandAdjustmentRepo providers.DemandAdjustmentRepository
}

func NewListFADsUseCase(repo providers.DemandAdjustmentRepository) *ListFADsUseCase {
	return &ListFADsUseCase{
		demandAdjustmentRepo: repo,
	}
}

type ListFADsByProductInput struct {
	ProductID      uuid.UUID
	OrganizationID uuid.UUID
}

func (uc *ListFADsUseCase) ExecuteByProduct(ctx context.Context, input ListFADsByProductInput) ([]domain.DemandAdjustment, error) {
	if input.ProductID == uuid.Nil {
		return nil, errors.NewBadRequest("product_id is required")
	}
	if input.OrganizationID == uuid.Nil {
		return nil, errors.NewBadRequest("organization_id is required")
	}

	fads, err := uc.demandAdjustmentRepo.ListByProduct(ctx, input.ProductID, input.OrganizationID)
	if err != nil {
		return nil, errors.NewInternalServerError("failed to list demand adjustments")
	}

	return fads, nil
}

type ListFADsByOrganizationInput struct {
	OrganizationID uuid.UUID
	Limit          int
	Offset         int
}

func (uc *ListFADsUseCase) ExecuteByOrganization(ctx context.Context, input ListFADsByOrganizationInput) ([]domain.DemandAdjustment, error) {
	if input.OrganizationID == uuid.Nil {
		return nil, errors.NewBadRequest("organization_id is required")
	}

	fads, err := uc.demandAdjustmentRepo.ListByOrganization(ctx, input.OrganizationID, input.Limit, input.Offset)
	if err != nil {
		return nil, errors.NewInternalServerError("failed to list demand adjustments")
	}

	return fads, nil
}
