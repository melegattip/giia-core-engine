package purchase_order_test

import (
	"context"
	"testing"
	"time"

	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/providers"
	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/usecases/purchase_order"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockPORepository struct {
	mock.Mock
}

func (m *MockPORepository) Create(ctx context.Context, po *domain.PurchaseOrder) error {
	args := m.Called(ctx, po)
	return args.Error(0)
}

func (m *MockPORepository) GetByID(ctx context.Context, id, organizationID uuid.UUID) (*domain.PurchaseOrder, error) {
	args := m.Called(ctx, id, organizationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PurchaseOrder), args.Error(1)
}

func (m *MockPORepository) GetByPONumber(ctx context.Context, poNumber string, organizationID uuid.UUID) (*domain.PurchaseOrder, error) {
	args := m.Called(ctx, poNumber, organizationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PurchaseOrder), args.Error(1)
}

func (m *MockPORepository) Update(ctx context.Context, po *domain.PurchaseOrder) error {
	args := m.Called(ctx, po)
	return args.Error(0)
}

func (m *MockPORepository) Delete(ctx context.Context, id, organizationID uuid.UUID) error {
	args := m.Called(ctx, id, organizationID)
	return args.Error(0)
}

func (m *MockPORepository) List(ctx context.Context, organizationID uuid.UUID, filters map[string]interface{}, page, pageSize int) ([]*domain.PurchaseOrder, int64, error) {
	args := m.Called(ctx, organizationID, filters, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*domain.PurchaseOrder), args.Get(1).(int64), args.Error(2)
}

func (m *MockPORepository) GetDelayedOrders(ctx context.Context, organizationID uuid.UUID) ([]*domain.PurchaseOrder, error) {
	args := m.Called(ctx, organizationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.PurchaseOrder), args.Error(1)
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

func TestCreatePOUseCase_Execute_WithValidData_CreatesPurchaseOrder(t *testing.T) {
	givenOrgID := uuid.New()
	givenSupplierID := uuid.New()
	givenProductID := uuid.New()
	givenCreatedBy := uuid.New()
	givenPONumber := "PO-001"
	givenOrderDate := time.Now()
	givenExpectedArrivalDate := time.Now().AddDate(0, 0, 7)
	givenLineItems := []domain.POLineItem{
		{
			ID:        uuid.New(),
			ProductID: givenProductID,
			Quantity:  10,
			UnitCost:  100,
			LineTotal: 1000,
		},
	}

	givenSupplier := &providers.Supplier{
		ID:   givenSupplierID,
		Name: "Test Supplier",
	}
	givenProduct := &providers.Product{
		ID:   givenProductID,
		Name: "Test Product",
	}

	mockPORepo := new(MockPORepository)
	mockCatalogClient := new(MockCatalogClient)
	mockPublisher := new(MockEventPublisher)

	mockPORepo.On("GetByPONumber", mock.Anything, givenPONumber, givenOrgID).Return((*domain.PurchaseOrder)(nil), assert.AnError)
	mockCatalogClient.On("GetSupplier", mock.Anything, givenSupplierID).Return(givenSupplier, nil)
	mockCatalogClient.On("GetProduct", mock.Anything, givenProductID).Return(givenProduct, nil)
	mockPORepo.On("Create", mock.Anything, mock.MatchedBy(func(po *domain.PurchaseOrder) bool {
		return po.PONumber == givenPONumber && po.OrganizationID == givenOrgID
	})).Return(nil)
	mockPublisher.On("PublishPOCreated", mock.Anything, mock.Anything).Return(nil)

	useCase := purchase_order.NewCreatePOUseCase(mockPORepo, mockCatalogClient, mockPublisher)

	input := &purchase_order.CreatePOInput{
		OrganizationID:      givenOrgID,
		PONumber:            givenPONumber,
		SupplierID:          givenSupplierID,
		OrderDate:           givenOrderDate,
		ExpectedArrivalDate: givenExpectedArrivalDate,
		LineItems:           givenLineItems,
		CreatedBy:           givenCreatedBy,
	}

	po, err := useCase.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, po)
	assert.Equal(t, givenPONumber, po.PONumber)
	assert.Equal(t, givenOrgID, po.OrganizationID)
	assert.Equal(t, givenSupplierID, po.SupplierID)
	assert.Equal(t, domain.POStatusDraft, po.Status)
	mockPORepo.AssertExpectations(t)
	mockCatalogClient.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}

func TestCreatePOUseCase_Execute_WithNilInput_ReturnsError(t *testing.T) {
	mockPORepo := new(MockPORepository)
	mockCatalogClient := new(MockCatalogClient)
	mockPublisher := new(MockEventPublisher)

	useCase := purchase_order.NewCreatePOUseCase(mockPORepo, mockCatalogClient, mockPublisher)

	po, err := useCase.Execute(context.Background(), nil)

	assert.Error(t, err)
	assert.Nil(t, po)
	assert.Contains(t, err.Error(), "input cannot be nil")
}

func TestCreatePOUseCase_Execute_WithDuplicatePONumber_ReturnsError(t *testing.T) {
	givenOrgID := uuid.New()
	givenPONumber := "PO-001"
	givenExistingPO := &domain.PurchaseOrder{
		ID:       uuid.New(),
		PONumber: givenPONumber,
	}

	mockPORepo := new(MockPORepository)
	mockCatalogClient := new(MockCatalogClient)
	mockPublisher := new(MockEventPublisher)

	mockPORepo.On("GetByPONumber", mock.Anything, givenPONumber, givenOrgID).Return(givenExistingPO, nil)

	useCase := purchase_order.NewCreatePOUseCase(mockPORepo, mockCatalogClient, mockPublisher)

	input := &purchase_order.CreatePOInput{
		OrganizationID:      givenOrgID,
		PONumber:            givenPONumber,
		SupplierID:          uuid.New(),
		OrderDate:           time.Now(),
		ExpectedArrivalDate: time.Now().AddDate(0, 0, 7),
		LineItems: []domain.POLineItem{
			{ProductID: uuid.New(), Quantity: 10, UnitCost: 100, LineTotal: 1000},
		},
		CreatedBy: uuid.New(),
	}

	po, err := useCase.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, po)
	assert.Contains(t, err.Error(), "already exists")
	mockPORepo.AssertExpectations(t)
}

func TestCreatePOUseCase_Execute_WithInvalidSupplier_ReturnsError(t *testing.T) {
	givenOrgID := uuid.New()
	givenSupplierID := uuid.New()
	givenPONumber := "PO-001"

	mockPORepo := new(MockPORepository)
	mockCatalogClient := new(MockCatalogClient)
	mockPublisher := new(MockEventPublisher)

	mockPORepo.On("GetByPONumber", mock.Anything, givenPONumber, givenOrgID).Return((*domain.PurchaseOrder)(nil), assert.AnError)
	mockCatalogClient.On("GetSupplier", mock.Anything, givenSupplierID).Return((*providers.Supplier)(nil), assert.AnError)

	useCase := purchase_order.NewCreatePOUseCase(mockPORepo, mockCatalogClient, mockPublisher)

	input := &purchase_order.CreatePOInput{
		OrganizationID:      givenOrgID,
		PONumber:            givenPONumber,
		SupplierID:          givenSupplierID,
		OrderDate:           time.Now(),
		ExpectedArrivalDate: time.Now().AddDate(0, 0, 7),
		LineItems: []domain.POLineItem{
			{ProductID: uuid.New(), Quantity: 10, UnitCost: 100, LineTotal: 1000},
		},
		CreatedBy: uuid.New(),
	}

	po, err := useCase.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, po)
	assert.Contains(t, err.Error(), "invalid supplier_id")
	mockPORepo.AssertExpectations(t)
	mockCatalogClient.AssertExpectations(t)
}

func TestCreatePOUseCase_Execute_WithInvalidProduct_ReturnsError(t *testing.T) {
	givenOrgID := uuid.New()
	givenSupplierID := uuid.New()
	givenProductID := uuid.New()
	givenPONumber := "PO-001"
	givenSupplier := &providers.Supplier{ID: givenSupplierID, Name: "Test Supplier"}

	mockPORepo := new(MockPORepository)
	mockCatalogClient := new(MockCatalogClient)
	mockPublisher := new(MockEventPublisher)

	mockPORepo.On("GetByPONumber", mock.Anything, givenPONumber, givenOrgID).Return((*domain.PurchaseOrder)(nil), assert.AnError)
	mockCatalogClient.On("GetSupplier", mock.Anything, givenSupplierID).Return(givenSupplier, nil)
	mockCatalogClient.On("GetProduct", mock.Anything, givenProductID).Return((*providers.Product)(nil), assert.AnError)

	useCase := purchase_order.NewCreatePOUseCase(mockPORepo, mockCatalogClient, mockPublisher)

	input := &purchase_order.CreatePOInput{
		OrganizationID:      givenOrgID,
		PONumber:            givenPONumber,
		SupplierID:          givenSupplierID,
		OrderDate:           time.Now(),
		ExpectedArrivalDate: time.Now().AddDate(0, 0, 7),
		LineItems: []domain.POLineItem{
			{ProductID: givenProductID, Quantity: 10, UnitCost: 100, LineTotal: 1000},
		},
		CreatedBy: uuid.New(),
	}

	po, err := useCase.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, po)
	assert.Contains(t, err.Error(), "invalid product_id")
	mockPORepo.AssertExpectations(t)
	mockCatalogClient.AssertExpectations(t)
}

func TestCreatePOUseCase_Execute_WhenRepositoryCreateFails_ReturnsError(t *testing.T) {
	givenOrgID := uuid.New()
	givenSupplierID := uuid.New()
	givenProductID := uuid.New()
	givenPONumber := "PO-001"
	givenSupplier := &providers.Supplier{ID: givenSupplierID, Name: "Test Supplier"}
	givenProduct := &providers.Product{ID: givenProductID, Name: "Test Product"}

	mockPORepo := new(MockPORepository)
	mockCatalogClient := new(MockCatalogClient)
	mockPublisher := new(MockEventPublisher)

	mockPORepo.On("GetByPONumber", mock.Anything, givenPONumber, givenOrgID).Return((*domain.PurchaseOrder)(nil), assert.AnError)
	mockCatalogClient.On("GetSupplier", mock.Anything, givenSupplierID).Return(givenSupplier, nil)
	mockCatalogClient.On("GetProduct", mock.Anything, givenProductID).Return(givenProduct, nil)
	mockPORepo.On("Create", mock.Anything, mock.Anything).Return(assert.AnError)

	useCase := purchase_order.NewCreatePOUseCase(mockPORepo, mockCatalogClient, mockPublisher)

	input := &purchase_order.CreatePOInput{
		OrganizationID:      givenOrgID,
		PONumber:            givenPONumber,
		SupplierID:          givenSupplierID,
		OrderDate:           time.Now(),
		ExpectedArrivalDate: time.Now().AddDate(0, 0, 7),
		LineItems: []domain.POLineItem{
			{ProductID: givenProductID, Quantity: 10, UnitCost: 100, LineTotal: 1000},
		},
		CreatedBy: uuid.New(),
	}

	po, err := useCase.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, po)
	mockPORepo.AssertExpectations(t)
	mockCatalogClient.AssertExpectations(t)
}
func TestCreatePOUseCase_Execute_WithMultipleLineItems_CreatesSuccessfully(t *testing.T) {
	givenOrgID := uuid.New()
	givenSupplierID := uuid.New()
	givenProduct1ID := uuid.New()
	givenProduct2ID := uuid.New()
	givenCreatedBy := uuid.New()
	givenPONumber := "PO-MULTI-001"

	givenSupplier := &providers.Supplier{ID: givenSupplierID, Name: "Test Supplier"}
	givenProduct1 := &providers.Product{ID: givenProduct1ID, Name: "Product 1"}
	givenProduct2 := &providers.Product{ID: givenProduct2ID, Name: "Product 2"}

	mockPORepo := new(MockPORepository)
	mockCatalogClient := new(MockCatalogClient)
	mockPublisher := new(MockEventPublisher)

	mockPORepo.On("GetByPONumber", mock.Anything, givenPONumber, givenOrgID).Return((*domain.PurchaseOrder)(nil), assert.AnError)
	mockCatalogClient.On("GetSupplier", mock.Anything, givenSupplierID).Return(givenSupplier, nil)
	mockCatalogClient.On("GetProduct", mock.Anything, givenProduct1ID).Return(givenProduct1, nil)
	mockCatalogClient.On("GetProduct", mock.Anything, givenProduct2ID).Return(givenProduct2, nil)
	mockPORepo.On("Create", mock.Anything, mock.Anything).Return(nil)
	mockPublisher.On("PublishPOCreated", mock.Anything, mock.Anything).Return(nil)

	useCase := purchase_order.NewCreatePOUseCase(mockPORepo, mockCatalogClient, mockPublisher)

	input := &purchase_order.CreatePOInput{
		OrganizationID:      givenOrgID,
		PONumber:            givenPONumber,
		SupplierID:          givenSupplierID,
		OrderDate:           time.Now(),
		ExpectedArrivalDate: time.Now().AddDate(0, 0, 7),
		LineItems: []domain.POLineItem{
			{ID: uuid.New(), ProductID: givenProduct1ID, Quantity: 10, UnitCost: 100, LineTotal: 1000},
			{ID: uuid.New(), ProductID: givenProduct2ID, Quantity: 5, UnitCost: 200, LineTotal: 1000},
		},
		CreatedBy: givenCreatedBy,
	}

	po, err := useCase.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, po)
	assert.Len(t, po.LineItems, 2)
	mockPORepo.AssertExpectations(t)
	mockCatalogClient.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}

func TestCreatePOUseCase_Execute_WhenPORepoGetByPONumberFails_StillCreatesIfNotConflict(t *testing.T) {
	givenOrgID := uuid.New()
	givenSupplierID := uuid.New()
	givenProductID := uuid.New()
	givenCreatedBy := uuid.New()
	givenPONumber := "PO-002"

	givenSupplier := &providers.Supplier{ID: givenSupplierID, Name: "Test Supplier"}
	givenProduct := &providers.Product{ID: givenProductID, Name: "Test Product"}

	mockPORepo := new(MockPORepository)
	mockCatalogClient := new(MockCatalogClient)
	mockPublisher := new(MockEventPublisher)

	mockPORepo.On("GetByPONumber", mock.Anything, givenPONumber, givenOrgID).Return((*domain.PurchaseOrder)(nil), assert.AnError)
	mockCatalogClient.On("GetSupplier", mock.Anything, givenSupplierID).Return(givenSupplier, nil)
	mockCatalogClient.On("GetProduct", mock.Anything, givenProductID).Return(givenProduct, nil)
	mockPORepo.On("Create", mock.Anything, mock.Anything).Return(nil)
	mockPublisher.On("PublishPOCreated", mock.Anything, mock.Anything).Return(nil)

	useCase := purchase_order.NewCreatePOUseCase(mockPORepo, mockCatalogClient, mockPublisher)

	input := &purchase_order.CreatePOInput{
		OrganizationID:      givenOrgID,
		PONumber:            givenPONumber,
		SupplierID:          givenSupplierID,
		OrderDate:           time.Now(),
		ExpectedArrivalDate: time.Now().AddDate(0, 0, 7),
		LineItems: []domain.POLineItem{
			{ID: uuid.New(), ProductID: givenProductID, Quantity: 10, UnitCost: 100, LineTotal: 1000},
		},
		CreatedBy: givenCreatedBy,
	}

	po, err := useCase.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, po)
	mockPORepo.AssertExpectations(t)
	mockCatalogClient.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}
