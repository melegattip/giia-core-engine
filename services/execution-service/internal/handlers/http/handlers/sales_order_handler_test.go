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
	"github.com/melegattip/giia-core-engine/services/execution-service/internal/handlers/http/dto"
	"github.com/melegattip/giia-core-engine/services/execution-service/internal/handlers/http/handlers"
	"github.com/melegattip/giia-core-engine/services/execution-service/internal/handlers/http/middleware"
)

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

func TestSalesOrderHandler_List(t *testing.T) {
	mockRepo := new(MockSalesOrderRepository)

	handler := handlers.NewSalesOrderHandler(nil, nil, mockRepo)

	userID := uuid.New()
	orgID := uuid.New()

	orders := []*domain.SalesOrder{
		{
			ID:             uuid.New(),
			OrganizationID: orgID,
			SONumber:       "SO-001",
			Status:         domain.SOStatusPending,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
	}

	mockRepo.On("List", mock.Anything, orgID, mock.Anything, 1, 20).
		Return(orders, int64(1), nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/sales-orders?page=1&page_size=20", nil)
	req = withSOAuthContext(req, userID, orgID)
	w := httptest.NewRecorder()

	handler.List(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestSalesOrderHandler_Get(t *testing.T) {
	mockRepo := new(MockSalesOrderRepository)

	handler := handlers.NewSalesOrderHandler(nil, nil, mockRepo)

	userID := uuid.New()
	orgID := uuid.New()
	soID := uuid.New()

	so := &domain.SalesOrder{
		ID:             soID,
		OrganizationID: orgID,
		SONumber:       "SO-001",
		Status:         domain.SOStatusPending,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	mockRepo.On("GetByID", mock.Anything, soID, orgID).Return(so, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/sales-orders/"+soID.String(), nil)
	req = withSOAuthContext(req, userID, orgID)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", soID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.Get(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.SalesOrderResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, soID.String(), response.ID)

	mockRepo.AssertExpectations(t)
}

func TestSalesOrderHandler_Get_InvalidID(t *testing.T) {
	mockRepo := new(MockSalesOrderRepository)

	handler := handlers.NewSalesOrderHandler(nil, nil, mockRepo)

	userID := uuid.New()
	orgID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/sales-orders/invalid-id", nil)
	req = withSOAuthContext(req, userID, orgID)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "invalid-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.Get(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSalesOrderHandler_Create_MissingAuth(t *testing.T) {
	mockRepo := new(MockSalesOrderRepository)

	handler := handlers.NewSalesOrderHandler(nil, nil, mockRepo)

	reqBody := dto.CreateSalesOrderRequest{
		SONumber:   "SO-001",
		CustomerID: uuid.New().String(),
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/sales-orders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// No auth context

	w := httptest.NewRecorder()

	handler.Create(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSalesOrderHandler_Cancel(t *testing.T) {
	mockRepo := new(MockSalesOrderRepository)

	handler := handlers.NewSalesOrderHandler(nil, nil, mockRepo)

	userID := uuid.New()
	orgID := uuid.New()
	soID := uuid.New()

	so := &domain.SalesOrder{
		ID:             soID,
		OrganizationID: orgID,
		SONumber:       "SO-001",
		Status:         domain.SOStatusPending,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	mockRepo.On("GetByID", mock.Anything, soID, orgID).Return(so, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.SalesOrder")).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/sales-orders/"+soID.String()+"/cancel", nil)
	req = withSOAuthContext(req, userID, orgID)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", soID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.Cancel(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.SalesOrderResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "cancelled", response.Status)

	mockRepo.AssertExpectations(t)
}

func TestSalesOrderHandler_Ship_InvalidID(t *testing.T) {
	mockRepo := new(MockSalesOrderRepository)

	handler := handlers.NewSalesOrderHandler(nil, nil, mockRepo)

	userID := uuid.New()
	orgID := uuid.New()

	reqBody := dto.ShipSalesOrderRequest{
		LocationID:         uuid.New().String(),
		DeliveryNoteNumber: "DN-001",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/sales-orders/invalid-id/ship", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withSOAuthContext(req, userID, orgID)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "invalid-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.Ship(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Helper to add auth context to request.
func withSOAuthContext(r *http.Request, userID, orgID uuid.UUID) *http.Request {
	ctx := context.WithValue(r.Context(), middleware.UserIDKey, userID)
	ctx = context.WithValue(ctx, middleware.OrganizationIDKey, orgID)
	return r.WithContext(ctx)
}
