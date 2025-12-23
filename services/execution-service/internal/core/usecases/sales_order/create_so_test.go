package sales_order_test

import (
	"context"
	"testing"
	"time"

	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/providers"
	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/usecases/sales_order"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockSORepository struct {
	mock.Mock
}

func (m *MockSORepository) Create(ctx context.Context, so *domain.SalesOrder) error {
	args := m.Called(ctx, so)
	return args.Error(0)
}

func (m *MockSORepository) GetByID(ctx context.Context, id, organizationID uuid.UUID) (*domain.SalesOrder, error) {
	args := m.Called(ctx, id, organizationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SalesOrder), args.Error(1)
}

func (m *MockSORepository) GetBySONumber(ctx context.Context, soNumber string, organizationID uuid.UUID) (*domain.SalesOrder, error) {
	args := m.Called(ctx, soNumber, organizationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SalesOrder), args.Error(1)
}

func (m *MockSORepository) Update(ctx context.Context, so *domain.SalesOrder) error {
	args := m.Called(ctx, so)
	return args.Error(0)
}

func (m *MockSORepository) Delete(ctx context.Context, id, organizationID uuid.UUID) error {
	args := m.Called(ctx, id, organizationID)
	return args.Error(0)
}

func (m *MockSORepository) List(ctx context.Context, organizationID uuid.UUID, filters map[string]interface{}, page, pageSize int) ([]*domain.SalesOrder, int64, error) {
	args := m.Called(ctx, organizationID, filters, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*domain.SalesOrder), args.Get(1).(int64), args.Error(2)
}

func (m *MockSORepository) GetQualifiedDemand(ctx context.Context, organizationID, productID uuid.UUID) (float64, error) {
	args := m.Called(ctx, organizationID, productID)
	return args.Get(0).(float64), args.Error(1)
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

func (m *MockCatalogClient) GetSupplier(ctx context.Context, supplierID uuid.UUID) (*providers.Supplier, error) {
	args := m.Called(ctx, supplierID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*providers.Supplier), args.Error(1)
}

func (m *MockCatalogClient) GetProductsByIDs(ctx context.Context, productIDs []uuid.UUID) ([]*providers.Product, error) {
	args := m.Called(ctx, productIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*providers.Product), args.Error(1)
}

type MockEventPublisher struct {
	mock.Mock
}

func (m *MockEventPublisher) PublishPOCreated(ctx context.Context, po *domain.PurchaseOrder) error {
	args := m.Called(ctx, po)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishPOUpdated(ctx context.Context, po *domain.PurchaseOrder) error {
	args := m.Called(ctx, po)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishPOReceived(ctx context.Context, po *domain.PurchaseOrder) error {
	args := m.Called(ctx, po)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishPOCancelled(ctx context.Context, po *domain.PurchaseOrder) error {
	args := m.Called(ctx, po)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishSOCreated(ctx context.Context, so *domain.SalesOrder) error {
	args := m.Called(ctx, so)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishSOUpdated(ctx context.Context, so *domain.SalesOrder) error {
	args := m.Called(ctx, so)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishSOCancelled(ctx context.Context, so *domain.SalesOrder) error {
	args := m.Called(ctx, so)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishDeliveryNoteIssued(ctx context.Context, so *domain.SalesOrder) error {
	args := m.Called(ctx, so)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishInventoryUpdated(ctx context.Context, txn *domain.InventoryTransaction) error {
	args := m.Called(ctx, txn)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishAlertCreated(ctx context.Context, alert *domain.Alert) error {
	args := m.Called(ctx, alert)
	return args.Error(0)
}

func TestCreateSOUseCase_Execute_WithValidData_CreatesSalesOrder(t *testing.T) {
	givenOrgID := uuid.New()
	givenCustomerID := uuid.New()
	givenProductID := uuid.New()
	givenSONumber := "SO-001"
	givenOrderDate := time.Now()
	givenDueDate := time.Now().AddDate(0, 0, 7)
	givenLineItems := []domain.SOLineItem{
		{
			ID:        uuid.New(),
			ProductID: givenProductID,
			Quantity:  10,
			UnitPrice: 150,
			LineTotal: 1500,
		},
	}

	givenProduct := &providers.Product{
		ID:   givenProductID,
		Name: "Test Product",
	}

	mockSORepo := new(MockSORepository)
	mockCatalogClient := new(MockCatalogClient)
	mockPublisher := new(MockEventPublisher)

	mockSORepo.On("GetBySONumber", mock.Anything, givenSONumber, givenOrgID).Return((*domain.SalesOrder)(nil), assert.AnError)
	mockCatalogClient.On("GetProduct", mock.Anything, givenProductID).Return(givenProduct, nil)
	mockSORepo.On("Create", mock.Anything, mock.MatchedBy(func(so *domain.SalesOrder) bool {
		return so.SONumber == givenSONumber && so.OrganizationID == givenOrgID
	})).Return(nil)
	mockPublisher.On("PublishSOCreated", mock.Anything, mock.Anything).Return(nil)

	useCase := sales_order.NewCreateSOUseCase(mockSORepo, mockCatalogClient, mockPublisher)

	input := &sales_order.CreateSOInput{
		OrganizationID: givenOrgID,
		SONumber:       givenSONumber,
		CustomerID:     givenCustomerID,
		OrderDate:      givenOrderDate,
		DueDate:        givenDueDate,
		LineItems:      givenLineItems,
	}

	so, err := useCase.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, so)
	assert.Equal(t, givenSONumber, so.SONumber)
	assert.Equal(t, givenOrgID, so.OrganizationID)
	assert.Equal(t, domain.SOStatusPending, so.Status)
	mockSORepo.AssertExpectations(t)
	mockCatalogClient.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}

func TestCreateSOUseCase_Execute_WithNilInput_ReturnsError(t *testing.T) {
	mockSORepo := new(MockSORepository)
	mockCatalogClient := new(MockCatalogClient)
	mockPublisher := new(MockEventPublisher)

	useCase := sales_order.NewCreateSOUseCase(mockSORepo, mockCatalogClient, mockPublisher)

	so, err := useCase.Execute(context.Background(), nil)

	assert.Error(t, err)
	assert.Nil(t, so)
	assert.Contains(t, err.Error(), "input cannot be nil")
}

func TestCreateSOUseCase_Execute_WithDuplicateSONumber_ReturnsError(t *testing.T) {
	givenOrgID := uuid.New()
	givenSONumber := "SO-001"
	givenExistingSO := &domain.SalesOrder{
		ID:       uuid.New(),
		SONumber: givenSONumber,
	}

	mockSORepo := new(MockSORepository)
	mockCatalogClient := new(MockCatalogClient)
	mockPublisher := new(MockEventPublisher)

	mockSORepo.On("GetBySONumber", mock.Anything, givenSONumber, givenOrgID).Return(givenExistingSO, nil)

	useCase := sales_order.NewCreateSOUseCase(mockSORepo, mockCatalogClient, mockPublisher)

	input := &sales_order.CreateSOInput{
		OrganizationID: givenOrgID,
		SONumber:       givenSONumber,
		CustomerID:     uuid.New(),
		OrderDate:      time.Now(),
		DueDate:        time.Now().AddDate(0, 0, 7),
		LineItems: []domain.SOLineItem{
			{ProductID: uuid.New(), Quantity: 10, UnitPrice: 150, LineTotal: 1500},
		},
	}

	so, err := useCase.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, so)
	assert.Contains(t, err.Error(), "already exists")
	mockSORepo.AssertExpectations(t)
}

func TestCreateSOUseCase_Execute_WithInvalidProduct_ReturnsError(t *testing.T) {
	givenOrgID := uuid.New()
	givenSONumber := "SO-001"
	givenProductID := uuid.New()

	mockSORepo := new(MockSORepository)
	mockCatalogClient := new(MockCatalogClient)
	mockPublisher := new(MockEventPublisher)

	mockSORepo.On("GetBySONumber", mock.Anything, givenSONumber, givenOrgID).Return((*domain.SalesOrder)(nil), assert.AnError)
	mockCatalogClient.On("GetProduct", mock.Anything, givenProductID).Return((*providers.Product)(nil), assert.AnError)

	useCase := sales_order.NewCreateSOUseCase(mockSORepo, mockCatalogClient, mockPublisher)

	input := &sales_order.CreateSOInput{
		OrganizationID: givenOrgID,
		SONumber:       givenSONumber,
		CustomerID:     uuid.New(),
		OrderDate:      time.Now(),
		DueDate:        time.Now().AddDate(0, 0, 7),
		LineItems: []domain.SOLineItem{
			{ProductID: givenProductID, Quantity: 10, UnitPrice: 150, LineTotal: 1500},
		},
	}

	so, err := useCase.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, so)
	assert.Contains(t, err.Error(), "invalid product_id")
	mockSORepo.AssertExpectations(t)
	mockCatalogClient.AssertExpectations(t)
}
func TestCreateSOUseCase_Execute_WhenRepositoryCreateFails_ReturnsError(t *testing.T) {
	givenOrgID := uuid.New()
	givenSONumber := "SO-001"
	givenProductID := uuid.New()
	givenProduct := &providers.Product{ID: givenProductID, Name: "Test Product"}

	mockSORepo := new(MockSORepository)
	mockCatalogClient := new(MockCatalogClient)
	mockPublisher := new(MockEventPublisher)

	mockSORepo.On("GetBySONumber", mock.Anything, givenSONumber, givenOrgID).Return((*domain.SalesOrder)(nil), assert.AnError)
	mockCatalogClient.On("GetProduct", mock.Anything, givenProductID).Return(givenProduct, nil)
	mockSORepo.On("Create", mock.Anything, mock.Anything).Return(assert.AnError)

	useCase := sales_order.NewCreateSOUseCase(mockSORepo, mockCatalogClient, mockPublisher)

	input := &sales_order.CreateSOInput{
		OrganizationID: givenOrgID,
		SONumber:       givenSONumber,
		CustomerID:     uuid.New(),
		OrderDate:      time.Now(),
		DueDate:        time.Now().AddDate(0, 0, 7),
		LineItems: []domain.SOLineItem{
			{ProductID: givenProductID, Quantity: 10, UnitPrice: 150, LineTotal: 1500},
		},
	}

	so, err := useCase.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, so)
	mockSORepo.AssertExpectations(t)
	mockCatalogClient.AssertExpectations(t)
}
