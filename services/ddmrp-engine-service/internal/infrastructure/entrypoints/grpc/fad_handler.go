package grpc

import (
	"context"
	"time"

	"github.com/melegattip/giia-core-engine/services/ddmrp-engine-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/ddmrp-engine-service/internal/core/usecases/demand_adjustment"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type FADHandler struct {
	createFADUC *demand_adjustment.CreateFADUseCase
	updateFADUC *demand_adjustment.UpdateFADUseCase
	deleteFADUC *demand_adjustment.DeleteFADUseCase
	listFADsUC  *demand_adjustment.ListFADsUseCase
}

func NewFADHandler(
	createUC *demand_adjustment.CreateFADUseCase,
	updateUC *demand_adjustment.UpdateFADUseCase,
	deleteUC *demand_adjustment.DeleteFADUseCase,
	listUC *demand_adjustment.ListFADsUseCase,
) *FADHandler {
	return &FADHandler{
		createFADUC: createUC,
		updateFADUC: updateUC,
		deleteFADUC: deleteUC,
		listFADsUC:  listUC,
	}
}

func (h *FADHandler) CreateFAD(ctx context.Context, productID, organizationID, createdBy uuid.UUID, startDate, endDate time.Time, adjustmentType domain.DemandAdjustmentType, factor float64, reason string) (*domain.DemandAdjustment, error) {
	return h.createFADUC.Execute(ctx, demand_adjustment.CreateFADInput{
		ProductID:      productID,
		OrganizationID: organizationID,
		StartDate:      startDate,
		EndDate:        endDate,
		AdjustmentType: adjustmentType,
		Factor:         factor,
		Reason:         reason,
		CreatedBy:      createdBy,
	})
}

func (h *FADHandler) UpdateFAD(ctx context.Context, id uuid.UUID, startDate, endDate time.Time, adjustmentType domain.DemandAdjustmentType, factor float64, reason string) (*domain.DemandAdjustment, error) {
	return h.updateFADUC.Execute(ctx, demand_adjustment.UpdateFADInput{
		ID:             id,
		StartDate:      startDate,
		EndDate:        endDate,
		AdjustmentType: adjustmentType,
		Factor:         factor,
		Reason:         reason,
	})
}

func (h *FADHandler) DeleteFAD(ctx context.Context, id uuid.UUID) error {
	return h.deleteFADUC.Execute(ctx, id)
}

func (h *FADHandler) ListFADsByProduct(ctx context.Context, productID, organizationID uuid.UUID) ([]domain.DemandAdjustment, error) {
	return h.listFADsUC.ExecuteByProduct(ctx, demand_adjustment.ListFADsByProductInput{
		ProductID:      productID,
		OrganizationID: organizationID,
	})
}

func (h *FADHandler) ListFADsByOrganization(ctx context.Context, organizationID uuid.UUID, limit, offset int) ([]domain.DemandAdjustment, error) {
	return h.listFADsUC.ExecuteByOrganization(ctx, demand_adjustment.ListFADsByOrganizationInput{
		OrganizationID: organizationID,
		Limit:          limit,
		Offset:         offset,
	})
}

func FADToProto(fad *domain.DemandAdjustment) map[string]interface{} {
	return map[string]interface{}{
		"id":              fad.ID.String(),
		"product_id":      fad.ProductID.String(),
		"organization_id": fad.OrganizationID.String(),
		"start_date":      timestamppb.New(fad.StartDate),
		"end_date":        timestamppb.New(fad.EndDate),
		"adjustment_type": string(fad.AdjustmentType),
		"factor":          fad.Factor,
		"reason":          fad.Reason,
		"created_at":      timestamppb.New(fad.CreatedAt),
		"created_by":      fad.CreatedBy.String(),
	}
}
