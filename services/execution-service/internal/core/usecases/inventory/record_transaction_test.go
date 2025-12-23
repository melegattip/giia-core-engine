package inventory_test

import (
	"context"
	"testing"

	"github.com/giia/giia-core-engine/services/execution-service/internal/core/domain"
	"github.com/giia/giia-core-engine/services/execution-service/internal/core/providers"
	"github.com/giia/giia-core-engine/services/execution-service/internal/core/usecases/inventory"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockInventoryTxnRepository struct {
	mock.Mock
}

func (m *MockInventoryTxnRepository) Create(ctx context.Context, txn *domain.InventoryTransaction) error {
	args := m.Called(ctx, txn)
	return args.Error(0)
}

func (m *MockInventoryTxnRepository) GetByID(ctx context.Context, id, organizationID uuid.UUID) (*domain.InventoryTransaction, error) {
	args := m.Called(ctx, id, organizationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.InventoryTransaction), args.Error(1)
}

func (m *MockInventoryTxnRepository) List(ctx context.Context, organizationID, productID uuid.UUID, filters map[string]interface{}, page, pageSize int) ([]*domain.InventoryTransaction, int64, error) {
	args := m.Called(ctx, organizationID, productID, filters, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*domain.InventoryTransaction), args.Get(1).(int64), args.Error(2)
}

func (m *MockInventoryTxnRepository) GetByReferenceID(ctx context.Context, referenceType string, referenceID, organizationID uuid.UUID) ([]*domain.InventoryTransaction, error) {
	args := m.Called(ctx, referenceType, referenceID, organizationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.InventoryTransaction), args.Error(1)
}

type MockInventoryBalanceRepository struct {
	mock.Mock
}

func (m *MockInventoryBalanceRepository) GetOrCreate(ctx context.Context, organizationID, productID, locationID uuid.UUID) (*domain.InventoryBalance, error) {
	args := m.Called(ctx, organizationID, productID, locationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.InventoryBalance), args.Error(1)
}

func (m *MockInventoryBalanceRepository) UpdateOnHand(ctx context.Context, organizationID, productID, locationID uuid.UUID, quantity float64) error {
	args := m.Called(ctx, organizationID, productID, locationID, quantity)
	return args.Error(0)
}

func (m *MockInventoryBalanceRepository) UpdateReserved(ctx context.Context, organizationID, productID, locationID uuid.UUID, quantity float64) error {
	args := m.Called(ctx, organizationID, productID, locationID, quantity)
	return args.Error(0)
}

func (m *MockInventoryBalanceRepository) GetByProduct(ctx context.Context, organizationID, productID uuid.UUID) ([]*domain.InventoryBalance, error) {
	args := m.Called(ctx, organizationID, productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.InventoryBalance), args.Error(1)
}

func (m *MockInventoryBalanceRepository) GetByLocation(ctx context.Context, organizationID, locationID uuid.UUID) ([]*domain.InventoryBalance, error) {
	args := m.Called(ctx, organizationID, locationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.InventoryBalance), args.Error(1)
}

type MockDDMRPClient struct {
	mock.Mock
}

func (m *MockDDMRPClient) GetBufferStatus(ctx context.Context, organizationID, productID uuid.UUID) (*providers.BufferStatus, error) {
	args := m.Called(ctx, organizationID, productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*providers.BufferStatus), args.Error(1)
}

func (m *MockDDMRPClient) UpdateNetFlowPosition(ctx context.Context, organizationID, productID uuid.UUID) error {
	args := m.Called(ctx, organizationID, productID)
	return args.Error(0)
}

func (m *MockDDMRPClient) GetProductsInRedZone(ctx context.Context, organizationID uuid.UUID) ([]*providers.BufferStatus, error) {
	args := m.Called(ctx, organizationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*providers.BufferStatus), args.Error(1)
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

func TestRecordTransactionUseCase_Execute_WithValidReceiptData_CreatesTransaction(t *testing.T) {
	givenOrgID := uuid.New()
	givenProductID := uuid.New()
	givenLocationID := uuid.New()
	givenCreatedBy := uuid.New()
	givenQuantity := 100.0
	givenUnitCost := 50.0

	mockInventoryTxnRepo := new(MockInventoryTxnRepository)
	mockInventoryBalanceRepo := new(MockInventoryBalanceRepository)
	mockDDMRPClient := new(MockDDMRPClient)
	mockPublisher := new(MockEventPublisher)

	mockInventoryTxnRepo.On("Create", mock.Anything, mock.MatchedBy(func(txn *domain.InventoryTransaction) bool {
		return txn.ProductID == givenProductID && txn.Quantity == givenQuantity
	})).Return(nil)
	mockInventoryBalanceRepo.On("UpdateOnHand", mock.Anything, givenOrgID, givenProductID, givenLocationID, givenQuantity).Return(nil)
	mockDDMRPClient.On("UpdateNetFlowPosition", mock.Anything, givenOrgID, givenProductID).Return(nil)
	mockPublisher.On("PublishInventoryUpdated", mock.Anything, mock.Anything).Return(nil)

	useCase := inventory.NewRecordTransactionUseCase(
		mockInventoryTxnRepo,
		mockInventoryBalanceRepo,
		mockDDMRPClient,
		mockPublisher,
	)

	input := &inventory.RecordTransactionInput{
		OrganizationID: givenOrgID,
		ProductID:      givenProductID,
		LocationID:     givenLocationID,
		Type:           domain.TransactionReceipt,
		Quantity:       givenQuantity,
		UnitCost:       givenUnitCost,
		ReferenceType:  "purchase_order",
		ReferenceID:    uuid.New(),
		Reason:         "PO receipt",
		CreatedBy:      givenCreatedBy,
	}

	txn, err := useCase.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, txn)
	assert.Equal(t, givenProductID, txn.ProductID)
	assert.Equal(t, givenQuantity, txn.Quantity)
	assert.Equal(t, domain.TransactionReceipt, txn.Type)
	mockInventoryTxnRepo.AssertExpectations(t)
	mockInventoryBalanceRepo.AssertExpectations(t)
	mockDDMRPClient.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}

func TestRecordTransactionUseCase_Execute_WithNilInput_ReturnsError(t *testing.T) {
	mockInventoryTxnRepo := new(MockInventoryTxnRepository)
	mockInventoryBalanceRepo := new(MockInventoryBalanceRepository)
	mockDDMRPClient := new(MockDDMRPClient)
	mockPublisher := new(MockEventPublisher)

	useCase := inventory.NewRecordTransactionUseCase(
		mockInventoryTxnRepo,
		mockInventoryBalanceRepo,
		mockDDMRPClient,
		mockPublisher,
	)

	txn, err := useCase.Execute(context.Background(), nil)

	assert.Error(t, err)
	assert.Nil(t, txn)
	assert.Contains(t, err.Error(), "input cannot be nil")
}

func TestRecordTransactionUseCase_Execute_WithIssueType_CreatesNegativeTransaction(t *testing.T) {
	givenOrgID := uuid.New()
	givenProductID := uuid.New()
	givenLocationID := uuid.New()
	givenCreatedBy := uuid.New()
	givenQuantity := -50.0

	mockInventoryTxnRepo := new(MockInventoryTxnRepository)
	mockInventoryBalanceRepo := new(MockInventoryBalanceRepository)
	mockDDMRPClient := new(MockDDMRPClient)
	mockPublisher := new(MockEventPublisher)

	mockInventoryTxnRepo.On("Create", mock.Anything, mock.MatchedBy(func(txn *domain.InventoryTransaction) bool {
		return txn.ProductID == givenProductID && txn.Quantity == givenQuantity
	})).Return(nil)
	mockInventoryBalanceRepo.On("UpdateOnHand", mock.Anything, givenOrgID, givenProductID, givenLocationID, givenQuantity).Return(nil)
	mockDDMRPClient.On("UpdateNetFlowPosition", mock.Anything, givenOrgID, givenProductID).Return(nil)
	mockPublisher.On("PublishInventoryUpdated", mock.Anything, mock.Anything).Return(nil)

	useCase := inventory.NewRecordTransactionUseCase(
		mockInventoryTxnRepo,
		mockInventoryBalanceRepo,
		mockDDMRPClient,
		mockPublisher,
	)

	input := &inventory.RecordTransactionInput{
		OrganizationID: givenOrgID,
		ProductID:      givenProductID,
		LocationID:     givenLocationID,
		Type:           domain.TransactionIssue,
		Quantity:       givenQuantity,
		UnitCost:       0,
		ReferenceType:  "sales_order",
		ReferenceID:    uuid.New(),
		Reason:         "SO fulfillment",
		CreatedBy:      givenCreatedBy,
	}

	txn, err := useCase.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, txn)
	assert.Equal(t, givenQuantity, txn.Quantity)
	assert.Equal(t, domain.TransactionIssue, txn.Type)
	mockInventoryTxnRepo.AssertExpectations(t)
	mockInventoryBalanceRepo.AssertExpectations(t)
	mockDDMRPClient.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}
func TestRecordTransactionUseCase_Execute_WhenTransactionCreationFails_ReturnsError(t *testing.T) {
	mockInventoryTxnRepo := new(MockInventoryTxnRepository)
	mockInventoryBalanceRepo := new(MockInventoryBalanceRepository)
	mockDDMRPClient := new(MockDDMRPClient)
	mockPublisher := new(MockEventPublisher)

	mockInventoryTxnRepo.On("Create", mock.Anything, mock.Anything).Return(assert.AnError)

	useCase := inventory.NewRecordTransactionUseCase(
		mockInventoryTxnRepo,
		mockInventoryBalanceRepo,
		mockDDMRPClient,
		mockPublisher,
	)

	input := &inventory.RecordTransactionInput{
		OrganizationID: uuid.New(),
		ProductID:      uuid.New(),
		LocationID:     uuid.New(),
		Type:           domain.TransactionReceipt,
		Quantity:       100.0,
		UnitCost:       50.0,
		ReferenceType:  "purchase_order",
		ReferenceID:    uuid.New(),
		Reason:         "PO receipt",
		CreatedBy:      uuid.New(),
	}

	txn, err := useCase.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, txn)
	mockInventoryTxnRepo.AssertExpectations(t)
}

func TestRecordTransactionUseCase_Execute_WhenBalanceUpdateFails_ReturnsError(t *testing.T) {
	givenOrgID := uuid.New()
	givenProductID := uuid.New()
	givenLocationID := uuid.New()

	mockInventoryTxnRepo := new(MockInventoryTxnRepository)
	mockInventoryBalanceRepo := new(MockInventoryBalanceRepository)
	mockDDMRPClient := new(MockDDMRPClient)
	mockPublisher := new(MockEventPublisher)

	mockInventoryTxnRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
	mockInventoryBalanceRepo.On("UpdateOnHand", mock.Anything, givenOrgID, givenProductID, givenLocationID, 100.0).Return(assert.AnError)

	useCase := inventory.NewRecordTransactionUseCase(
		mockInventoryTxnRepo,
		mockInventoryBalanceRepo,
		mockDDMRPClient,
		mockPublisher,
	)

	input := &inventory.RecordTransactionInput{
		OrganizationID: givenOrgID,
		ProductID:      givenProductID,
		LocationID:     givenLocationID,
		Type:           domain.TransactionReceipt,
		Quantity:       100.0,
		UnitCost:       50.0,
		ReferenceType:  "purchase_order",
		ReferenceID:    uuid.New(),
		Reason:         "PO receipt",
		CreatedBy:      uuid.New(),
	}

	txn, err := useCase.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, txn)
	mockInventoryTxnRepo.AssertExpectations(t)
	mockInventoryBalanceRepo.AssertExpectations(t)
}
