package kpi_test

import (
	"context"
	"testing"
	"time"

	"github.com/melegattip/giia-core-engine/services/analytics-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/analytics-service/internal/core/providers"
	"github.com/melegattip/giia-core-engine/services/analytics-service/internal/core/usecases/kpi"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCalculateImmobilizedInventoryUseCase_Execute_WithValidData_CalculatesKPI(t *testing.T) {
	givenOrgID := uuid.New()
	givenDate := time.Now()
	purchaseDate := givenDate.AddDate(-3, 0, 0)

	givenProducts := []*providers.ProductWithInventory{
		{
			ProductID:        uuid.New(),
			SKU:              "P001",
			Name:             "Old Product",
			Quantity:         100,
			StandardCost:     50,
			LastPurchaseDate: &purchaseDate,
		},
		{
			ProductID:        uuid.New(),
			SKU:              "P002",
			Name:             "Recent Product",
			Quantity:         200,
			StandardCost:     75,
			LastPurchaseDate: &givenDate,
		},
	}

	mockKPIRepo := new(MockKPIRepository)
	mockCatalogClient := new(MockCatalogServiceClient)

	mockCatalogClient.On("ListProductsWithInventory", mock.Anything, givenOrgID).Return(givenProducts, nil)
	mockKPIRepo.On("SaveImmobilizedInventoryKPI", mock.Anything, mock.MatchedBy(func(kpiData *domain.ImmobilizedInventoryKPI) bool {
		return kpiData.OrganizationID == givenOrgID && kpiData.ImmobilizedCount == 1
	})).Return(nil)

	useCase := kpi.NewCalculateImmobilizedInventoryUseCase(mockKPIRepo, mockCatalogClient)

	input := &kpi.CalculateImmobilizedInventoryInput{
		OrganizationID: givenOrgID,
		SnapshotDate:   givenDate,
		ThresholdYears: 2,
	}

	result, err := useCase.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, givenOrgID, result.OrganizationID)
	assert.Equal(t, 1, result.ImmobilizedCount)
	mockKPIRepo.AssertExpectations(t)
	mockCatalogClient.AssertExpectations(t)
}

func TestCalculateImmobilizedInventoryUseCase_Execute_WithNilInput_ReturnsError(t *testing.T) {
	mockKPIRepo := new(MockKPIRepository)
	mockCatalogClient := new(MockCatalogServiceClient)

	useCase := kpi.NewCalculateImmobilizedInventoryUseCase(mockKPIRepo, mockCatalogClient)

	result, err := useCase.Execute(context.Background(), nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "input cannot be nil")
}

func TestCalculateImmobilizedInventoryUseCase_Execute_WithNilOrganizationID_ReturnsError(t *testing.T) {
	mockKPIRepo := new(MockKPIRepository)
	mockCatalogClient := new(MockCatalogServiceClient)

	useCase := kpi.NewCalculateImmobilizedInventoryUseCase(mockKPIRepo, mockCatalogClient)

	input := &kpi.CalculateImmobilizedInventoryInput{
		OrganizationID: uuid.Nil,
		SnapshotDate:   time.Now(),
		ThresholdYears: 2,
	}

	result, err := useCase.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "organization_id is required")
}

func TestCalculateImmobilizedInventoryUseCase_Execute_WithZeroSnapshotDate_ReturnsError(t *testing.T) {
	mockKPIRepo := new(MockKPIRepository)
	mockCatalogClient := new(MockCatalogServiceClient)

	useCase := kpi.NewCalculateImmobilizedInventoryUseCase(mockKPIRepo, mockCatalogClient)

	input := &kpi.CalculateImmobilizedInventoryInput{
		OrganizationID: uuid.New(),
		SnapshotDate:   time.Time{},
		ThresholdYears: 2,
	}

	result, err := useCase.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "snapshot_date is required")
}

func TestCalculateImmobilizedInventoryUseCase_Execute_WithZeroThresholdYears_ReturnsError(t *testing.T) {
	mockKPIRepo := new(MockKPIRepository)
	mockCatalogClient := new(MockCatalogServiceClient)

	useCase := kpi.NewCalculateImmobilizedInventoryUseCase(mockKPIRepo, mockCatalogClient)

	input := &kpi.CalculateImmobilizedInventoryInput{
		OrganizationID: uuid.New(),
		SnapshotDate:   time.Now(),
		ThresholdYears: 0,
	}

	result, err := useCase.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "threshold_years must be positive")
}

func TestCalculateImmobilizedInventoryUseCase_Execute_WhenCatalogClientFails_ReturnsError(t *testing.T) {
	givenOrgID := uuid.New()

	mockKPIRepo := new(MockKPIRepository)
	mockCatalogClient := new(MockCatalogServiceClient)

	mockCatalogClient.On("ListProductsWithInventory", mock.Anything, givenOrgID).Return(nil, assert.AnError)

	useCase := kpi.NewCalculateImmobilizedInventoryUseCase(mockKPIRepo, mockCatalogClient)

	input := &kpi.CalculateImmobilizedInventoryInput{
		OrganizationID: givenOrgID,
		SnapshotDate:   time.Now(),
		ThresholdYears: 2,
	}

	result, err := useCase.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, result)
	mockCatalogClient.AssertExpectations(t)
}

func TestCalculateImmobilizedInventoryUseCase_Execute_WhenRepositorySaveFails_ReturnsError(t *testing.T) {
	givenOrgID := uuid.New()
	givenDate := time.Now()

	givenProducts := []*providers.ProductWithInventory{
		{
			ProductID:        uuid.New(),
			SKU:              "P001",
			Name:             "Product 1",
			Quantity:         100,
			StandardCost:     50,
			LastPurchaseDate: &givenDate,
		},
	}

	mockKPIRepo := new(MockKPIRepository)
	mockCatalogClient := new(MockCatalogServiceClient)

	mockCatalogClient.On("ListProductsWithInventory", mock.Anything, givenOrgID).Return(givenProducts, nil)
	mockKPIRepo.On("SaveImmobilizedInventoryKPI", mock.Anything, mock.Anything).Return(assert.AnError)

	useCase := kpi.NewCalculateImmobilizedInventoryUseCase(mockKPIRepo, mockCatalogClient)

	input := &kpi.CalculateImmobilizedInventoryInput{
		OrganizationID: givenOrgID,
		SnapshotDate:   givenDate,
		ThresholdYears: 2,
	}

	result, err := useCase.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, result)
	mockKPIRepo.AssertExpectations(t)
	mockCatalogClient.AssertExpectations(t)
}
