package kpi

import (
	"context"
	"time"

	"github.com/melegattip/giia-core-engine/services/analytics-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/analytics-service/internal/core/providers"
	"github.com/google/uuid"
)

type SyncBufferAnalyticsUseCase struct {
	kpiRepo      providers.KPIRepository
	ddmrpClient  providers.DDMRPServiceClient
}

func NewSyncBufferAnalyticsUseCase(
	kpiRepo providers.KPIRepository,
	ddmrpClient providers.DDMRPServiceClient,
) *SyncBufferAnalyticsUseCase {
	return &SyncBufferAnalyticsUseCase{
		kpiRepo:     kpiRepo,
		ddmrpClient: ddmrpClient,
	}
}

type SyncBufferAnalyticsInput struct {
	OrganizationID uuid.UUID
	Date           time.Time
}

func (uc *SyncBufferAnalyticsUseCase) Execute(ctx context.Context, input *SyncBufferAnalyticsInput) (int, error) {
	if input == nil {
		return 0, domain.NewValidationError("input cannot be nil")
	}
	if input.OrganizationID == uuid.Nil {
		return 0, domain.NewValidationError("organization_id is required")
	}
	if input.Date.IsZero() {
		return 0, domain.NewValidationError("date is required")
	}

	bufferHistories, err := uc.ddmrpClient.ListBufferHistory(ctx, input.OrganizationID, input.Date, input.Date)
	if err != nil {
		return 0, err
	}

	synced := 0
	for _, history := range bufferHistories {
		analytics, err := domain.NewBufferAnalytics(
			history.ProductID,
			history.OrganizationID,
			history.Date,
			history.CPD,
			history.RedZone,
			history.RedBase,
			history.RedSafe,
			history.YellowZone,
			history.GreenZone,
			history.LTD,
			history.LeadTimeFactor,
			history.VariabilityFactor,
			history.MOQ,
			history.OrderFrequency,
			history.HasAdjustments,
		)
		if err != nil {
			continue
		}

		if err := uc.kpiRepo.SaveBufferAnalytics(ctx, analytics); err != nil {
			continue
		}

		synced++
	}

	return synced, nil
}
