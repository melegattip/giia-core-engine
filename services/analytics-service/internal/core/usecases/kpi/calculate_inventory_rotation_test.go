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

type MockExecutionServiceClient struct {
	mock.Mock
}

func (m *MockExecutionServiceClient) GetSalesData(ctx context.Context, organizationID uuid.UUID, startDate, endDate time.Time) (*providers.SalesData, error) {
	args := m.Called(ctx, organizationID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*providers.SalesData), args.Error(1)
}

func (m *MockExecutionServiceClient) GetInventorySnapshots(ctx context.Context, organizationID uuid.UUID, startDate, endDate time.Time) ([]*providers.InventorySnapshot, error) {
	args := m.Called(ctx, organizationID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*providers.InventorySnapshot), args.Error(1)
}

func (m *MockExecutionServiceClient) GetProductSales(ctx context.Context, organizationID uuid.UUID, startDate, endDate time.Time) ([]*providers.ProductSales, error) {
	args := m.Called(ctx, organizationID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*providers.ProductSales), args.Error(1)
}

func TestCalculateInventoryRotationUseCase_Execute_WithValidData_CalculatesKPI(t *testing.T) {
	givenOrgID := uuid.New()
	givenDate := time.Now()
	startDate := givenDate.AddDate(0, 0, -30)

	givenSalesData := &providers.SalesData{
		TotalValue: 150000.0,
		OrderCount: 100,
	}

	givenSnapshots := []*providers.InventorySnapshot{
		{Date: startDate.AddDate(0, 0, 1), TotalValue: 50000},
		{Date: startDate.AddDate(0, 0, 15), TotalValue: 50000},
		{Date: givenDate, TotalValue: 50000},
	}

	givenProductSales := []*providers.ProductSales{
		{ProductID: uuid.New(), SKU: "P001", Name: "Product 1", Sales30Days: 10000, AvgStockValue: 2000},
		{ProductID: uuid.New(), SKU: "P002", Name: "Product 2", Sales30Days: 100, AvgStockValue: 1000},
	}

	mockKPIRepo := new(MockKPIRepository)
	mockExecutionClient := new(MockExecutionServiceClient)

	mockExecutionClient.On("GetSalesData", mock.Anything, givenOrgID, mock.Anything, givenDate).Return(givenSalesData, nil)
	mockExecutionClient.On("GetInventorySnapshots", mock.Anything, givenOrgID, mock.Anything, givenDate).Return(givenSnapshots, nil)
	mockExecutionClient.On("GetProductSales", mock.Anything, givenOrgID, mock.Anything, givenDate).Return(givenProductSales, nil)
	mockKPIRepo.On("SaveInventoryRotationKPI", mock.Anything, mock.MatchedBy(func(kpiData *domain.InventoryRotationKPI) bool {
		return kpiData.OrganizationID == givenOrgID
	})).Return(nil)

	useCase := kpi.NewCalculateInventoryRotationUseCase(mockKPIRepo, mockExecutionClient)

	input := &kpi.CalculateInventoryRotationInput{
		OrganizationID: givenOrgID,
		SnapshotDate:   givenDate,
	}

	result, err := useCase.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, givenOrgID, result.OrganizationID)
	assert.Equal(t, 150000.0, result.SalesLast30Days)
	mockKPIRepo.AssertExpectations(t)
	mockExecutionClient.AssertExpectations(t)
}

func TestCalculateInventoryRotationUseCase_Execute_WithNilInput_ReturnsError(t *testing.T) {
	mockKPIRepo := new(MockKPIRepository)
	mockExecutionClient := new(MockExecutionServiceClient)

	useCase := kpi.NewCalculateInventoryRotationUseCase(mockKPIRepo, mockExecutionClient)

	result, err := useCase.Execute(context.Background(), nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "input cannot be nil")
}

func TestCalculateInventoryRotationUseCase_Execute_WithNilOrganizationID_ReturnsError(t *testing.T) {
	mockKPIRepo := new(MockKPIRepository)
	mockExecutionClient := new(MockExecutionServiceClient)

	useCase := kpi.NewCalculateInventoryRotationUseCase(mockKPIRepo, mockExecutionClient)

	input := &kpi.CalculateInventoryRotationInput{
		OrganizationID: uuid.Nil,
		SnapshotDate:   time.Now(),
	}

	result, err := useCase.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "organization_id is required")
}

func TestCalculateInventoryRotationUseCase_Execute_WithZeroSnapshotDate_ReturnsError(t *testing.T) {
	mockKPIRepo := new(MockKPIRepository)
	mockExecutionClient := new(MockExecutionServiceClient)

	useCase := kpi.NewCalculateInventoryRotationUseCase(mockKPIRepo, mockExecutionClient)

	input := &kpi.CalculateInventoryRotationInput{
		OrganizationID: uuid.New(),
		SnapshotDate:   time.Time{},
	}

	result, err := useCase.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "snapshot_date is required")
}

func TestCalculateInventoryRotationUseCase_Execute_WhenGetSalesDataFails_ReturnsError(t *testing.T) {
	givenOrgID := uuid.New()

	mockKPIRepo := new(MockKPIRepository)
	mockExecutionClient := new(MockExecutionServiceClient)

	mockExecutionClient.On("GetSalesData", mock.Anything, givenOrgID, mock.Anything, mock.Anything).Return(nil, assert.AnError)

	useCase := kpi.NewCalculateInventoryRotationUseCase(mockKPIRepo, mockExecutionClient)

	input := &kpi.CalculateInventoryRotationInput{
		OrganizationID: givenOrgID,
		SnapshotDate:   time.Now(),
	}

	result, err := useCase.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, result)
	mockExecutionClient.AssertExpectations(t)
}

func TestCalculateInventoryRotationUseCase_Execute_WhenGetInventorySnapshotsFails_ReturnsError(t *testing.T) {
	givenOrgID := uuid.New()
	givenDate := time.Now()

	givenSalesData := &providers.SalesData{TotalValue: 150000.0}

	mockKPIRepo := new(MockKPIRepository)
	mockExecutionClient := new(MockExecutionServiceClient)

	mockExecutionClient.On("GetSalesData", mock.Anything, givenOrgID, mock.Anything, givenDate).Return(givenSalesData, nil)
	mockExecutionClient.On("GetInventorySnapshots", mock.Anything, givenOrgID, mock.Anything, givenDate).Return(nil, assert.AnError)

	useCase := kpi.NewCalculateInventoryRotationUseCase(mockKPIRepo, mockExecutionClient)

	input := &kpi.CalculateInventoryRotationInput{
		OrganizationID: givenOrgID,
		SnapshotDate:   givenDate,
	}

	result, err := useCase.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, result)
	mockExecutionClient.AssertExpectations(t)
}

func TestCalculateInventoryRotationUseCase_Execute_WhenGetProductSalesFails_ReturnsError(t *testing.T) {
	givenOrgID := uuid.New()
	givenDate := time.Now()

	givenSalesData := &providers.SalesData{TotalValue: 150000.0}
	givenSnapshots := []*providers.InventorySnapshot{{Date: givenDate, TotalValue: 50000}}

	mockKPIRepo := new(MockKPIRepository)
	mockExecutionClient := new(MockExecutionServiceClient)

	mockExecutionClient.On("GetSalesData", mock.Anything, givenOrgID, mock.Anything, givenDate).Return(givenSalesData, nil)
	mockExecutionClient.On("GetInventorySnapshots", mock.Anything, givenOrgID, mock.Anything, givenDate).Return(givenSnapshots, nil)
	mockExecutionClient.On("GetProductSales", mock.Anything, givenOrgID, mock.Anything, givenDate).Return(nil, assert.AnError)

	useCase := kpi.NewCalculateInventoryRotationUseCase(mockKPIRepo, mockExecutionClient)

	input := &kpi.CalculateInventoryRotationInput{
		OrganizationID: givenOrgID,
		SnapshotDate:   givenDate,
	}

	result, err := useCase.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, result)
	mockExecutionClient.AssertExpectations(t)
}

func TestCalculateInventoryRotationUseCase_Execute_WhenRepositorySaveFails_ReturnsError(t *testing.T) {
	givenOrgID := uuid.New()
	givenDate := time.Now()

	givenSalesData := &providers.SalesData{TotalValue: 150000.0}
	givenSnapshots := []*providers.InventorySnapshot{{Date: givenDate, TotalValue: 50000}}
	givenProductSales := []*providers.ProductSales{}

	mockKPIRepo := new(MockKPIRepository)
	mockExecutionClient := new(MockExecutionServiceClient)

	mockExecutionClient.On("GetSalesData", mock.Anything, givenOrgID, mock.Anything, givenDate).Return(givenSalesData, nil)
	mockExecutionClient.On("GetInventorySnapshots", mock.Anything, givenOrgID, mock.Anything, givenDate).Return(givenSnapshots, nil)
	mockExecutionClient.On("GetProductSales", mock.Anything, givenOrgID, mock.Anything, givenDate).Return(givenProductSales, nil)
	mockKPIRepo.On("SaveInventoryRotationKPI", mock.Anything, mock.Anything).Return(assert.AnError)

	useCase := kpi.NewCalculateInventoryRotationUseCase(mockKPIRepo, mockExecutionClient)

	input := &kpi.CalculateInventoryRotationInput{
		OrganizationID: givenOrgID,
		SnapshotDate:   givenDate,
	}

	result, err := useCase.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, result)
	mockKPIRepo.AssertExpectations(t)
	mockExecutionClient.AssertExpectations(t)
}
