package kpi

import (
	"context"
	"time"

	"github.com/giia/giia-core-engine/services/analytics-service/internal/core/domain"
	"github.com/giia/giia-core-engine/services/analytics-service/internal/core/providers"
	"github.com/google/uuid"
)

type CalculateDaysInInventoryUseCase struct {
	kpiRepo       providers.KPIRepository
	catalogClient providers.CatalogServiceClient
}

func NewCalculateDaysInInventoryUseCase(
	kpiRepo providers.KPIRepository,
	catalogClient providers.CatalogServiceClient,
) *CalculateDaysInInventoryUseCase {
	return &CalculateDaysInInventoryUseCase{
		kpiRepo:       kpiRepo,
		catalogClient: catalogClient,
	}
}

type CalculateDaysInInventoryInput struct {
	OrganizationID uuid.UUID
	SnapshotDate   time.Time
}

func (uc *CalculateDaysInInventoryUseCase) Execute(ctx context.Context, input *CalculateDaysInInventoryInput) (*domain.DaysInInventoryKPI, error) {
	if input == nil {
		return nil, domain.NewValidationError("input cannot be nil")
	}
	if input.OrganizationID == uuid.Nil {
		return nil, domain.NewValidationError("organization_id is required")
	}
	if input.SnapshotDate.IsZero() {
		return nil, domain.NewValidationError("snapshot_date is required")
	}

	products, err := uc.catalogClient.ListProductsWithInventory(ctx, input.OrganizationID)
	if err != nil {
		return nil, err
	}

	totalValuedDays := 0.0
	totalProducts := 0

	for _, product := range products {
		if product.LastPurchaseDate == nil {
			continue
		}

		daysInInventory := int(input.SnapshotDate.Sub(*product.LastPurchaseDate).Hours() / 24)
		if daysInInventory < 0 {
			continue
		}

		totalValue := product.Quantity * product.StandardCost
		valuedDays := float64(daysInInventory) * totalValue

		totalValuedDays += valuedDays
		totalProducts++
	}

	averageValuedDays := 0.0
	if totalProducts > 0 {
		averageValuedDays = totalValuedDays / float64(totalProducts)
	}

	kpi, err := domain.NewDaysInInventoryKPI(
		input.OrganizationID,
		input.SnapshotDate,
		totalValuedDays,
		averageValuedDays,
		totalProducts,
	)
	if err != nil {
		return nil, err
	}

	if err := uc.kpiRepo.SaveDaysInInventoryKPI(ctx, kpi); err != nil {
		return nil, err
	}

	return kpi, nil
}
