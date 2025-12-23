package kpi

import (
	"context"
	"sort"
	"time"

	"github.com/melegattip/giia-core-engine/services/analytics-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/analytics-service/internal/core/providers"
	"github.com/google/uuid"
)

type CalculateInventoryRotationUseCase struct {
	kpiRepo         providers.KPIRepository
	executionClient providers.ExecutionServiceClient
}

func NewCalculateInventoryRotationUseCase(
	kpiRepo providers.KPIRepository,
	executionClient providers.ExecutionServiceClient,
) *CalculateInventoryRotationUseCase {
	return &CalculateInventoryRotationUseCase{
		kpiRepo:         kpiRepo,
		executionClient: executionClient,
	}
}

type CalculateInventoryRotationInput struct {
	OrganizationID uuid.UUID
	SnapshotDate   time.Time
}

func (uc *CalculateInventoryRotationUseCase) Execute(ctx context.Context, input *CalculateInventoryRotationInput) (*domain.InventoryRotationKPI, error) {
	if input == nil {
		return nil, domain.NewValidationError("input cannot be nil")
	}
	if input.OrganizationID == uuid.Nil {
		return nil, domain.NewValidationError("organization_id is required")
	}
	if input.SnapshotDate.IsZero() {
		return nil, domain.NewValidationError("snapshot_date is required")
	}

	startDate := input.SnapshotDate.AddDate(0, 0, -30)

	salesData, err := uc.executionClient.GetSalesData(ctx, input.OrganizationID, startDate, input.SnapshotDate)
	if err != nil {
		return nil, err
	}

	inventorySnapshots, err := uc.executionClient.GetInventorySnapshots(ctx, input.OrganizationID, startDate, input.SnapshotDate)
	if err != nil {
		return nil, err
	}

	totalStockValue := 0.0
	for _, snapshot := range inventorySnapshots {
		totalStockValue += snapshot.TotalValue
	}

	avgMonthlyStock := 0.0
	if len(inventorySnapshots) > 0 {
		avgMonthlyStock = totalStockValue / float64(len(inventorySnapshots))
	}

	productSales, err := uc.executionClient.GetProductSales(ctx, input.OrganizationID, startDate, input.SnapshotDate)
	if err != nil {
		return nil, err
	}

	topProducts := selectTopRotatingProducts(productSales, 10)
	slowProducts := selectSlowRotatingProducts(productSales, 10)

	kpi, err := domain.NewInventoryRotationKPI(
		input.OrganizationID,
		input.SnapshotDate,
		salesData.TotalValue,
		avgMonthlyStock,
		topProducts,
		slowProducts,
	)
	if err != nil {
		return nil, err
	}

	if err := uc.kpiRepo.SaveInventoryRotationKPI(ctx, kpi); err != nil {
		return nil, err
	}

	return kpi, nil
}

func selectTopRotatingProducts(productSales []*providers.ProductSales, limit int) []domain.RotatingProduct {
	products := make([]domain.RotatingProduct, 0, len(productSales))
	for _, ps := range productSales {
		products = append(products, domain.NewRotatingProduct(
			ps.ProductID,
			ps.SKU,
			ps.Name,
			ps.Sales30Days,
			ps.AvgStockValue,
		))
	}

	sort.Slice(products, func(i, j int) bool {
		return products[i].RotationRatio > products[j].RotationRatio
	})

	if len(products) > limit {
		products = products[:limit]
	}

	return products
}

func selectSlowRotatingProducts(productSales []*providers.ProductSales, limit int) []domain.RotatingProduct {
	products := make([]domain.RotatingProduct, 0, len(productSales))
	for _, ps := range productSales {
		products = append(products, domain.NewRotatingProduct(
			ps.ProductID,
			ps.SKU,
			ps.Name,
			ps.Sales30Days,
			ps.AvgStockValue,
		))
	}

	sort.Slice(products, func(i, j int) bool {
		return products[i].RotationRatio < products[j].RotationRatio
	})

	if len(products) > limit {
		products = products[:limit]
	}

	return products
}
