package kpi

import (
	"context"
	"time"

	"github.com/melegattip/giia-core-engine/services/analytics-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/analytics-service/internal/core/providers"
	"github.com/google/uuid"
)

type CalculateImmobilizedInventoryUseCase struct {
	kpiRepo       providers.KPIRepository
	catalogClient providers.CatalogServiceClient
}

func NewCalculateImmobilizedInventoryUseCase(
	kpiRepo providers.KPIRepository,
	catalogClient providers.CatalogServiceClient,
) *CalculateImmobilizedInventoryUseCase {
	return &CalculateImmobilizedInventoryUseCase{
		kpiRepo:       kpiRepo,
		catalogClient: catalogClient,
	}
}

type CalculateImmobilizedInventoryInput struct {
	OrganizationID uuid.UUID
	SnapshotDate   time.Time
	ThresholdYears int
}

func (uc *CalculateImmobilizedInventoryUseCase) Execute(ctx context.Context, input *CalculateImmobilizedInventoryInput) (*domain.ImmobilizedInventoryKPI, error) {
	if input == nil {
		return nil, domain.NewValidationError("input cannot be nil")
	}
	if input.OrganizationID == uuid.Nil {
		return nil, domain.NewValidationError("organization_id is required")
	}
	if input.SnapshotDate.IsZero() {
		return nil, domain.NewValidationError("snapshot_date is required")
	}
	if input.ThresholdYears <= 0 {
		return nil, domain.NewValidationError("threshold_years must be positive")
	}

	products, err := uc.catalogClient.ListProductsWithInventory(ctx, input.OrganizationID)
	if err != nil {
		return nil, err
	}

	thresholdDate := input.SnapshotDate.AddDate(-input.ThresholdYears, 0, 0)
	immobilizedCount := 0
	immobilizedValue := 0.0
	totalStockValue := 0.0

	for _, product := range products {
		productValue := product.Quantity * product.StandardCost
		totalStockValue += productValue

		if product.LastPurchaseDate != nil && product.LastPurchaseDate.Before(thresholdDate) {
			immobilizedCount++
			immobilizedValue += productValue
		}
	}

	kpi, err := domain.NewImmobilizedInventoryKPI(
		input.OrganizationID,
		input.SnapshotDate,
		input.ThresholdYears,
		immobilizedCount,
		immobilizedValue,
		totalStockValue,
	)
	if err != nil {
		return nil, err
	}

	if err := uc.kpiRepo.SaveImmobilizedInventoryKPI(ctx, kpi); err != nil {
		return nil, err
	}

	return kpi, nil
}
