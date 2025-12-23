package alert

import (
	"context"

	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/providers"
	"github.com/google/uuid"
)

type GeneratePODelayAlertsUseCase struct {
	poRepo         providers.PurchaseOrderRepository
	alertRepo      providers.AlertRepository
	eventPublisher providers.EventPublisher
}

func NewGeneratePODelayAlertsUseCase(
	poRepo providers.PurchaseOrderRepository,
	alertRepo providers.AlertRepository,
	publisher providers.EventPublisher,
) *GeneratePODelayAlertsUseCase {
	return &GeneratePODelayAlertsUseCase{
		poRepo:         poRepo,
		alertRepo:      alertRepo,
		eventPublisher: publisher,
	}
}

func (uc *GeneratePODelayAlertsUseCase) Execute(ctx context.Context, organizationID uuid.UUID) error {
	if organizationID == uuid.Nil {
		return domain.NewValidationError("organization_id is required")
	}

	delayedPOs, err := uc.poRepo.GetDelayedOrders(ctx, organizationID)
	if err != nil {
		return err
	}

	for _, po := range delayedPOs {
		po.CheckDelay()

		if !po.IsDelayed {
			continue
		}

		existingAlerts, _ := uc.alertRepo.GetByResourceID(ctx, "purchase_order", po.ID, organizationID)
		hasActiveAlert := false
		for _, alert := range existingAlerts {
			if alert.AlertType == domain.AlertTypePODelayed && alert.IsActive() {
				hasActiveAlert = true
				break
			}
		}

		if hasActiveAlert {
			continue
		}

		alert := domain.NewPODelayedAlert(po)
		if err := uc.alertRepo.Create(ctx, alert); err != nil {
			return err
		}

		uc.eventPublisher.PublishAlertCreated(ctx, alert)
	}

	return nil
}