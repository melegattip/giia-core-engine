package grpc

import (
	"context"

	"github.com/giia/giia-core-engine/services/ddmrp-engine-service/internal/core/domain"
	"github.com/giia/giia-core-engine/services/ddmrp-engine-service/internal/core/usecases/buffer"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type BufferHandler struct {
	calculateBufferUC *buffer.CalculateBufferUseCase
	getBufferUC       *buffer.GetBufferUseCase
	listBuffersUC     *buffer.ListBuffersUseCase
}

func NewBufferHandler(
	calculateUC *buffer.CalculateBufferUseCase,
	getUC *buffer.GetBufferUseCase,
	listUC *buffer.ListBuffersUseCase,
) *BufferHandler {
	return &BufferHandler{
		calculateBufferUC: calculateUC,
		getBufferUC:       getUC,
		listBuffersUC:     listUC,
	}
}

func (h *BufferHandler) CalculateBuffer(ctx context.Context, productID, organizationID uuid.UUID) (*domain.Buffer, error) {
	return h.calculateBufferUC.Execute(ctx, buffer.CalculateBufferInput{
		ProductID:      productID,
		OrganizationID: organizationID,
	})
}

func (h *BufferHandler) GetBuffer(ctx context.Context, productID, organizationID uuid.UUID) (*domain.Buffer, error) {
	return h.getBufferUC.Execute(ctx, productID, organizationID)
}

func (h *BufferHandler) ListBuffers(ctx context.Context, input buffer.ListBuffersInput) ([]domain.Buffer, error) {
	return h.listBuffersUC.Execute(ctx, input)
}

func BufferToProto(b *domain.Buffer) map[string]interface{} {
	return map[string]interface{}{
		"id":                   b.ID.String(),
		"product_id":           b.ProductID.String(),
		"organization_id":      b.OrganizationID.String(),
		"buffer_profile_id":    b.BufferProfileID.String(),
		"cpd":                  b.CPD,
		"ltd":                  int32(b.LTD),
		"red_base":             b.RedBase,
		"red_safe":             b.RedSafe,
		"red_zone":             b.RedZone,
		"yellow_zone":          b.YellowZone,
		"green_zone":           b.GreenZone,
		"top_of_red":           b.TopOfRed,
		"top_of_yellow":        b.TopOfYellow,
		"top_of_green":         b.TopOfGreen,
		"on_hand":              b.OnHand,
		"on_order":             b.OnOrder,
		"qualified_demand":     b.QualifiedDemand,
		"net_flow_position":    b.NetFlowPosition,
		"buffer_penetration":   b.BufferPenetration,
		"zone":                 string(b.Zone),
		"alert_level":          string(b.AlertLevel),
		"last_recalculated_at": timestamppb.New(b.LastRecalculatedAt),
		"created_at":           timestamppb.New(b.CreatedAt),
		"updated_at":           timestamppb.New(b.UpdatedAt),
	}
}
