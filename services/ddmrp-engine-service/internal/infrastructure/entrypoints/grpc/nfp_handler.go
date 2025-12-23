package grpc

import (
	"context"

	"github.com/melegattip/giia-core-engine/services/ddmrp-engine-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/ddmrp-engine-service/internal/core/usecases/nfp"
	"github.com/google/uuid"
)

type NFPHandler struct {
	updateNFPUC          *nfp.UpdateNFPUseCase
	checkReplenishmentUC *nfp.CheckReplenishmentUseCase
}

func NewNFPHandler(
	updateUC *nfp.UpdateNFPUseCase,
	checkUC *nfp.CheckReplenishmentUseCase,
) *NFPHandler {
	return &NFPHandler{
		updateNFPUC:          updateUC,
		checkReplenishmentUC: checkUC,
	}
}

func (h *NFPHandler) UpdateNFP(ctx context.Context, productID, organizationID uuid.UUID, onHand, onOrder, qualifiedDemand float64) (*domain.Buffer, error) {
	return h.updateNFPUC.Execute(ctx, nfp.UpdateNFPInput{
		ProductID:       productID,
		OrganizationID:  organizationID,
		OnHand:          onHand,
		OnOrder:         onOrder,
		QualifiedDemand: qualifiedDemand,
	})
}

func (h *NFPHandler) CheckReplenishment(ctx context.Context, organizationID uuid.UUID) ([]domain.Buffer, error) {
	return h.checkReplenishmentUC.Execute(ctx, organizationID)
}
