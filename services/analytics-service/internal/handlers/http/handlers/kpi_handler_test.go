package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/melegattip/giia-core-engine/services/analytics-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/analytics-service/internal/handlers/http/dto"
)

// MockKPIRepository is a mock implementation of KPIRepository.
type MockKPIRepository struct {
	mock.Mock
}

func (m *MockKPIRepository) SaveKPISnapshot(ctx context.Context, snapshot *domain.KPISnapshot) error {
	args := m.Called(ctx, snapshot)
	return args.Error(0)
}

func (m *MockKPIRepository) GetKPISnapshot(ctx context.Context, organizationID uuid.UUID, date time.Time) (*domain.KPISnapshot, error) {
	args := m.Called(ctx, organizationID, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.KPISnapshot), args.Error(1)
}

func (m *MockKPIRepository) ListKPISnapshots(ctx context.Context, organizationID uuid.UUID, startDate, endDate time.Time) ([]*domain.KPISnapshot, error) {
	args := m.Called(ctx, organizationID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.KPISnapshot), args.Error(1)
}

func (m *MockKPIRepository) SaveDaysInInventoryKPI(ctx context.Context, kpi *domain.DaysInInventoryKPI) error {
	args := m.Called(ctx, kpi)
	return args.Error(0)
}

func (m *MockKPIRepository) GetDaysInInventoryKPI(ctx context.Context, organizationID uuid.UUID, date time.Time) (*domain.DaysInInventoryKPI, error) {
	args := m.Called(ctx, organizationID, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DaysInInventoryKPI), args.Error(1)
}

func (m *MockKPIRepository) ListDaysInInventoryKPI(ctx context.Context, organizationID uuid.UUID, startDate, endDate time.Time) ([]*domain.DaysInInventoryKPI, error) {
	args := m.Called(ctx, organizationID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.DaysInInventoryKPI), args.Error(1)
}

func (m *MockKPIRepository) SaveImmobilizedInventoryKPI(ctx context.Context, kpi *domain.ImmobilizedInventoryKPI) error {
	args := m.Called(ctx, kpi)
	return args.Error(0)
}

func (m *MockKPIRepository) GetImmobilizedInventoryKPI(ctx context.Context, organizationID uuid.UUID, date time.Time, thresholdYears int) (*domain.ImmobilizedInventoryKPI, error) {
	args := m.Called(ctx, organizationID, date, thresholdYears)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ImmobilizedInventoryKPI), args.Error(1)
}

func (m *MockKPIRepository) ListImmobilizedInventoryKPI(ctx context.Context, organizationID uuid.UUID, startDate, endDate time.Time, thresholdYears int) ([]*domain.ImmobilizedInventoryKPI, error) {
	args := m.Called(ctx, organizationID, startDate, endDate, thresholdYears)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.ImmobilizedInventoryKPI), args.Error(1)
}

func (m *MockKPIRepository) SaveInventoryRotationKPI(ctx context.Context, kpi *domain.InventoryRotationKPI) error {
	args := m.Called(ctx, kpi)
	return args.Error(0)
}

func (m *MockKPIRepository) GetInventoryRotationKPI(ctx context.Context, organizationID uuid.UUID, date time.Time) (*domain.InventoryRotationKPI, error) {
	args := m.Called(ctx, organizationID, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.InventoryRotationKPI), args.Error(1)
}

func (m *MockKPIRepository) ListInventoryRotationKPI(ctx context.Context, organizationID uuid.UUID, startDate, endDate time.Time) ([]*domain.InventoryRotationKPI, error) {
	args := m.Called(ctx, organizationID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.InventoryRotationKPI), args.Error(1)
}

func (m *MockKPIRepository) SaveBufferAnalytics(ctx context.Context, analytics *domain.BufferAnalytics) error {
	args := m.Called(ctx, analytics)
	return args.Error(0)
}

func (m *MockKPIRepository) GetBufferAnalyticsByProduct(ctx context.Context, organizationID, productID uuid.UUID, date time.Time) (*domain.BufferAnalytics, error) {
	args := m.Called(ctx, organizationID, productID, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.BufferAnalytics), args.Error(1)
}

func (m *MockKPIRepository) ListBufferAnalytics(ctx context.Context, organizationID uuid.UUID, startDate, endDate time.Time) ([]*domain.BufferAnalytics, error) {
	args := m.Called(ctx, organizationID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.BufferAnalytics), args.Error(1)
}

func TestKPIHandler_GetDaysInInventory_Success(t *testing.T) {
	mockRepo := new(MockKPIRepository)
	handler := NewKPIHandler(mockRepo, nil, nil, nil, nil)

	orgID := uuid.New()
	snapshotDate := time.Now().Truncate(24 * time.Hour)

	expectedKPI := &domain.DaysInInventoryKPI{
		ID:                uuid.New(),
		OrganizationID:    orgID,
		SnapshotDate:      snapshotDate,
		TotalValuedDays:   15000.50,
		AverageValuedDays: 750.25,
		TotalProducts:     20,
		CreatedAt:         time.Now(),
	}

	mockRepo.On("GetDaysInInventoryKPI", mock.Anything, orgID, mock.AnythingOfType("time.Time")).Return(expectedKPI, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/days-in-inventory?organization_id="+orgID.String(), nil)
	w := httptest.NewRecorder()

	handler.GetDaysInInventory(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.DaysInInventoryKPIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedKPI.TotalValuedDays, response.TotalValuedDays)
	assert.Equal(t, expectedKPI.TotalProducts, response.TotalProducts)

	mockRepo.AssertExpectations(t)
}

func TestKPIHandler_GetDaysInInventory_MissingOrganizationID(t *testing.T) {
	handler := NewKPIHandler(nil, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/days-in-inventory", nil)
	w := httptest.NewRecorder()

	handler.GetDaysInInventory(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "BAD_REQUEST", response.ErrorCode)
}

func TestKPIHandler_GetImmobilizedInventory_Success(t *testing.T) {
	mockRepo := new(MockKPIRepository)
	handler := NewKPIHandler(mockRepo, nil, nil, nil, nil)

	orgID := uuid.New()
	snapshotDate := time.Now().Truncate(24 * time.Hour)

	expectedKPI := &domain.ImmobilizedInventoryKPI{
		ID:                    uuid.New(),
		OrganizationID:        orgID,
		SnapshotDate:          snapshotDate,
		ThresholdYears:        2,
		ImmobilizedCount:      10,
		ImmobilizedValue:      50000.00,
		TotalStockValue:       500000.00,
		ImmobilizedPercentage: 10.0,
		CreatedAt:             time.Now(),
	}

	mockRepo.On("GetImmobilizedInventoryKPI", mock.Anything, orgID, mock.AnythingOfType("time.Time"), 2).Return(expectedKPI, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/immobilized-inventory?organization_id="+orgID.String()+"&threshold_years=2", nil)
	w := httptest.NewRecorder()

	handler.GetImmobilizedInventory(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.ImmobilizedInventoryKPIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedKPI.ImmobilizedCount, response.ImmobilizedCount)
	assert.Equal(t, expectedKPI.ImmobilizedPercentage, response.ImmobilizedPercentage)

	mockRepo.AssertExpectations(t)
}

func TestKPIHandler_GetInventoryRotation_Success(t *testing.T) {
	mockRepo := new(MockKPIRepository)
	handler := NewKPIHandler(mockRepo, nil, nil, nil, nil)

	orgID := uuid.New()
	snapshotDate := time.Now().Truncate(24 * time.Hour)

	expectedKPI := &domain.InventoryRotationKPI{
		ID:                   uuid.New(),
		OrganizationID:       orgID,
		SnapshotDate:         snapshotDate,
		SalesLast30Days:      100000.00,
		AvgMonthlyStock:      200000.00,
		RotationRatio:        0.5,
		TopRotatingProducts:  []domain.RotatingProduct{},
		SlowRotatingProducts: []domain.RotatingProduct{},
		CreatedAt:            time.Now(),
	}

	mockRepo.On("GetInventoryRotationKPI", mock.Anything, orgID, mock.AnythingOfType("time.Time")).Return(expectedKPI, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/inventory-rotation?organization_id="+orgID.String(), nil)
	w := httptest.NewRecorder()

	handler.GetInventoryRotation(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.InventoryRotationKPIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedKPI.RotationRatio, response.RotationRatio)

	mockRepo.AssertExpectations(t)
}

func TestKPIHandler_GetBufferAnalytics_ListSuccess(t *testing.T) {
	mockRepo := new(MockKPIRepository)
	handler := NewKPIHandler(mockRepo, nil, nil, nil, nil)

	orgID := uuid.New()
	productID := uuid.New()

	expectedAnalytics := []*domain.BufferAnalytics{
		{
			ID:             uuid.New(),
			ProductID:      productID,
			OrganizationID: orgID,
			Date:           time.Now(),
			CPD:            10.5,
			RedZone:        100,
			YellowZone:     200,
			GreenZone:      150,
			CreatedAt:      time.Now(),
		},
	}

	mockRepo.On("ListBufferAnalytics", mock.Anything, orgID, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).Return(expectedAnalytics, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/buffer-analytics?organization_id="+orgID.String(), nil)
	w := httptest.NewRecorder()

	handler.GetBufferAnalytics(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	mockRepo.AssertExpectations(t)
}

func TestKPIHandler_GetBufferAnalytics_SingleProduct(t *testing.T) {
	mockRepo := new(MockKPIRepository)
	handler := NewKPIHandler(mockRepo, nil, nil, nil, nil)

	orgID := uuid.New()
	productID := uuid.New()

	expectedAnalytics := &domain.BufferAnalytics{
		ID:             uuid.New(),
		ProductID:      productID,
		OrganizationID: orgID,
		Date:           time.Now(),
		CPD:            10.5,
		RedZone:        100,
		YellowZone:     200,
		GreenZone:      150,
		CreatedAt:      time.Now(),
	}

	mockRepo.On("GetBufferAnalyticsByProduct", mock.Anything, orgID, productID, mock.AnythingOfType("time.Time")).Return(expectedAnalytics, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/buffer-analytics?organization_id="+orgID.String()+"&product_id="+productID.String(), nil)
	w := httptest.NewRecorder()

	handler.GetBufferAnalytics(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	mockRepo.AssertExpectations(t)
}

func TestKPIHandler_GetSnapshot_Success(t *testing.T) {
	mockRepo := new(MockKPIRepository)
	handler := NewKPIHandler(mockRepo, nil, nil, nil, nil)

	orgID := uuid.New()
	snapshotDate := time.Now().Truncate(24 * time.Hour)

	expectedSnapshot := &domain.KPISnapshot{
		ID:                  uuid.New(),
		OrganizationID:      orgID,
		SnapshotDate:        snapshotDate,
		InventoryTurnover:   2.5,
		StockoutRate:        5.0,
		ServiceLevel:        95.0,
		ExcessInventoryPct:  10.0,
		BufferScoreGreen:    60.0,
		BufferScoreYellow:   30.0,
		BufferScoreRed:      10.0,
		TotalInventoryValue: 500000.00,
		CreatedAt:           time.Now(),
	}

	mockRepo.On("GetKPISnapshot", mock.Anything, orgID, mock.AnythingOfType("time.Time")).Return(expectedSnapshot, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/snapshot?organization_id="+orgID.String(), nil)
	w := httptest.NewRecorder()

	handler.GetSnapshot(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.KPISnapshotResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedSnapshot.InventoryTurnover, response.InventoryTurnover)
	assert.Equal(t, expectedSnapshot.ServiceLevel, response.ServiceLevel)

	mockRepo.AssertExpectations(t)
}

func TestKPIHandler_SyncBufferData_NilUseCase(t *testing.T) {
	mockRepo := new(MockKPIRepository)
	// Test that handler returns 503 when syncBufferUseCase is nil
	handler := NewKPIHandler(mockRepo, nil, nil, nil, nil)

	orgID := uuid.New()
	reqBody := dto.SyncBufferRequest{
		OrganizationID: orgID,
		Date:           time.Now(),
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/analytics/sync-buffer", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.SyncBufferData(w, req)

	// Should return 503 Service Unavailable when use case is not configured
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestKPIHandler_SyncBufferData_MissingOrganizationID(t *testing.T) {
	handler := NewKPIHandler(nil, nil, nil, nil, nil)

	reqBody := dto.SyncBufferRequest{
		Date: time.Now(),
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/analytics/sync-buffer", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.SyncBufferData(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestKPIHandler_CacheHit(t *testing.T) {
	mockRepo := new(MockKPIRepository)
	handler := NewKPIHandler(mockRepo, nil, nil, nil, nil)

	orgID := uuid.New()
	snapshotDate := time.Now().Truncate(24 * time.Hour)

	expectedKPI := &domain.DaysInInventoryKPI{
		ID:                uuid.New(),
		OrganizationID:    orgID,
		SnapshotDate:      snapshotDate,
		TotalValuedDays:   15000.50,
		AverageValuedDays: 750.25,
		TotalProducts:     20,
		CreatedAt:         time.Now(),
	}

	// First call - goes to repository
	mockRepo.On("GetDaysInInventoryKPI", mock.Anything, orgID, mock.AnythingOfType("time.Time")).Return(expectedKPI, nil).Once()

	req1 := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/days-in-inventory?organization_id="+orgID.String()+"&date="+snapshotDate.Format("2006-01-02"), nil)
	w1 := httptest.NewRecorder()
	handler.GetDaysInInventory(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Second call - should hit cache
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/days-in-inventory?organization_id="+orgID.String()+"&date="+snapshotDate.Format("2006-01-02"), nil)
	w2 := httptest.NewRecorder()
	handler.GetDaysInInventory(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	// Repository should only be called once due to caching
	mockRepo.AssertNumberOfCalls(t, "GetDaysInInventoryKPI", 1)
}
