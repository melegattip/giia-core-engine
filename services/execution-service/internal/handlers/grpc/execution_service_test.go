package grpc_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/execution-service/internal/handlers/grpc"
)

// MockInventoryBalanceRepository is a mock for InventoryBalanceRepository.
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
	return args.Get(0).([]*domain.InventoryBalance), args.Error(1)
}

func (m *MockInventoryBalanceRepository) GetByLocation(ctx context.Context, organizationID, locationID uuid.UUID) ([]*domain.InventoryBalance, error) {
	args := m.Called(ctx, organizationID, locationID)
	return args.Get(0).([]*domain.InventoryBalance), args.Error(1)
}

// MockInventoryTransactionRepository is a mock for InventoryTransactionRepository.
type MockInventoryTransactionRepository struct {
	mock.Mock
}

func (m *MockInventoryTransactionRepository) Create(ctx context.Context, txn *domain.InventoryTransaction) error {
	args := m.Called(ctx, txn)
	return args.Error(0)
}

func (m *MockInventoryTransactionRepository) GetByID(ctx context.Context, id, organizationID uuid.UUID) (*domain.InventoryTransaction, error) {
	args := m.Called(ctx, id, organizationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.InventoryTransaction), args.Error(1)
}

func (m *MockInventoryTransactionRepository) List(ctx context.Context, organizationID, productID uuid.UUID, filters map[string]interface{}, page, pageSize int) ([]*domain.InventoryTransaction, int64, error) {
	args := m.Called(ctx, organizationID, productID, filters, page, pageSize)
	return args.Get(0).([]*domain.InventoryTransaction), args.Get(1).(int64), args.Error(2)
}

func (m *MockInventoryTransactionRepository) GetByReferenceID(ctx context.Context, referenceType string, referenceID, organizationID uuid.UUID) ([]*domain.InventoryTransaction, error) {
	args := m.Called(ctx, referenceType, referenceID, organizationID)
	return args.Get(0).([]*domain.InventoryTransaction), args.Error(1)
}

// MockPurchaseOrderRepository is a mock for PurchaseOrderRepository.
type MockPurchaseOrderRepository struct {
	mock.Mock
}

func (m *MockPurchaseOrderRepository) Create(ctx context.Context, po *domain.PurchaseOrder) error {
	args := m.Called(ctx, po)
	return args.Error(0)
}

func (m *MockPurchaseOrderRepository) GetByID(ctx context.Context, id, organizationID uuid.UUID) (*domain.PurchaseOrder, error) {
	args := m.Called(ctx, id, organizationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PurchaseOrder), args.Error(1)
}

func (m *MockPurchaseOrderRepository) GetByPONumber(ctx context.Context, poNumber string, organizationID uuid.UUID) (*domain.PurchaseOrder, error) {
	args := m.Called(ctx, poNumber, organizationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PurchaseOrder), args.Error(1)
}

func (m *MockPurchaseOrderRepository) Update(ctx context.Context, po *domain.PurchaseOrder) error {
	args := m.Called(ctx, po)
	return args.Error(0)
}

func (m *MockPurchaseOrderRepository) Delete(ctx context.Context, id, organizationID uuid.UUID) error {
	args := m.Called(ctx, id, organizationID)
	return args.Error(0)
}

func (m *MockPurchaseOrderRepository) List(ctx context.Context, organizationID uuid.UUID, filters map[string]interface{}, page, pageSize int) ([]*domain.PurchaseOrder, int64, error) {
	args := m.Called(ctx, organizationID, filters, page, pageSize)
	return args.Get(0).([]*domain.PurchaseOrder), args.Get(1).(int64), args.Error(2)
}

func (m *MockPurchaseOrderRepository) GetDelayedOrders(ctx context.Context, organizationID uuid.UUID) ([]*domain.PurchaseOrder, error) {
	args := m.Called(ctx, organizationID)
	return args.Get(0).([]*domain.PurchaseOrder), args.Error(1)
}

// MockSalesOrderRepository is a mock for SalesOrderRepository.
type MockSalesOrderRepository struct {
	mock.Mock
}

func (m *MockSalesOrderRepository) Create(ctx context.Context, so *domain.SalesOrder) error {
	args := m.Called(ctx, so)
	return args.Error(0)
}

func (m *MockSalesOrderRepository) GetByID(ctx context.Context, id, organizationID uuid.UUID) (*domain.SalesOrder, error) {
	args := m.Called(ctx, id, organizationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SalesOrder), args.Error(1)
}

func (m *MockSalesOrderRepository) GetBySONumber(ctx context.Context, soNumber string, organizationID uuid.UUID) (*domain.SalesOrder, error) {
	args := m.Called(ctx, soNumber, organizationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SalesOrder), args.Error(1)
}

func (m *MockSalesOrderRepository) Update(ctx context.Context, so *domain.SalesOrder) error {
	args := m.Called(ctx, so)
	return args.Error(0)
}

func (m *MockSalesOrderRepository) Delete(ctx context.Context, id, organizationID uuid.UUID) error {
	args := m.Called(ctx, id, organizationID)
	return args.Error(0)
}

func (m *MockSalesOrderRepository) List(ctx context.Context, organizationID uuid.UUID, filters map[string]interface{}, page, pageSize int) ([]*domain.SalesOrder, int64, error) {
	args := m.Called(ctx, organizationID, filters, page, pageSize)
	return args.Get(0).([]*domain.SalesOrder), args.Get(1).(int64), args.Error(2)
}

func (m *MockSalesOrderRepository) GetQualifiedDemand(ctx context.Context, organizationID, productID uuid.UUID) (float64, error) {
	args := m.Called(ctx, organizationID, productID)
	return args.Get(0).(float64), args.Error(1)
}

func TestExecutionService_GetInventoryBalance(t *testing.T) {
	mockBalanceRepo := new(MockInventoryBalanceRepository)
	mockTxnRepo := new(MockInventoryTransactionRepository)
	mockPORepo := new(MockPurchaseOrderRepository)
	mockSORepo := new(MockSalesOrderRepository)

	service := grpc.NewExecutionService(
		mockBalanceRepo,
		mockTxnRepo,
		mockPORepo,
		mockSORepo,
		nil, // recordTxnUseCase
	)

	orgID := uuid.New()
	productID := uuid.New()
	locationID := uuid.New()

	balances := []*domain.InventoryBalance{
		{
			ID:             uuid.New(),
			OrganizationID: orgID,
			ProductID:      productID,
			LocationID:     locationID,
			OnHand:         100.0,
			Reserved:       10.0,
			Available:      90.0,
			UpdatedAt:      time.Now(),
		},
	}

	mockBalanceRepo.On("GetByProduct", mock.Anything, orgID, productID).
		Return(balances, nil)

	ctx := context.Background()
	result, err := service.GetInventoryBalance(ctx, orgID, productID, nil)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 100.0, result.TotalOnHand)
	assert.Equal(t, 90.0, result.TotalAvailable)
	assert.Len(t, result.Balances, 1)

	mockBalanceRepo.AssertExpectations(t)
}

func TestExecutionService_GetPendingPurchaseOrders(t *testing.T) {
	mockBalanceRepo := new(MockInventoryBalanceRepository)
	mockTxnRepo := new(MockInventoryTransactionRepository)
	mockPORepo := new(MockPurchaseOrderRepository)
	mockSORepo := new(MockSalesOrderRepository)

	service := grpc.NewExecutionService(
		mockBalanceRepo,
		mockTxnRepo,
		mockPORepo,
		mockSORepo,
		nil,
	)

	orgID := uuid.New()
	productID := uuid.New()

	orders := []*domain.PurchaseOrder{
		{
			ID:                  uuid.New(),
			OrganizationID:      orgID,
			PONumber:            "PO-001",
			Status:              domain.POStatusConfirmed,
			ExpectedArrivalDate: time.Now().Add(7 * 24 * time.Hour),
			LineItems: []domain.POLineItem{
				{
					ID:          uuid.New(),
					ProductID:   productID,
					Quantity:    100.0,
					ReceivedQty: 0.0,
				},
			},
		},
	}

	mockPORepo.On("List", mock.Anything, orgID, mock.Anything, 1, 1000).
		Return(orders, int64(1), nil)

	ctx := context.Background()
	result, err := service.GetPendingPurchaseOrders(ctx, orgID, productID)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 100.0, result.TotalPendingQuantity)
	assert.Len(t, result.Orders, 1)

	mockPORepo.AssertExpectations(t)
}

func TestExecutionService_GetPendingOnOrder(t *testing.T) {
	mockBalanceRepo := new(MockInventoryBalanceRepository)
	mockTxnRepo := new(MockInventoryTransactionRepository)
	mockPORepo := new(MockPurchaseOrderRepository)
	mockSORepo := new(MockSalesOrderRepository)

	service := grpc.NewExecutionService(
		mockBalanceRepo,
		mockTxnRepo,
		mockPORepo,
		mockSORepo,
		nil,
	)

	orgID := uuid.New()
	productID := uuid.New()

	orders := []*domain.PurchaseOrder{
		{
			ID:                  uuid.New(),
			OrganizationID:      orgID,
			PONumber:            "PO-001",
			Status:              domain.POStatusConfirmed,
			ExpectedArrivalDate: time.Now().Add(7 * 24 * time.Hour),
			LineItems: []domain.POLineItem{
				{
					ID:          uuid.New(),
					ProductID:   productID,
					Quantity:    50.0,
					ReceivedQty: 0.0,
				},
			},
		},
		{
			ID:                  uuid.New(),
			OrganizationID:      orgID,
			PONumber:            "PO-002",
			Status:              domain.POStatusPartial,
			ExpectedArrivalDate: time.Now().Add(14 * 24 * time.Hour),
			LineItems: []domain.POLineItem{
				{
					ID:          uuid.New(),
					ProductID:   productID,
					Quantity:    100.0,
					ReceivedQty: 30.0,
				},
			},
		},
	}

	mockPORepo.On("List", mock.Anything, orgID, mock.Anything, 1, 1000).
		Return(orders, int64(2), nil)

	ctx := context.Background()
	result, err := service.GetPendingOnOrder(ctx, orgID, productID)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 120.0, result.OnOrder) // 50 + 70
	assert.Equal(t, int32(2), result.PendingPOCount)

	mockPORepo.AssertExpectations(t)
}

func TestExecutionService_GetQualifiedDemand(t *testing.T) {
	mockBalanceRepo := new(MockInventoryBalanceRepository)
	mockTxnRepo := new(MockInventoryTransactionRepository)
	mockPORepo := new(MockPurchaseOrderRepository)
	mockSORepo := new(MockSalesOrderRepository)

	service := grpc.NewExecutionService(
		mockBalanceRepo,
		mockTxnRepo,
		mockPORepo,
		mockSORepo,
		nil,
	)

	orgID := uuid.New()
	productID := uuid.New()

	mockSORepo.On("GetQualifiedDemand", mock.Anything, orgID, productID).
		Return(150.0, nil)

	orders := []*domain.SalesOrder{
		{
			ID:             uuid.New(),
			OrganizationID: orgID,
			Status:         domain.SOStatusConfirmed,
			LineItems: []domain.SOLineItem{
				{ProductID: productID, Quantity: 100.0},
			},
		},
		{
			ID:             uuid.New(),
			OrganizationID: orgID,
			Status:         domain.SOStatusConfirmed,
			LineItems: []domain.SOLineItem{
				{ProductID: productID, Quantity: 50.0},
			},
		},
	}

	mockSORepo.On("List", mock.Anything, orgID, mock.Anything, 1, 1000).
		Return(orders, int64(2), nil)

	ctx := context.Background()
	result, err := service.GetQualifiedDemand(ctx, orgID, productID)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 150.0, result.QualifiedDemand)
	assert.Equal(t, int32(2), result.ConfirmedSOCount)

	mockSORepo.AssertExpectations(t)
}
