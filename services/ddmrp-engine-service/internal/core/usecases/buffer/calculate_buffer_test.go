package buffer

import (
	"context"
	"testing"
	"time"

	"github.com/giia/giia-core-engine/services/ddmrp-engine-service/internal/core/domain"
	"github.com/giia/giia-core-engine/services/ddmrp-engine-service/internal/core/providers"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockBufferRepository struct {
	mock.Mock
}

func (m *MockBufferRepository) Create(ctx context.Context, buffer *domain.Buffer) error {
	args := m.Called(ctx, buffer)
	return args.Error(0)
}

func (m *MockBufferRepository) Save(ctx context.Context, buffer *domain.Buffer) error {
	args := m.Called(ctx, buffer)
	return args.Error(0)
}

func (m *MockBufferRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Buffer, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Buffer), args.Error(1)
}

func (m *MockBufferRepository) GetByProduct(ctx context.Context, productID, organizationID uuid.UUID) (*domain.Buffer, error) {
	args := m.Called(ctx, productID, organizationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Buffer), args.Error(1)
}

func (m *MockBufferRepository) List(ctx context.Context, organizationID uuid.UUID, limit, offset int) ([]domain.Buffer, error) {
	args := m.Called(ctx, organizationID, limit, offset)
	return args.Get(0).([]domain.Buffer), args.Error(1)
}

func (m *MockBufferRepository) ListByZone(ctx context.Context, organizationID uuid.UUID, zone domain.ZoneType) ([]domain.Buffer, error) {
	args := m.Called(ctx, organizationID, zone)
	return args.Get(0).([]domain.Buffer), args.Error(1)
}

func (m *MockBufferRepository) ListByAlertLevel(ctx context.Context, organizationID uuid.UUID, alertLevel domain.AlertLevel) ([]domain.Buffer, error) {
	args := m.Called(ctx, organizationID, alertLevel)
	return args.Get(0).([]domain.Buffer), args.Error(1)
}

func (m *MockBufferRepository) ListAll(ctx context.Context, organizationID uuid.UUID) ([]domain.Buffer, error) {
	args := m.Called(ctx, organizationID)
	return args.Get(0).([]domain.Buffer), args.Error(1)
}

func (m *MockBufferRepository) UpdateNFP(ctx context.Context, bufferID uuid.UUID, onHand, onOrder, qualifiedDemand float64) error {
	args := m.Called(ctx, bufferID, onHand, onOrder, qualifiedDemand)
	return args.Error(0)
}

func (m *MockBufferRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockDemandAdjustmentRepository struct {
	mock.Mock
}

func (m *MockDemandAdjustmentRepository) Create(ctx context.Context, adjustment *domain.DemandAdjustment) error {
	args := m.Called(ctx, adjustment)
	return args.Error(0)
}

func (m *MockDemandAdjustmentRepository) Update(ctx context.Context, adjustment *domain.DemandAdjustment) error {
	args := m.Called(ctx, adjustment)
	return args.Error(0)
}

func (m *MockDemandAdjustmentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.DemandAdjustment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DemandAdjustment), args.Error(1)
}

func (m *MockDemandAdjustmentRepository) GetActiveForDate(ctx context.Context, productID, organizationID uuid.UUID, date time.Time) ([]domain.DemandAdjustment, error) {
	args := m.Called(ctx, productID, organizationID, mock.Anything)
	return args.Get(0).([]domain.DemandAdjustment), args.Error(1)
}

func (m *MockDemandAdjustmentRepository) ListByProduct(ctx context.Context, productID, organizationID uuid.UUID) ([]domain.DemandAdjustment, error) {
	args := m.Called(ctx, productID, organizationID)
	return args.Get(0).([]domain.DemandAdjustment), args.Error(1)
}

func (m *MockDemandAdjustmentRepository) ListByOrganization(ctx context.Context, organizationID uuid.UUID, limit, offset int) ([]domain.DemandAdjustment, error) {
	args := m.Called(ctx, organizationID, limit, offset)
	return args.Get(0).([]domain.DemandAdjustment), args.Error(1)
}

func (m *MockDemandAdjustmentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockBufferAdjustmentRepository struct {
	mock.Mock
}

func (m *MockBufferAdjustmentRepository) Create(ctx context.Context, adjustment *domain.BufferAdjustment) error {
	args := m.Called(ctx, adjustment)
	return args.Error(0)
}

func (m *MockBufferAdjustmentRepository) Update(ctx context.Context, adjustment *domain.BufferAdjustment) error {
	args := m.Called(ctx, adjustment)
	return args.Error(0)
}

func (m *MockBufferAdjustmentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.BufferAdjustment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.BufferAdjustment), args.Error(1)
}

func (m *MockBufferAdjustmentRepository) GetActiveForDate(ctx context.Context, bufferID uuid.UUID, date time.Time) ([]domain.BufferAdjustment, error) {
	args := m.Called(ctx, bufferID, mock.Anything)
	return args.Get(0).([]domain.BufferAdjustment), args.Error(1)
}

func (m *MockBufferAdjustmentRepository) ListByBuffer(ctx context.Context, bufferID uuid.UUID) ([]domain.BufferAdjustment, error) {
	args := m.Called(ctx, bufferID)
	return args.Get(0).([]domain.BufferAdjustment), args.Error(1)
}

func (m *MockBufferAdjustmentRepository) ListByProduct(ctx context.Context, productID, organizationID uuid.UUID) ([]domain.BufferAdjustment, error) {
	args := m.Called(ctx, productID, organizationID)
	return args.Get(0).([]domain.BufferAdjustment), args.Error(1)
}

func (m *MockBufferAdjustmentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockBufferHistoryRepository struct {
	mock.Mock
}

func (m *MockBufferHistoryRepository) Create(ctx context.Context, history *domain.BufferHistory) error {
	args := m.Called(ctx, history)
	return args.Error(0)
}

func (m *MockBufferHistoryRepository) GetByBufferAndDate(ctx context.Context, bufferID uuid.UUID, date time.Time) (*domain.BufferHistory, error) {
	args := m.Called(ctx, bufferID, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.BufferHistory), args.Error(1)
}

func (m *MockBufferHistoryRepository) ListByBuffer(ctx context.Context, bufferID uuid.UUID, limit int) ([]domain.BufferHistory, error) {
	args := m.Called(ctx, bufferID, limit)
	return args.Get(0).([]domain.BufferHistory), args.Error(1)
}

func (m *MockBufferHistoryRepository) ListByProduct(ctx context.Context, productID, organizationID uuid.UUID, startDate, endDate time.Time) ([]domain.BufferHistory, error) {
	args := m.Called(ctx, productID, organizationID, startDate, endDate)
	return args.Get(0).([]domain.BufferHistory), args.Error(1)
}

func (m *MockBufferHistoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockCatalogClient struct {
	mock.Mock
}

func (m *MockCatalogClient) GetProduct(ctx context.Context, productID uuid.UUID) (*providers.Product, error) {
	args := m.Called(ctx, productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*providers.Product), args.Error(1)
}

func (m *MockCatalogClient) GetBufferProfile(ctx context.Context, bufferProfileID uuid.UUID) (*providers.BufferProfile, error) {
	args := m.Called(ctx, bufferProfileID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*providers.BufferProfile), args.Error(1)
}

func (m *MockCatalogClient) GetSupplier(ctx context.Context, supplierID uuid.UUID) (*providers.Supplier, error) {
	args := m.Called(ctx, supplierID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*providers.Supplier), args.Error(1)
}

func (m *MockCatalogClient) GetProductSuppliers(ctx context.Context, productID uuid.UUID) ([]providers.ProductSupplier, error) {
	args := m.Called(ctx, productID)
	return args.Get(0).([]providers.ProductSupplier), args.Error(1)
}

func (m *MockCatalogClient) GetPrimarySupplier(ctx context.Context, productID uuid.UUID) (*providers.ProductSupplier, error) {
	args := m.Called(ctx, productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*providers.ProductSupplier), args.Error(1)
}

type MockADURepository struct {
	mock.Mock
}

func (m *MockADURepository) Create(ctx context.Context, adu *domain.ADUCalculation) error {
	args := m.Called(ctx, adu)
	return args.Error(0)
}

func (m *MockADURepository) GetLatest(ctx context.Context, productID, organizationID uuid.UUID) (*domain.ADUCalculation, error) {
	args := m.Called(ctx, productID, organizationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ADUCalculation), args.Error(1)
}

func (m *MockADURepository) GetByDate(ctx context.Context, productID, organizationID uuid.UUID, date time.Time) (*domain.ADUCalculation, error) {
	args := m.Called(ctx, productID, organizationID, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ADUCalculation), args.Error(1)
}

func (m *MockADURepository) ListHistory(ctx context.Context, productID, organizationID uuid.UUID, limit int) ([]domain.ADUCalculation, error) {
	args := m.Called(ctx, productID, organizationID, limit)
	return args.Get(0).([]domain.ADUCalculation), args.Error(1)
}

func (m *MockADURepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockEventPublisher struct {
	mock.Mock
}

func (m *MockEventPublisher) PublishBufferCalculated(ctx context.Context, buffer *domain.Buffer) error {
	args := m.Called(ctx, buffer)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishBufferStatusChanged(ctx context.Context, buffer *domain.Buffer, oldZone domain.ZoneType) error {
	args := m.Called(ctx, buffer, oldZone)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishBufferAlertTriggered(ctx context.Context, buffer *domain.Buffer) error {
	args := m.Called(ctx, buffer)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishFADCreated(ctx context.Context, fad *domain.DemandAdjustment) error {
	args := m.Called(ctx, fad)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishFADUpdated(ctx context.Context, fad *domain.DemandAdjustment) error {
	args := m.Called(ctx, fad)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishFADDeleted(ctx context.Context, fadID string) error {
	args := m.Called(ctx, fadID)
	return args.Error(0)
}

func TestCalculateBuffer_Success_WithFAD(t *testing.T) {
	mockBufferRepo := new(MockBufferRepository)
	mockDemandAdjRepo := new(MockDemandAdjustmentRepository)
	mockBufferAdjRepo := new(MockBufferAdjustmentRepository)
	mockBufferHistoryRepo := new(MockBufferHistoryRepository)
	mockCatalogClient := new(MockCatalogClient)
	mockADURepo := new(MockADURepository)
	mockEventPublisher := new(MockEventPublisher)

	useCase := NewCalculateBufferUseCase(
		mockBufferRepo,
		mockDemandAdjRepo,
		mockBufferAdjRepo,
		mockBufferHistoryRepo,
		mockCatalogClient,
		mockADURepo,
		mockEventPublisher,
	)

	givenProductID := uuid.New()
	givenOrgID := uuid.New()
	givenBufferProfileID := uuid.New()
	givenBaseCPD := 100.0
	givenFADFactor := 1.5

	mockCatalogClient.On("GetProduct", mock.Anything, givenProductID).
		Return(&providers.Product{
			ID:              givenProductID,
			BufferProfileID: &givenBufferProfileID,
		}, nil)

	mockCatalogClient.On("GetBufferProfile", mock.Anything, givenBufferProfileID).
		Return(&providers.BufferProfile{
			LeadTimeFactor:    0.5,
			VariabilityFactor: 0.5,
			OrderFrequency:    7,
		}, nil)

	mockADURepo.On("GetLatest", mock.Anything, givenProductID, givenOrgID).
		Return(&domain.ADUCalculation{
			ADUValue: givenBaseCPD,
		}, nil)

	mockDemandAdjRepo.On("GetActiveForDate", mock.Anything, givenProductID, givenOrgID, mock.Anything).
		Return([]domain.DemandAdjustment{
			{Factor: givenFADFactor},
		}, nil)

	mockCatalogClient.On("GetPrimarySupplier", mock.Anything, givenProductID).
		Return(&providers.ProductSupplier{
			MOQ:          100,
			LeadTimeDays: 30,
		}, nil)

	mockBufferRepo.On("GetByProduct", mock.Anything, givenProductID, givenOrgID).
		Return(nil, domain.NewNotFoundError("buffer not found"))

	mockBufferAdjRepo.On("GetActiveForDate", mock.Anything, mock.Anything, mock.Anything).
		Return([]domain.BufferAdjustment{}, nil)

	mockBufferRepo.On("Save", mock.Anything, mock.AnythingOfType("*domain.Buffer")).
		Return(nil)

	mockBufferHistoryRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.BufferHistory")).
		Return(nil)

	mockEventPublisher.On("PublishBufferCalculated", mock.Anything, mock.AnythingOfType("*domain.Buffer")).
		Return(nil)

	mockEventPublisher.On("PublishBufferStatusChanged", mock.Anything, mock.AnythingOfType("*domain.Buffer"), mock.AnythingOfType("domain.ZoneType")).
		Return(nil)

	mockEventPublisher.On("PublishBufferAlertTriggered", mock.Anything, mock.AnythingOfType("*domain.Buffer")).
		Return(nil)

	result, err := useCase.Execute(context.Background(), CalculateBufferInput{
		ProductID:      givenProductID,
		OrganizationID: givenOrgID,
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 150.0, result.CPD)

	mockBufferRepo.AssertExpectations(t)
	mockDemandAdjRepo.AssertExpectations(t)
	mockCatalogClient.AssertExpectations(t)
	mockADURepo.AssertExpectations(t)
	mockEventPublisher.AssertExpectations(t)
}

func TestCalculateBuffer_Failure_ProductNotFound(t *testing.T) {
	mockBufferRepo := new(MockBufferRepository)
	mockDemandAdjRepo := new(MockDemandAdjustmentRepository)
	mockBufferAdjRepo := new(MockBufferAdjustmentRepository)
	mockBufferHistoryRepo := new(MockBufferHistoryRepository)
	mockCatalogClient := new(MockCatalogClient)
	mockADURepo := new(MockADURepository)
	mockEventPublisher := new(MockEventPublisher)

	useCase := NewCalculateBufferUseCase(
		mockBufferRepo,
		mockDemandAdjRepo,
		mockBufferAdjRepo,
		mockBufferHistoryRepo,
		mockCatalogClient,
		mockADURepo,
		mockEventPublisher,
	)

	givenProductID := uuid.New()
	givenOrgID := uuid.New()

	mockCatalogClient.On("GetProduct", mock.Anything, givenProductID).
		Return(nil, domain.NewNotFoundError("product not found"))

	result, err := useCase.Execute(context.Background(), CalculateBufferInput{
		ProductID:      givenProductID,
		OrganizationID: givenOrgID,
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get product from catalog")

	mockCatalogClient.AssertExpectations(t)
}
