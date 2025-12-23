package product

import (
	"context"
	"testing"

	"github.com/giia/giia-core-engine/pkg/logger"
	"github.com/giia/giia-core-engine/services/catalog-service/internal/core/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockProductRepository struct {
	mock.Mock
}

func (m *MockProductRepository) Create(ctx context.Context, product *domain.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockProductRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Product), args.Error(1)
}

func (m *MockProductRepository) GetByIDWithSuppliers(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Product), args.Error(1)
}

func (m *MockProductRepository) GetBySKU(ctx context.Context, sku string) (*domain.Product, error) {
	args := m.Called(ctx, sku)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Product), args.Error(1)
}

func (m *MockProductRepository) Update(ctx context.Context, product *domain.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProductRepository) List(ctx context.Context, filters map[string]interface{}, page, pageSize int) ([]*domain.Product, int64, error) {
	args := m.Called(ctx, filters, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*domain.Product), args.Get(1).(int64), args.Error(2)
}

func (m *MockProductRepository) Search(ctx context.Context, query string, filters map[string]interface{}, page, pageSize int) ([]*domain.Product, int64, error) {
	args := m.Called(ctx, query, filters, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*domain.Product), args.Get(1).(int64), args.Error(2)
}

func (m *MockProductRepository) AssociateSupplier(ctx context.Context, productSupplier *domain.ProductSupplier) error {
	args := m.Called(ctx, productSupplier)
	return args.Error(0)
}

func (m *MockProductRepository) RemoveSupplier(ctx context.Context, productID, supplierID uuid.UUID) error {
	args := m.Called(ctx, productID, supplierID)
	return args.Error(0)
}

func (m *MockProductRepository) GetProductSuppliers(ctx context.Context, productID uuid.UUID) ([]*domain.ProductSupplier, error) {
	args := m.Called(ctx, productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.ProductSupplier), args.Error(1)
}

type MockEventPublisher struct {
	mock.Mock
}

func (m *MockEventPublisher) PublishProductCreated(ctx context.Context, product *domain.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishProductUpdated(ctx context.Context, product *domain.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishProductDeleted(ctx context.Context, product *domain.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishSupplierCreated(ctx context.Context, supplier *domain.Supplier) error {
	args := m.Called(ctx, supplier)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishSupplierUpdated(ctx context.Context, supplier *domain.Supplier) error {
	args := m.Called(ctx, supplier)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishSupplierDeleted(ctx context.Context, supplier *domain.Supplier) error {
	args := m.Called(ctx, supplier)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishBufferProfileCreated(ctx context.Context, profile *domain.BufferProfile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishBufferProfileUpdated(ctx context.Context, profile *domain.BufferProfile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishBufferProfileDeleted(ctx context.Context, profile *domain.BufferProfile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishBufferProfileAssigned(ctx context.Context, product *domain.Product, profile *domain.BufferProfile) error {
	args := m.Called(ctx, product, profile)
	return args.Error(0)
}

func TestCreateProductUseCase_Execute_WithValidData_ReturnsProduct(t *testing.T) {
	givenSKU := "WIDGET-001"
	givenName := "Premium Widget"
	givenUnitOfMeasure := "EA"
	givenOrgID := uuid.New()

	mockRepo := new(MockProductRepository)
	mockPublisher := new(MockEventPublisher)
	mockLogger := logger.New("test", "debug")

	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(p *domain.Product) bool {
		return p.SKU == givenSKU && p.Name == givenName
	})).Return(nil)

	mockPublisher.On("PublishProductCreated", mock.Anything, mock.Anything).Return(nil)

	useCase := NewCreateProductUseCase(mockRepo, mockPublisher, mockLogger)

	ctx := context.WithValue(context.Background(), "organization_id", givenOrgID)

	request := &CreateProductRequest{
		SKU:           givenSKU,
		Name:          givenName,
		UnitOfMeasure: givenUnitOfMeasure,
	}

	product, err := useCase.Execute(ctx, request)

	assert.NoError(t, err)
	assert.NotNil(t, product)
	assert.Equal(t, givenSKU, product.SKU)
	assert.Equal(t, givenName, product.Name)
	assert.Equal(t, givenUnitOfMeasure, product.UnitOfMeasure)
	assert.Equal(t, domain.ProductStatusActive, product.Status)

	mockRepo.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}

func TestCreateProductUseCase_Execute_WithNilRequest_ReturnsError(t *testing.T) {
	mockRepo := new(MockProductRepository)
	mockPublisher := new(MockEventPublisher)
	mockLogger := logger.New("test", "debug")

	useCase := NewCreateProductUseCase(mockRepo, mockPublisher, mockLogger)

	product, err := useCase.Execute(context.Background(), nil)

	assert.Error(t, err)
	assert.Nil(t, product)
	assert.Contains(t, err.Error(), "request cannot be nil")
}

func TestCreateProductUseCase_Execute_WithMissingSKU_ReturnsError(t *testing.T) {
	givenOrgID := uuid.New()

	mockRepo := new(MockProductRepository)
	mockPublisher := new(MockEventPublisher)
	mockLogger := logger.New("test", "debug")

	useCase := NewCreateProductUseCase(mockRepo, mockPublisher, mockLogger)

	ctx := context.WithValue(context.Background(), "organization_id", givenOrgID)

	request := &CreateProductRequest{
		Name:          "Product Name",
		UnitOfMeasure: "EA",
	}

	product, err := useCase.Execute(ctx, request)

	assert.Error(t, err)
	assert.Nil(t, product)
	assert.Contains(t, err.Error(), "SKU is required")
}
