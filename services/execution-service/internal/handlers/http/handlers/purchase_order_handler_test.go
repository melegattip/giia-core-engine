package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/usecases/purchase_order"
	"github.com/melegattip/giia-core-engine/services/execution-service/internal/handlers/http/dto"
	"github.com/melegattip/giia-core-engine/services/execution-service/internal/handlers/http/handlers"
	"github.com/melegattip/giia-core-engine/services/execution-service/internal/handlers/http/middleware"
)

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

// MockCatalogClient is a mock for CatalogServiceClient.
type MockCatalogClient struct {
	mock.Mock
}

func (m *MockCatalogClient) GetProduct(ctx context.Context, productID uuid.UUID) (*Product, error) {
	args := m.Called(ctx, productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Product), args.Error(1)
}

func (m *MockCatalogClient) GetSupplier(ctx context.Context, supplierID uuid.UUID) (*Supplier, error) {
	args := m.Called(ctx, supplierID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Supplier), args.Error(1)
}

func (m *MockCatalogClient) GetProductsByIDs(ctx context.Context, productIDs []uuid.UUID) ([]*Product, error) {
	args := m.Called(ctx, productIDs)
	return args.Get(0).([]*Product), args.Error(1)
}

type Product struct {
	ID            uuid.UUID
	SKU           string
	Name          string
	UnitOfMeasure string
}

type Supplier struct {
	ID   uuid.UUID
	Name string
	Code string
}

// MockEventPublisher is a mock for EventPublisher.
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

// Helper to add auth context to request.
func withAuthContext(r *http.Request, userID, orgID uuid.UUID) *http.Request {
	ctx := context.WithValue(r.Context(), middleware.UserIDKey, userID)
	ctx = context.WithValue(ctx, middleware.OrganizationIDKey, orgID)
	return r.WithContext(ctx)
}

func TestPurchaseOrderHandler_List(t *testing.T) {
	mockRepo := new(MockPurchaseOrderRepository)

	handler := handlers.NewPurchaseOrderHandler(
		nil, // createUC
		nil, // receiveUC
		nil, // cancelUC
		mockRepo,
	)

	userID := uuid.New()
	orgID := uuid.New()

	orders := []*domain.PurchaseOrder{
		{
			ID:             uuid.New(),
			OrganizationID: orgID,
			PONumber:       "PO-001",
			Status:         domain.POStatusDraft,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
	}

	mockRepo.On("List", mock.Anything, orgID, mock.Anything, 1, 20).
		Return(orders, int64(1), nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/purchase-orders?page=1&page_size=20", nil)
	req = withAuthContext(req, userID, orgID)
	w := httptest.NewRecorder()

	handler.List(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestPurchaseOrderHandler_Get(t *testing.T) {
	mockRepo := new(MockPurchaseOrderRepository)

	handler := handlers.NewPurchaseOrderHandler(
		nil, nil, nil, mockRepo,
	)

	userID := uuid.New()
	orgID := uuid.New()
	poID := uuid.New()

	po := &domain.PurchaseOrder{
		ID:             poID,
		OrganizationID: orgID,
		PONumber:       "PO-001",
		Status:         domain.POStatusDraft,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	mockRepo.On("GetByID", mock.Anything, poID, orgID).Return(po, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/purchase-orders/"+poID.String(), nil)
	req = withAuthContext(req, userID, orgID)

	// Set URL params with chi
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", poID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.Get(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.PurchaseOrderResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, poID.String(), response.ID)

	mockRepo.AssertExpectations(t)
}

func TestPurchaseOrderHandler_Get_InvalidID(t *testing.T) {
	mockRepo := new(MockPurchaseOrderRepository)

	handler := handlers.NewPurchaseOrderHandler(
		nil, nil, nil, mockRepo,
	)

	userID := uuid.New()
	orgID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/purchase-orders/invalid-id", nil)
	req = withAuthContext(req, userID, orgID)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "invalid-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.Get(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPurchaseOrderHandler_Create_MissingAuth(t *testing.T) {
	mockRepo := new(MockPurchaseOrderRepository)

	handler := handlers.NewPurchaseOrderHandler(
		nil, nil, nil, mockRepo,
	)

	reqBody := dto.CreatePurchaseOrderRequest{
		PONumber:   "PO-001",
		SupplierID: uuid.New().String(),
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/purchase-orders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// No auth context

	w := httptest.NewRecorder()

	handler.Create(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestPurchaseOrderHandler_Cancel(t *testing.T) {
	mockRepo := new(MockPurchaseOrderRepository)
	mockPublisher := new(MockEventPublisher)

	cancelUC := purchase_order.NewCancelPOUseCase(mockRepo, mockPublisher)

	handler := handlers.NewPurchaseOrderHandler(
		nil, nil, cancelUC, mockRepo,
	)

	userID := uuid.New()
	orgID := uuid.New()
	poID := uuid.New()

	po := &domain.PurchaseOrder{
		ID:             poID,
		OrganizationID: orgID,
		PONumber:       "PO-001",
		Status:         domain.POStatusDraft,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	mockRepo.On("GetByID", mock.Anything, poID, orgID).Return(po, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.PurchaseOrder")).Return(nil)
	mockPublisher.On("PublishPOCancelled", mock.Anything, mock.AnythingOfType("*domain.PurchaseOrder")).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/purchase-orders/"+poID.String()+"/cancel", nil)
	req = withAuthContext(req, userID, orgID)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", poID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.Cancel(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.PurchaseOrderResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "cancelled", response.Status)

	mockRepo.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}
