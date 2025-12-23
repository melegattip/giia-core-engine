package kpi_test

import (
	"context"
	"testing"
	"time"

	"github.com/giia/giia-core-engine/services/analytics-service/internal/core/domain"
	"github.com/giia/giia-core-engine/services/analytics-service/internal/core/providers"
	"github.com/giia/giia-core-engine/services/analytics-service/internal/core/usecases/kpi"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockKPIRepository struct {
	mock.Mock
}

func (m *MockKPIRepository) SaveKPISnapshot(ctx context.Context, snapshot *domain.KPISnapshot) error {
	args := m.Called(ctx, snapshot)
	return args.Error(0)
}

func (m *MockKPIRepository) GetKPISnapshot(ctx context.Context, organizationID uuid.UUID, date time.Time) (*domain.KPISnapshot, error) {
	args := m.Called(ctx, organizationID, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.KPISnapshot), args.Error(1)
}

func (m *MockKPIRepository) ListKPISnapshots(ctx context.Context, organizationID uuid.UUID, startDate, endDate time.Time) ([]*domain.KPISnapshot, error) {
	args := m.Called(ctx, organizationID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.KPISnapshot), args.Error(1)
}

func (m *MockKPIRepository) SaveDaysInInventoryKPI(ctx context.Context, kpiData *domain.DaysInInventoryKPI) error {
	args := m.Called(ctx, kpiData)
	return args.Error(0)
}

func (m *MockKPIRepository) GetDaysInInventoryKPI(ctx context.Context, organizationID uuid.UUID, date time.Time) (*domain.DaysInInventoryKPI, error) {
	args := m.Called(ctx, organizationID, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DaysInInventoryKPI), args.Error(1)
}

func (m *MockKPIRepository) ListDaysInInventoryKPI(ctx context.Context, organizationID uuid.UUID, startDate, endDate time.Time) ([]*domain.DaysInInventoryKPI, error) {
	args := m.Called(ctx, organizationID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.DaysInInventoryKPI), args.Error(1)
}

func (m *MockKPIRepository) SaveImmobilizedInventoryKPI(ctx context.Context, kpiData *domain.ImmobilizedInventoryKPI) error {
	args := m.Called(ctx, kpiData)
	return args.Error(0)
}

func (m *MockKPIRepository) GetImmobilizedInventoryKPI(ctx context.Context, organizationID uuid.UUID, date time.Time, thresholdYears int) (*domain.ImmobilizedInventoryKPI, error) {
	args := m.Called(ctx, organizationID, date, thresholdYears)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ImmobilizedInventoryKPI), args.Error(1)
}

func (m *MockKPIRepository) ListImmobilizedInventoryKPI(ctx context.Context, organizationID uuid.UUID, startDate, endDate time.Time, thresholdYears int) ([]*domain.ImmobilizedInventoryKPI, error) {
	args := m.Called(ctx, organizationID, startDate, endDate, thresholdYears)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.ImmobilizedInventoryKPI), args.Error(1)
}

func (m *MockKPIRepository) SaveInventoryRotationKPI(ctx context.Context, kpiData *domain.InventoryRotationKPI) error {
	args := m.Called(ctx, kpiData)
	return args.Error(0)
}

func (m *MockKPIRepository) GetInventoryRotationKPI(ctx context.Context, organizationID uuid.UUID, date time.Time) (*domain.InventoryRotationKPI, error) {
	args := m.Called(ctx, organizationID, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.InventoryRotationKPI), args.Error(1)
}

func (m *MockKPIRepository) ListInventoryRotationKPI(ctx context.Context, organizationID uuid.UUID, startDate, endDate time.Time) ([]*domain.InventoryRotationKPI, error) {
	args := m.Called(ctx, organizationID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.InventoryRotationKPI), args.Error(1)
}

func (m *MockKPIRepository) SaveBufferAnalytics(ctx context.Context, analytics *domain.BufferAnalytics) error {
	args := m.Called(ctx, analytics)
	return args.Error(0)
}

func (m *MockKPIRepository) GetBufferAnalyticsByProduct(ctx context.Context, organizationID, productID uuid.UUID, date time.Time) (*domain.BufferAnalytics, error) {
	args := m.Called(ctx, organizationID, productID, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.BufferAnalytics), args.Error(1)
}

func (m *MockKPIRepository) ListBufferAnalytics(ctx context.Context, organizationID uuid.UUID, startDate, endDate time.Time) ([]*domain.BufferAnalytics, error) {
	args := m.Called(ctx, organizationID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.BufferAnalytics), args.Error(1)
}

type MockCatalogServiceClient struct {
	mock.Mock
}

func (m *MockCatalogServiceClient) ListProductsWithInventory(ctx context.Context, organizationID uuid.UUID) ([]*providers.ProductWithInventory, error) {
	args := m.Called(ctx, organizationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*providers.ProductWithInventory), args.Error(1)
}

func (m *MockCatalogServiceClient) GetProduct(ctx context.Context, productID uuid.UUID) (*providers.ProductWithInventory, error) {
	args := m.Called(ctx, productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*providers.ProductWithInventory), args.Error(1)
}

func TestCalculateDaysInInventoryUseCase_Execute_WithValidData_CalculatesKPI(t *testing.T) {
	givenOrgID := uuid.New()
	givenDate := time.Now()
	purchaseDate := givenDate.AddDate(0, 0, -30)

	givenProducts := []*providers.ProductWithInventory{
		{
			ProductID:        uuid.New(),
			SKU:              "P001",
			Name:             "Product 1",
			Quantity:         100,
			StandardCost:     50,
			LastPurchaseDate: &purchaseDate,
		},
		{
			ProductID:        uuid.New(),
			SKU:              "P002",
			Name:             "Product 2",
			Quantity:         200,
			StandardCost:     75,
			LastPurchaseDate: &purchaseDate,
		},
	}

	mockKPIRepo := new(MockKPIRepository)
	mockCatalogClient := new(MockCatalogServiceClient)

	mockCatalogClient.On("ListProductsWithInventory", mock.Anything, givenOrgID).Return(givenProducts, nil)
	mockKPIRepo.On("SaveDaysInInventoryKPI", mock.Anything, mock.MatchedBy(func(kpiData *domain.DaysInInventoryKPI) bool {
		return kpiData.OrganizationID == givenOrgID && kpiData.TotalProducts == 2
	})).Return(nil)

	useCase := kpi.NewCalculateDaysInInventoryUseCase(mockKPIRepo, mockCatalogClient)

	input := &kpi.CalculateDaysInInventoryInput{
		OrganizationID: givenOrgID,
		SnapshotDate:   givenDate,
	}

	result, err := useCase.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, givenOrgID, result.OrganizationID)
	assert.Equal(t, 2, result.TotalProducts)
	mockKPIRepo.AssertExpectations(t)
	mockCatalogClient.AssertExpectations(t)
}

func TestCalculateDaysInInventoryUseCase_Execute_WithNilInput_ReturnsError(t *testing.T) {
	mockKPIRepo := new(MockKPIRepository)
	mockCatalogClient := new(MockCatalogServiceClient)

	useCase := kpi.NewCalculateDaysInInventoryUseCase(mockKPIRepo, mockCatalogClient)

	result, err := useCase.Execute(context.Background(), nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "input cannot be nil")
}

func TestCalculateDaysInInventoryUseCase_Execute_WithNilOrganizationID_ReturnsError(t *testing.T) {
	mockKPIRepo := new(MockKPIRepository)
	mockCatalogClient := new(MockCatalogServiceClient)

	useCase := kpi.NewCalculateDaysInInventoryUseCase(mockKPIRepo, mockCatalogClient)

	input := &kpi.CalculateDaysInInventoryInput{
		OrganizationID: uuid.Nil,
		SnapshotDate:   time.Now(),
	}

	result, err := useCase.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "organization_id is required")
}

func TestCalculateDaysInInventoryUseCase_Execute_WithZeroSnapshotDate_ReturnsError(t *testing.T) {
	mockKPIRepo := new(MockKPIRepository)
	mockCatalogClient := new(MockCatalogServiceClient)

	useCase := kpi.NewCalculateDaysInInventoryUseCase(mockKPIRepo, mockCatalogClient)

	input := &kpi.CalculateDaysInInventoryInput{
		OrganizationID: uuid.New(),
		SnapshotDate:   time.Time{},
	}

	result, err := useCase.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "snapshot_date is required")
}

func TestCalculateDaysInInventoryUseCase_Execute_WhenCatalogClientFails_ReturnsError(t *testing.T) {
	givenOrgID := uuid.New()

	mockKPIRepo := new(MockKPIRepository)
	mockCatalogClient := new(MockCatalogServiceClient)

	mockCatalogClient.On("ListProductsWithInventory", mock.Anything, givenOrgID).Return(nil, assert.AnError)

	useCase := kpi.NewCalculateDaysInInventoryUseCase(mockKPIRepo, mockCatalogClient)

	input := &kpi.CalculateDaysInInventoryInput{
		OrganizationID: givenOrgID,
		SnapshotDate:   time.Now(),
	}

	result, err := useCase.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, result)
	mockCatalogClient.AssertExpectations(t)
}

func TestCalculateDaysInInventoryUseCase_Execute_WhenRepositorySaveFails_ReturnsError(t *testing.T) {
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
	mockKPIRepo.On("SaveDaysInInventoryKPI", mock.Anything, mock.Anything).Return(assert.AnError)

	useCase := kpi.NewCalculateDaysInInventoryUseCase(mockKPIRepo, mockCatalogClient)

	input := &kpi.CalculateDaysInInventoryInput{
		OrganizationID: givenOrgID,
		SnapshotDate:   givenDate,
	}

	result, err := useCase.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, result)
	mockKPIRepo.AssertExpectations(t)
	mockCatalogClient.AssertExpectations(t)
}

func TestCalculateDaysInInventoryUseCase_Execute_WithProductsWithoutPurchaseDate_SkipsThem(t *testing.T) {
	givenOrgID := uuid.New()
	givenDate := time.Now()

	givenProducts := []*providers.ProductWithInventory{
		{
			ProductID:        uuid.New(),
			SKU:              "P001",
			Name:             "Product 1",
			Quantity:         100,
			StandardCost:     50,
			LastPurchaseDate: nil,
		},
	}

	mockKPIRepo := new(MockKPIRepository)
	mockCatalogClient := new(MockCatalogServiceClient)

	mockCatalogClient.On("ListProductsWithInventory", mock.Anything, givenOrgID).Return(givenProducts, nil)
	mockKPIRepo.On("SaveDaysInInventoryKPI", mock.Anything, mock.MatchedBy(func(kpiData *domain.DaysInInventoryKPI) bool {
		return kpiData.TotalProducts == 0
	})).Return(nil)

	useCase := kpi.NewCalculateDaysInInventoryUseCase(mockKPIRepo, mockCatalogClient)

	input := &kpi.CalculateDaysInInventoryInput{
		OrganizationID: givenOrgID,
		SnapshotDate:   givenDate,
	}

	result, err := useCase.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0, result.TotalProducts)
	mockKPIRepo.AssertExpectations(t)
	mockCatalogClient.AssertExpectations(t)
}
