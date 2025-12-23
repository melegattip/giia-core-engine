package supplier

import (
	"context"
	"testing"

	"github.com/melegattip/giia-core-engine/pkg/errors"
	"github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/catalog-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/catalog-service/internal/core/providers"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockSupplierRepository struct {
	mock.Mock
}

func (m *MockSupplierRepository) Create(ctx context.Context, supplier *domain.Supplier) error {
	args := m.Called(ctx, supplier)
	return args.Error(0)
}

func (m *MockSupplierRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Supplier, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Supplier), args.Error(1)
}

func (m *MockSupplierRepository) GetByCode(ctx context.Context, code string) (*domain.Supplier, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Supplier), args.Error(1)
}

func (m *MockSupplierRepository) Update(ctx context.Context, supplier *domain.Supplier) error {
	args := m.Called(ctx, supplier)
	return args.Error(0)
}

func (m *MockSupplierRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSupplierRepository) List(ctx context.Context, filters map[string]interface{}, page, pageSize int) ([]*domain.Supplier, int64, error) {
	args := m.Called(ctx, filters, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*domain.Supplier), args.Get(1).(int64), args.Error(2)
}

type MockSupplierEventPublisher struct {
	mock.Mock
}

func (m *MockSupplierEventPublisher) PublishSupplierCreated(ctx context.Context, supplier *domain.Supplier) error {
	args := m.Called(ctx, supplier)
	return args.Error(0)
}

func (m *MockSupplierEventPublisher) PublishSupplierUpdated(ctx context.Context, supplier *domain.Supplier) error {
	args := m.Called(ctx, supplier)
	return args.Error(0)
}

func (m *MockSupplierEventPublisher) PublishSupplierDeleted(ctx context.Context, supplier *domain.Supplier) error {
	args := m.Called(ctx, supplier)
	return args.Error(0)
}

func (m *MockSupplierEventPublisher) PublishProductCreated(ctx context.Context, product *domain.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockSupplierEventPublisher) PublishProductUpdated(ctx context.Context, product *domain.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockSupplierEventPublisher) PublishProductDeleted(ctx context.Context, product *domain.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockSupplierEventPublisher) PublishBufferProfileCreated(ctx context.Context, profile *domain.BufferProfile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *MockSupplierEventPublisher) PublishBufferProfileUpdated(ctx context.Context, profile *domain.BufferProfile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *MockSupplierEventPublisher) PublishBufferProfileDeleted(ctx context.Context, profile *domain.BufferProfile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *MockSupplierEventPublisher) PublishBufferProfileAssigned(ctx context.Context, product *domain.Product, profile *domain.BufferProfile) error {
	args := m.Called(ctx, product, profile)
	return args.Error(0)
}

var _ providers.SupplierRepository = (*MockSupplierRepository)(nil)
var _ providers.EventPublisher = (*MockSupplierEventPublisher)(nil)

func TestCreateSupplierUseCase_Execute_WithValidData_ReturnsSupplier(t *testing.T) {
	givenOrgID := uuid.New()
	givenRequest := &CreateSupplierRequest{
		Code:              "SUP-001",
		Name:              "Premium Supplier",
		LeadTimeDays:      7,
		ReliabilityRating: 95,
	}

	mockRepo := new(MockSupplierRepository)
	mockPublisher := new(MockSupplierEventPublisher)
	mockLogger := logger.New("test", "debug")

	mockRepo.On("GetByCode", mock.Anything, "SUP-001").Return(nil, errors.NewNotFound("not found"))
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(s *domain.Supplier) bool {
		return s.Code == "SUP-001" && s.Name == "Premium Supplier"
	})).Return(nil)
	mockPublisher.On("PublishSupplierCreated", mock.Anything, mock.Anything).Return(nil)

	useCase := NewCreateSupplierUseCase(mockRepo, mockPublisher, mockLogger)

	ctx := context.WithValue(context.Background(), "organization_id", givenOrgID)
	supplier, err := useCase.Execute(ctx, givenRequest)

	assert.NoError(t, err)
	assert.NotNil(t, supplier)
	assert.Equal(t, "SUP-001", supplier.Code)
	assert.Equal(t, "Premium Supplier", supplier.Name)
	assert.Equal(t, 7, supplier.LeadTimeDays)
	assert.Equal(t, 95, supplier.ReliabilityRating)
	assert.Equal(t, domain.SupplierStatusActive, supplier.Status)
	mockRepo.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}

func TestCreateSupplierUseCase_Execute_WithNilRequest_ReturnsError(t *testing.T) {
	mockRepo := new(MockSupplierRepository)
	mockPublisher := new(MockSupplierEventPublisher)
	mockLogger := logger.New("test", "debug")

	useCase := NewCreateSupplierUseCase(mockRepo, mockPublisher, mockLogger)

	ctx := context.WithValue(context.Background(), "organization_id", uuid.New())
	supplier, err := useCase.Execute(ctx, nil)

	assert.Error(t, err)
	assert.Nil(t, supplier)
	assert.True(t, errors.IsBadRequest(err))
	mockRepo.AssertNotCalled(t, "Create")
}

func TestCreateSupplierUseCase_Execute_WithMissingOrgID_ReturnsError(t *testing.T) {
	mockRepo := new(MockSupplierRepository)
	mockPublisher := new(MockSupplierEventPublisher)
	mockLogger := logger.New("test", "debug")

	useCase := NewCreateSupplierUseCase(mockRepo, mockPublisher, mockLogger)

	request := &CreateSupplierRequest{
		Code:              "SUP-001",
		Name:              "Premium Supplier",
		LeadTimeDays:      7,
		ReliabilityRating: 95,
	}

	supplier, err := useCase.Execute(context.Background(), request)

	assert.Error(t, err)
	assert.Nil(t, supplier)
	assert.True(t, errors.IsBadRequest(err))
	mockRepo.AssertNotCalled(t, "Create")
}

func TestCreateSupplierUseCase_Execute_WithDuplicateCode_ReturnsConflictError(t *testing.T) {
	givenOrgID := uuid.New()
	givenExistingSupplier := &domain.Supplier{
		ID:             uuid.New(),
		Code:           "SUP-001",
		Name:           "Existing Supplier",
		Status:         domain.SupplierStatusActive,
		OrganizationID: givenOrgID,
	}

	mockRepo := new(MockSupplierRepository)
	mockPublisher := new(MockSupplierEventPublisher)
	mockLogger := logger.New("test", "debug")

	mockRepo.On("GetByCode", mock.Anything, "SUP-001").Return(givenExistingSupplier, nil)

	useCase := NewCreateSupplierUseCase(mockRepo, mockPublisher, mockLogger)

	ctx := context.WithValue(context.Background(), "organization_id", givenOrgID)
	request := &CreateSupplierRequest{
		Code:              "SUP-001",
		Name:              "New Supplier",
		LeadTimeDays:      5,
		ReliabilityRating: 90,
	}

	supplier, err := useCase.Execute(ctx, request)

	assert.Error(t, err)
	assert.Nil(t, supplier)
	mockRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "Create")
	mockPublisher.AssertNotCalled(t, "PublishSupplierCreated")
}

func TestCreateSupplierUseCase_Execute_WithEventPublishError_StillSucceeds(t *testing.T) {
	givenOrgID := uuid.New()
	givenEventError := errors.NewInternalServerError("event publish failed")

	mockRepo := new(MockSupplierRepository)
	mockPublisher := new(MockSupplierEventPublisher)
	mockLogger := logger.New("test", "debug")

	mockRepo.On("GetByCode", mock.Anything, "SUP-001").Return(nil, errors.NewNotFound("not found"))
	mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
	mockPublisher.On("PublishSupplierCreated", mock.Anything, mock.Anything).Return(givenEventError)

	useCase := NewCreateSupplierUseCase(mockRepo, mockPublisher, mockLogger)

	ctx := context.WithValue(context.Background(), "organization_id", givenOrgID)
	request := &CreateSupplierRequest{
		Code:              "SUP-001",
		Name:              "Test Supplier",
		LeadTimeDays:      10,
		ReliabilityRating: 80,
	}

	supplier, err := useCase.Execute(ctx, request)

	assert.NoError(t, err)
	assert.NotNil(t, supplier)
	mockRepo.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}
