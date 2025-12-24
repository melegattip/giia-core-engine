package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/execution-service/internal/handlers/http/handlers"
	"github.com/melegattip/giia-core-engine/services/execution-service/internal/handlers/http/middleware"
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

func TestInventoryHandler_GetBalances_ByProduct(t *testing.T) {
	mockBalanceRepo := new(MockInventoryBalanceRepository)
	mockTxnRepo := new(MockInventoryTransactionRepository)

	handler := handlers.NewInventoryHandler(mockBalanceRepo, mockTxnRepo)

	userID := uuid.New()
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

	req := httptest.NewRequest(http.MethodGet, "/api/v1/inventory/balances?product_id="+productID.String(), nil)
	req = withInvAuthContext(req, userID, orgID)
	w := httptest.NewRecorder()

	handler.GetBalances(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Contains(t, response, "balances")

	mockBalanceRepo.AssertExpectations(t)
}

func TestInventoryHandler_GetBalances_ByLocation(t *testing.T) {
	mockBalanceRepo := new(MockInventoryBalanceRepository)
	mockTxnRepo := new(MockInventoryTransactionRepository)

	handler := handlers.NewInventoryHandler(mockBalanceRepo, mockTxnRepo)

	userID := uuid.New()
	orgID := uuid.New()
	locationID := uuid.New()

	balances := []*domain.InventoryBalance{
		{
			ID:             uuid.New(),
			OrganizationID: orgID,
			ProductID:      uuid.New(),
			LocationID:     locationID,
			OnHand:         50.0,
			Reserved:       5.0,
			Available:      45.0,
			UpdatedAt:      time.Now(),
		},
	}

	mockBalanceRepo.On("GetByLocation", mock.Anything, orgID, locationID).
		Return(balances, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/inventory/balances?location_id="+locationID.String(), nil)
	req = withInvAuthContext(req, userID, orgID)
	w := httptest.NewRecorder()

	handler.GetBalances(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockBalanceRepo.AssertExpectations(t)
}

func TestInventoryHandler_GetBalances_MissingParams(t *testing.T) {
	mockBalanceRepo := new(MockInventoryBalanceRepository)
	mockTxnRepo := new(MockInventoryTransactionRepository)

	handler := handlers.NewInventoryHandler(mockBalanceRepo, mockTxnRepo)

	userID := uuid.New()
	orgID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/inventory/balances", nil)
	req = withInvAuthContext(req, userID, orgID)
	w := httptest.NewRecorder()

	handler.GetBalances(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestInventoryHandler_GetBalances_InvalidProductID(t *testing.T) {
	mockBalanceRepo := new(MockInventoryBalanceRepository)
	mockTxnRepo := new(MockInventoryTransactionRepository)

	handler := handlers.NewInventoryHandler(mockBalanceRepo, mockTxnRepo)

	userID := uuid.New()
	orgID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/inventory/balances?product_id=invalid", nil)
	req = withInvAuthContext(req, userID, orgID)
	w := httptest.NewRecorder()

	handler.GetBalances(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestInventoryHandler_GetTransactions(t *testing.T) {
	mockBalanceRepo := new(MockInventoryBalanceRepository)
	mockTxnRepo := new(MockInventoryTransactionRepository)

	handler := handlers.NewInventoryHandler(mockBalanceRepo, mockTxnRepo)

	userID := uuid.New()
	orgID := uuid.New()
	productID := uuid.New()

	transactions := []*domain.InventoryTransaction{
		{
			ID:              uuid.New(),
			OrganizationID:  orgID,
			ProductID:       productID,
			LocationID:      uuid.New(),
			Type:            domain.TransactionReceipt,
			Quantity:        100.0,
			UnitCost:        10.0,
			ReferenceType:   "purchase_order",
			ReferenceID:     uuid.New(),
			TransactionDate: time.Now(),
			CreatedBy:       userID,
			CreatedAt:       time.Now(),
		},
	}

	mockTxnRepo.On("List", mock.Anything, orgID, productID, mock.Anything, 1, 20).
		Return(transactions, int64(1), nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/inventory/transactions?product_id="+productID.String()+"&page=1&page_size=20", nil)
	req = withInvAuthContext(req, userID, orgID)
	w := httptest.NewRecorder()

	handler.GetTransactions(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockTxnRepo.AssertExpectations(t)
}

func TestInventoryHandler_GetTransactions_MissingAuth(t *testing.T) {
	mockBalanceRepo := new(MockInventoryBalanceRepository)
	mockTxnRepo := new(MockInventoryTransactionRepository)

	handler := handlers.NewInventoryHandler(mockBalanceRepo, mockTxnRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/inventory/transactions", nil)
	// No auth context

	w := httptest.NewRecorder()

	handler.GetTransactions(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// Helper to add auth context to request.
func withInvAuthContext(r *http.Request, userID, orgID uuid.UUID) *http.Request {
	ctx := context.WithValue(r.Context(), middleware.UserIDKey, userID)
	ctx = context.WithValue(ctx, middleware.OrganizationIDKey, orgID)
	return r.WithContext(ctx)
}
