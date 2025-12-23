package purchase_order_test

import (
	"context"
	"testing"

	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/providers"
	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/usecases/purchase_order"
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

func TestReceivePOUseCase_Execute_WithValidData_ReceivesPurchaseOrder(t *testing.T) {
	givenOrgID := uuid.New()
	givenPOID := uuid.New()
	givenProductID := uuid.New()
	givenLocationID := uuid.New()
	givenReceivedBy := uuid.New()
	givenLineItemID := uuid.New()
	givenReceivedQty := 5.0

	givenPO := &domain.PurchaseOrder{
		ID:             givenPOID,
		OrganizationID: givenOrgID,
		PONumber:       "PO-001",
		Status:         domain.POStatusConfirmed,
		LineItems: []domain.POLineItem{
			{
				ID:          givenLineItemID,
				ProductID:   givenProductID,
				Quantity:    10,
				ReceivedQty: 0,
				UnitCost:    100,
				LineTotal:   1000,
			},
		},
	}

	mockPORepo := new(MockPORepository)
	mockInventoryTxnRepo := new(MockInventoryTxnRepository)
	mockInventoryBalanceRepo := new(MockInventoryBalanceRepository)
	mockDDMRPClient := new(MockDDMRPClient)
	mockPublisher := new(MockEventPublisher)

	mockPORepo.On("GetByID", mock.Anything, givenPOID, givenOrgID).Return(givenPO, nil)
	mockInventoryTxnRepo.On("Create", mock.Anything, mock.MatchedBy(func(txn *domain.InventoryTransaction) bool {
		return txn.ProductID == givenProductID && txn.Quantity == givenReceivedQty
	})).Return(nil)
	mockInventoryBalanceRepo.On("UpdateOnHand", mock.Anything, givenOrgID, givenProductID, givenLocationID, givenReceivedQty).Return(nil)
	mockDDMRPClient.On("UpdateNetFlowPosition", mock.Anything, givenOrgID, givenProductID).Return(nil)
	mockPublisher.On("PublishInventoryUpdated", mock.Anything, mock.Anything).Return(nil)
	mockPORepo.On("Update", mock.Anything, mock.MatchedBy(func(po *domain.PurchaseOrder) bool {
		return po.Status == domain.POStatusPartial
	})).Return(nil)
	mockPublisher.On("PublishPOReceived", mock.Anything, mock.Anything).Return(nil)

	useCase := purchase_order.NewReceivePOUseCase(
		mockPORepo,
		mockInventoryTxnRepo,
		mockInventoryBalanceRepo,
		mockDDMRPClient,
		mockPublisher,
	)

	input := &purchase_order.ReceivePOInput{
		POID:           givenPOID,
		OrganizationID: givenOrgID,
		LocationID:     givenLocationID,
		Receipts: []purchase_order.ReceiveLineItem{
			{LineItemID: givenLineItemID, ReceivedQty: givenReceivedQty},
		},
		ReceivedBy: givenReceivedBy,
	}

	po, err := useCase.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, po)
	assert.Equal(t, domain.POStatusPartial, po.Status)
	assert.Equal(t, givenReceivedQty, po.LineItems[0].ReceivedQty)
	mockPORepo.AssertExpectations(t)
	mockInventoryTxnRepo.AssertExpectations(t)
	mockInventoryBalanceRepo.AssertExpectations(t)
	mockDDMRPClient.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}

func TestReceivePOUseCase_Execute_WithNilInput_ReturnsError(t *testing.T) {
	mockPORepo := new(MockPORepository)
	mockInventoryTxnRepo := new(MockInventoryTxnRepository)
	mockInventoryBalanceRepo := new(MockInventoryBalanceRepository)
	mockDDMRPClient := new(MockDDMRPClient)
	mockPublisher := new(MockEventPublisher)

	useCase := purchase_order.NewReceivePOUseCase(
		mockPORepo,
		mockInventoryTxnRepo,
		mockInventoryBalanceRepo,
		mockDDMRPClient,
		mockPublisher,
	)

	po, err := useCase.Execute(context.Background(), nil)

	assert.Error(t, err)
	assert.Nil(t, po)
	assert.Contains(t, err.Error(), "input cannot be nil")
}

func TestReceivePOUseCase_Execute_WithInvalidStatus_ReturnsError(t *testing.T) {
	givenOrgID := uuid.New()
	givenPOID := uuid.New()
	givenPO := &domain.PurchaseOrder{
		ID:             givenPOID,
		OrganizationID: givenOrgID,
		Status:         domain.POStatusDraft,
	}

	mockPORepo := new(MockPORepository)
	mockInventoryTxnRepo := new(MockInventoryTxnRepository)
	mockInventoryBalanceRepo := new(MockInventoryBalanceRepository)
	mockDDMRPClient := new(MockDDMRPClient)
	mockPublisher := new(MockEventPublisher)

	mockPORepo.On("GetByID", mock.Anything, givenPOID, givenOrgID).Return(givenPO, nil)

	useCase := purchase_order.NewReceivePOUseCase(
		mockPORepo,
		mockInventoryTxnRepo,
		mockInventoryBalanceRepo,
		mockDDMRPClient,
		mockPublisher,
	)

	input := &purchase_order.ReceivePOInput{
		POID:           givenPOID,
		OrganizationID: givenOrgID,
		LocationID:     uuid.New(),
		Receipts: []purchase_order.ReceiveLineItem{
			{LineItemID: uuid.New(), ReceivedQty: 5},
		},
		ReceivedBy: uuid.New(),
	}

	po, err := useCase.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, po)
	assert.Contains(t, err.Error(), "can only receive confirmed")
	mockPORepo.AssertExpectations(t)
}

func TestReceivePOUseCase_Execute_WithExceedingQuantity_ReturnsError(t *testing.T) {
	givenOrgID := uuid.New()
	givenPOID := uuid.New()
	givenLineItemID := uuid.New()
	givenPO := &domain.PurchaseOrder{
		ID:             givenPOID,
		OrganizationID: givenOrgID,
		Status:         domain.POStatusConfirmed,
		LineItems: []domain.POLineItem{
			{
				ID:          givenLineItemID,
				ProductID:   uuid.New(),
				Quantity:    10,
				ReceivedQty: 0,
			},
		},
	}

	mockPORepo := new(MockPORepository)
	mockInventoryTxnRepo := new(MockInventoryTxnRepository)
	mockInventoryBalanceRepo := new(MockInventoryBalanceRepository)
	mockDDMRPClient := new(MockDDMRPClient)
	mockPublisher := new(MockEventPublisher)

	mockPORepo.On("GetByID", mock.Anything, givenPOID, givenOrgID).Return(givenPO, nil)

	useCase := purchase_order.NewReceivePOUseCase(
		mockPORepo,
		mockInventoryTxnRepo,
		mockInventoryBalanceRepo,
		mockDDMRPClient,
		mockPublisher,
	)

	input := &purchase_order.ReceivePOInput{
		POID:           givenPOID,
		OrganizationID: givenOrgID,
		LocationID:     uuid.New(),
		Receipts: []purchase_order.ReceiveLineItem{
			{LineItemID: givenLineItemID, ReceivedQty: 15},
		},
		ReceivedBy: uuid.New(),
	}

	po, err := useCase.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, po)
	assert.Contains(t, err.Error(), "exceeds ordered quantity")
	mockPORepo.AssertExpectations(t)
}
func TestReceivePOUseCase_Execute_WithInvalidLineItemID_ReturnsError(t *testing.T) {
	givenOrgID := uuid.New()
	givenPOID := uuid.New()
	givenPO := &domain.PurchaseOrder{
		ID:             givenPOID,
		OrganizationID: givenOrgID,
		Status:         domain.POStatusConfirmed,
		LineItems: []domain.POLineItem{
			{ID: uuid.New(), ProductID: uuid.New(), Quantity: 10},
		},
	}

	mockPORepo := new(MockPORepository)
	mockInventoryTxnRepo := new(MockInventoryTxnRepository)
	mockInventoryBalanceRepo := new(MockInventoryBalanceRepository)
	mockDDMRPClient := new(MockDDMRPClient)
	mockPublisher := new(MockEventPublisher)

	mockPORepo.On("GetByID", mock.Anything, givenPOID, givenOrgID).Return(givenPO, nil)

	useCase := purchase_order.NewReceivePOUseCase(
		mockPORepo,
		mockInventoryTxnRepo,
		mockInventoryBalanceRepo,
		mockDDMRPClient,
		mockPublisher,
	)

	input := &purchase_order.ReceivePOInput{
		POID:           givenPOID,
		OrganizationID: givenOrgID,
		LocationID:     uuid.New(),
		Receipts: []purchase_order.ReceiveLineItem{
			{LineItemID: uuid.New(), ReceivedQty: 5},
		},
		ReceivedBy: uuid.New(),
	}

	po, err := useCase.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, po)
	assert.Contains(t, err.Error(), "invalid line_item_id")
	mockPORepo.AssertExpectations(t)
}

func TestReceivePOUseCase_Execute_WithZeroReceivedQty_ReturnsError(t *testing.T) {
	givenOrgID := uuid.New()
	givenPOID := uuid.New()
	givenPO := &domain.PurchaseOrder{
		ID:             givenPOID,
		OrganizationID: givenOrgID,
		Status:         domain.POStatusConfirmed,
		LineItems:      []domain.POLineItem{{ID: uuid.New(), ProductID: uuid.New(), Quantity: 10}},
	}

	mockPORepo := new(MockPORepository)
	mockInventoryTxnRepo := new(MockInventoryTxnRepository)
	mockInventoryBalanceRepo := new(MockInventoryBalanceRepository)
	mockDDMRPClient := new(MockDDMRPClient)
	mockPublisher := new(MockEventPublisher)

	mockPORepo.On("GetByID", mock.Anything, givenPOID, givenOrgID).Return(givenPO, nil)

	useCase := purchase_order.NewReceivePOUseCase(
		mockPORepo,
		mockInventoryTxnRepo,
		mockInventoryBalanceRepo,
		mockDDMRPClient,
		mockPublisher,
	)

	input := &purchase_order.ReceivePOInput{
		POID:           givenPOID,
		OrganizationID: givenOrgID,
		LocationID:     uuid.New(),
		Receipts: []purchase_order.ReceiveLineItem{
			{LineItemID: uuid.New(), ReceivedQty: 0},
		},
		ReceivedBy: uuid.New(),
	}

	po, err := useCase.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, po)
	assert.Contains(t, err.Error(), "received quantity must be greater than zero")
	mockPORepo.AssertExpectations(t)
}

func TestReceivePOUseCase_Execute_WithEmptyReceipts_ReturnsError(t *testing.T) {
	mockPORepo := new(MockPORepository)
	mockInventoryTxnRepo := new(MockInventoryTxnRepository)
	mockInventoryBalanceRepo := new(MockInventoryBalanceRepository)
	mockDDMRPClient := new(MockDDMRPClient)
	mockPublisher := new(MockEventPublisher)

	useCase := purchase_order.NewReceivePOUseCase(
		mockPORepo,
		mockInventoryTxnRepo,
		mockInventoryBalanceRepo,
		mockDDMRPClient,
		mockPublisher,
	)

	input := &purchase_order.ReceivePOInput{
		POID:           uuid.New(),
		OrganizationID: uuid.New(),
		LocationID:     uuid.New(),
		Receipts:       []purchase_order.ReceiveLineItem{},
		ReceivedBy:     uuid.New(),
	}

	po, err := useCase.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, po)
	assert.Contains(t, err.Error(), "at least one receipt is required")
}

func TestReceivePOUseCase_Execute_WithFullyReceivedPO_UpdatesStatusToReceived(t *testing.T) {
	givenOrgID := uuid.New()
	givenPOID := uuid.New()
	givenProductID := uuid.New()
	givenLocationID := uuid.New()
	givenLineItemID := uuid.New()

	givenPO := &domain.PurchaseOrder{
		ID:             givenPOID,
		OrganizationID: givenOrgID,
		Status:         domain.POStatusConfirmed,
		LineItems: []domain.POLineItem{
			{
				ID:          givenLineItemID,
				ProductID:   givenProductID,
				Quantity:    10,
				ReceivedQty: 0,
				UnitCost:    100,
			},
		},
	}

	mockPORepo := new(MockPORepository)
	mockInventoryTxnRepo := new(MockInventoryTxnRepository)
	mockInventoryBalanceRepo := new(MockInventoryBalanceRepository)
	mockDDMRPClient := new(MockDDMRPClient)
	mockPublisher := new(MockEventPublisher)

	mockPORepo.On("GetByID", mock.Anything, givenPOID, givenOrgID).Return(givenPO, nil)
	mockInventoryTxnRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
	mockInventoryBalanceRepo.On("UpdateOnHand", mock.Anything, givenOrgID, givenProductID, givenLocationID, 10.0).Return(nil)
	mockDDMRPClient.On("UpdateNetFlowPosition", mock.Anything, givenOrgID, givenProductID).Return(nil)
	mockPublisher.On("PublishInventoryUpdated", mock.Anything, mock.Anything).Return(nil)
	mockPORepo.On("Update", mock.Anything, mock.MatchedBy(func(po *domain.PurchaseOrder) bool {
		return po.Status == domain.POStatusReceived
	})).Return(nil)
	mockPublisher.On("PublishPOReceived", mock.Anything, mock.Anything).Return(nil)

	useCase := purchase_order.NewReceivePOUseCase(
		mockPORepo,
		mockInventoryTxnRepo,
		mockInventoryBalanceRepo,
		mockDDMRPClient,
		mockPublisher,
	)

	input := &purchase_order.ReceivePOInput{
		POID:           givenPOID,
		OrganizationID: givenOrgID,
		LocationID:     givenLocationID,
		Receipts: []purchase_order.ReceiveLineItem{
			{LineItemID: givenLineItemID, ReceivedQty: 10},
		},
		ReceivedBy: uuid.New(),
	}

	po, err := useCase.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, po)
	assert.Equal(t, domain.POStatusReceived, po.Status)
	mockPORepo.AssertExpectations(t)
}

func TestReceivePOUseCase_Execute_WhenRepositoryUpdateFails_ReturnsError(t *testing.T) {
	givenOrgID := uuid.New()
	givenPOID := uuid.New()
	givenLocationID := uuid.New()
	givenReceivedBy := uuid.New()
	givenProductID := uuid.New()
	givenLineItemID := uuid.New()

	givenPO := &domain.PurchaseOrder{
		ID:             givenPOID,
		OrganizationID: givenOrgID,
		PONumber:       "PO-004",
		Status:         domain.POStatusConfirmed,
		LineItems: []domain.POLineItem{
			{
				ID:          givenLineItemID,
				ProductID:   givenProductID,
				Quantity:    100,
				ReceivedQty: 0,
				UnitCost:    50,
			},
		},
	}

	mockPORepo := new(MockPORepository)
	mockInventoryTxnRepo := new(MockInventoryTxnRepository)
	mockInventoryBalanceRepo := new(MockInventoryBalanceRepository)
	mockDDMRPClient := new(MockDDMRPClient)
	mockPublisher := new(MockEventPublisher)

	mockPORepo.On("GetByID", mock.Anything, givenPOID, givenOrgID).Return(givenPO, nil)
	mockInventoryTxnRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
	mockInventoryBalanceRepo.On("UpdateOnHand", mock.Anything, givenOrgID, givenProductID, givenLocationID, 50.0).Return(nil)
	mockDDMRPClient.On("UpdateNetFlowPosition", mock.Anything, givenOrgID, givenProductID).Return(nil)
	mockPublisher.On("PublishInventoryUpdated", mock.Anything, mock.Anything).Return(nil)
	mockPORepo.On("Update", mock.Anything, mock.Anything).Return(assert.AnError)

	useCase := purchase_order.NewReceivePOUseCase(
		mockPORepo,
		mockInventoryTxnRepo,
		mockInventoryBalanceRepo,
		mockDDMRPClient,
		mockPublisher,
	)

	input := &purchase_order.ReceivePOInput{
		POID:           givenPOID,
		OrganizationID: givenOrgID,
		LocationID:     givenLocationID,
		ReceivedBy:     givenReceivedBy,
		Receipts: []purchase_order.ReceiveLineItem{
			{LineItemID: givenLineItemID, ReceivedQty: 50},
		},
	}

	po, err := useCase.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, po)
	mockPORepo.AssertExpectations(t)
}

func TestReceivePOUseCase_Execute_WithNilPOID_ReturnsError(t *testing.T) {
	mockPORepo := new(MockPORepository)
	mockInventoryTxnRepo := new(MockInventoryTxnRepository)
	mockInventoryBalanceRepo := new(MockInventoryBalanceRepository)
	mockDDMRPClient := new(MockDDMRPClient)
	mockPublisher := new(MockEventPublisher)

	useCase := purchase_order.NewReceivePOUseCase(
		mockPORepo,
		mockInventoryTxnRepo,
		mockInventoryBalanceRepo,
		mockDDMRPClient,
		mockPublisher,
	)

	input := &purchase_order.ReceivePOInput{
		POID:           uuid.Nil,
		OrganizationID: uuid.New(),
		LocationID:     uuid.New(),
		ReceivedBy:     uuid.New(),
		Receipts:       []purchase_order.ReceiveLineItem{{LineItemID: uuid.New(), ReceivedQty: 10}},
	}

	po, err := useCase.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, po)
	assert.Contains(t, err.Error(), "po_id is required")
}

func TestReceivePOUseCase_Execute_WithNilOrganizationID_ReturnsError(t *testing.T) {
	mockPORepo := new(MockPORepository)
	mockInventoryTxnRepo := new(MockInventoryTxnRepository)
	mockInventoryBalanceRepo := new(MockInventoryBalanceRepository)
	mockDDMRPClient := new(MockDDMRPClient)
	mockPublisher := new(MockEventPublisher)

	useCase := purchase_order.NewReceivePOUseCase(
		mockPORepo,
		mockInventoryTxnRepo,
		mockInventoryBalanceRepo,
		mockDDMRPClient,
		mockPublisher,
	)

	input := &purchase_order.ReceivePOInput{
		POID:           uuid.New(),
		OrganizationID: uuid.Nil,
		LocationID:     uuid.New(),
		ReceivedBy:     uuid.New(),
		Receipts:       []purchase_order.ReceiveLineItem{{LineItemID: uuid.New(), ReceivedQty: 10}},
	}

	po, err := useCase.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, po)
	assert.Contains(t, err.Error(), "organization_id is required")
}

func TestReceivePOUseCase_Execute_WithNilLocationID_ReturnsError(t *testing.T) {
	mockPORepo := new(MockPORepository)
	mockInventoryTxnRepo := new(MockInventoryTxnRepository)
	mockInventoryBalanceRepo := new(MockInventoryBalanceRepository)
	mockDDMRPClient := new(MockDDMRPClient)
	mockPublisher := new(MockEventPublisher)

	useCase := purchase_order.NewReceivePOUseCase(
		mockPORepo,
		mockInventoryTxnRepo,
		mockInventoryBalanceRepo,
		mockDDMRPClient,
		mockPublisher,
	)

	input := &purchase_order.ReceivePOInput{
		POID:           uuid.New(),
		OrganizationID: uuid.New(),
		LocationID:     uuid.Nil,
		ReceivedBy:     uuid.New(),
		Receipts:       []purchase_order.ReceiveLineItem{{LineItemID: uuid.New(), ReceivedQty: 10}},
	}

	po, err := useCase.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, po)
	assert.Contains(t, err.Error(), "location_id is required")
}

func TestReceivePOUseCase_Execute_WithNilReceivedBy_ReturnsError(t *testing.T) {
	mockPORepo := new(MockPORepository)
	mockInventoryTxnRepo := new(MockInventoryTxnRepository)
	mockInventoryBalanceRepo := new(MockInventoryBalanceRepository)
	mockDDMRPClient := new(MockDDMRPClient)
	mockPublisher := new(MockEventPublisher)

	useCase := purchase_order.NewReceivePOUseCase(
		mockPORepo,
		mockInventoryTxnRepo,
		mockInventoryBalanceRepo,
		mockDDMRPClient,
		mockPublisher,
	)

	input := &purchase_order.ReceivePOInput{
		POID:           uuid.New(),
		OrganizationID: uuid.New(),
		LocationID:     uuid.New(),
		ReceivedBy:     uuid.Nil,
		Receipts:       []purchase_order.ReceiveLineItem{{LineItemID: uuid.New(), ReceivedQty: 10}},
	}

	po, err := useCase.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, po)
	assert.Contains(t, err.Error(), "received_by is required")
}
