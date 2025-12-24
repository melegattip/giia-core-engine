package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/melegattip/giia-core-engine/services/analytics-service/internal/core/domain"
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

func TestAnalyticsService_GetKPISnapshot_Success(t *testing.T) {
	mockRepo := new(MockKPIRepository)
	service := NewAnalyticsService(mockRepo, nil, nil, nil, nil)

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

	mockRepo.On("GetKPISnapshot", mock.Anything, orgID, snapshotDate).Return(expectedSnapshot, nil)

	result, err := service.GetKPISnapshot(context.Background(), orgID, snapshotDate)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedSnapshot.ID.String(), result.ID)
	assert.Equal(t, expectedSnapshot.InventoryTurnover, result.InventoryTurnover)

	mockRepo.AssertExpectations(t)
}

func TestAnalyticsService_GetDaysInInventoryKPI_Success(t *testing.T) {
	mockRepo := new(MockKPIRepository)
	service := NewAnalyticsService(mockRepo, nil, nil, nil, nil)

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

	mockRepo.On("GetDaysInInventoryKPI", mock.Anything, orgID, snapshotDate).Return(expectedKPI, nil)

	result, err := service.GetDaysInInventoryKPI(context.Background(), orgID, snapshotDate)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedKPI.TotalValuedDays, result.TotalValuedDays)
	assert.Equal(t, int32(expectedKPI.TotalProducts), result.TotalProducts)

	mockRepo.AssertExpectations(t)
}

func TestAnalyticsService_GetImmobilizedInventoryKPI_Success(t *testing.T) {
	mockRepo := new(MockKPIRepository)
	service := NewAnalyticsService(mockRepo, nil, nil, nil, nil)

	orgID := uuid.New()
	snapshotDate := time.Now().Truncate(24 * time.Hour)
	thresholdYears := 2

	expectedKPI := &domain.ImmobilizedInventoryKPI{
		ID:                    uuid.New(),
		OrganizationID:        orgID,
		SnapshotDate:          snapshotDate,
		ThresholdYears:        thresholdYears,
		ImmobilizedCount:      10,
		ImmobilizedValue:      50000.00,
		TotalStockValue:       500000.00,
		ImmobilizedPercentage: 10.0,
		CreatedAt:             time.Now(),
	}

	mockRepo.On("GetImmobilizedInventoryKPI", mock.Anything, orgID, snapshotDate, thresholdYears).Return(expectedKPI, nil)

	result, err := service.GetImmobilizedInventoryKPI(context.Background(), orgID, snapshotDate, thresholdYears)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int32(expectedKPI.ImmobilizedCount), result.ImmobilizedCount)
	assert.Equal(t, expectedKPI.ImmobilizedPercentage, result.ImmobilizedPercentage)

	mockRepo.AssertExpectations(t)
}

func TestAnalyticsService_GetInventoryRotationKPI_Success(t *testing.T) {
	mockRepo := new(MockKPIRepository)
	service := NewAnalyticsService(mockRepo, nil, nil, nil, nil)

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

	mockRepo.On("GetInventoryRotationKPI", mock.Anything, orgID, snapshotDate).Return(expectedKPI, nil)

	result, err := service.GetInventoryRotationKPI(context.Background(), orgID, snapshotDate)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedKPI.RotationRatio, result.RotationRatio)
	assert.Equal(t, expectedKPI.SalesLast30Days, result.SalesLast30Days)

	mockRepo.AssertExpectations(t)
}

func TestAnalyticsService_GetBufferAnalytics_Success(t *testing.T) {
	mockRepo := new(MockKPIRepository)
	service := NewAnalyticsService(mockRepo, nil, nil, nil, nil)

	orgID := uuid.New()
	productID := uuid.New()
	date := time.Now().Truncate(24 * time.Hour)

	expectedAnalytics := &domain.BufferAnalytics{
		ID:                uuid.New(),
		ProductID:         productID,
		OrganizationID:    orgID,
		Date:              date,
		CPD:               10.5,
		RedZone:           100,
		RedBase:           50,
		RedSafe:           50,
		YellowZone:        200,
		GreenZone:         150,
		LTD:               14,
		LeadTimeFactor:    1.0,
		VariabilityFactor: 0.5,
		MOQ:               10,
		OrderFrequency:    7,
		OptimalOrderFreq:  14.28,
		SafetyDays:        9.52,
		AvgOpenOrders:     1.33,
		HasAdjustments:    false,
		CreatedAt:         time.Now(),
	}

	mockRepo.On("GetBufferAnalyticsByProduct", mock.Anything, orgID, productID, date).Return(expectedAnalytics, nil)

	result, err := service.GetBufferAnalytics(context.Background(), orgID, productID, date)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedAnalytics.CPD, result.CPD)
	assert.Equal(t, expectedAnalytics.RedZone, result.RedZone)

	mockRepo.AssertExpectations(t)
}

func TestAnalyticsService_ListKPISnapshots_Success(t *testing.T) {
	mockRepo := new(MockKPIRepository)
	service := NewAnalyticsService(mockRepo, nil, nil, nil, nil)

	orgID := uuid.New()
	startDate := time.Now().AddDate(0, 0, -30)
	endDate := time.Now()

	expectedSnapshots := []*domain.KPISnapshot{
		{
			ID:                uuid.New(),
			OrganizationID:    orgID,
			SnapshotDate:      time.Now(),
			InventoryTurnover: 2.5,
			CreatedAt:         time.Now(),
		},
		{
			ID:                uuid.New(),
			OrganizationID:    orgID,
			SnapshotDate:      time.Now().AddDate(0, 0, -7),
			InventoryTurnover: 2.3,
			CreatedAt:         time.Now(),
		},
	}

	mockRepo.On("ListKPISnapshots", mock.Anything, orgID, startDate, endDate).Return(expectedSnapshots, nil)

	result, err := service.ListKPISnapshots(context.Background(), orgID, startDate, endDate)

	assert.NoError(t, err)
	assert.Len(t, result, 2)

	mockRepo.AssertExpectations(t)
}

func TestAnalyticsService_ListBufferAnalytics_Success(t *testing.T) {
	mockRepo := new(MockKPIRepository)
	service := NewAnalyticsService(mockRepo, nil, nil, nil, nil)

	orgID := uuid.New()
	startDate := time.Now().AddDate(0, 0, -30)
	endDate := time.Now()

	expectedAnalytics := []*domain.BufferAnalytics{
		{
			ID:             uuid.New(),
			ProductID:      uuid.New(),
			OrganizationID: orgID,
			Date:           time.Now(),
			CPD:            10.5,
			RedZone:        100,
			YellowZone:     200,
			GreenZone:      150,
			CreatedAt:      time.Now(),
		},
	}

	mockRepo.On("ListBufferAnalytics", mock.Anything, orgID, startDate, endDate).Return(expectedAnalytics, nil)

	result, err := service.ListBufferAnalytics(context.Background(), orgID, startDate, endDate)

	assert.NoError(t, err)
	assert.Len(t, result, 1)

	mockRepo.AssertExpectations(t)
}

func TestAnalyticsService_CacheHit(t *testing.T) {
	mockRepo := new(MockKPIRepository)
	service := NewAnalyticsService(mockRepo, nil, nil, nil, nil)

	orgID := uuid.New()
	snapshotDate := time.Now().Truncate(24 * time.Hour)

	expectedSnapshot := &domain.KPISnapshot{
		ID:                uuid.New(),
		OrganizationID:    orgID,
		SnapshotDate:      snapshotDate,
		InventoryTurnover: 2.5,
		CreatedAt:         time.Now(),
	}

	// Only expect one call to the repository
	mockRepo.On("GetKPISnapshot", mock.Anything, orgID, snapshotDate).Return(expectedSnapshot, nil).Once()

	// First call - goes to repository
	_, err := service.GetKPISnapshot(context.Background(), orgID, snapshotDate)
	assert.NoError(t, err)

	// Second call - should hit cache
	result, err := service.GetKPISnapshot(context.Background(), orgID, snapshotDate)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Repository should only be called once
	mockRepo.AssertNumberOfCalls(t, "GetKPISnapshot", 1)
}

func TestAnalyticsService_SyncBufferData_NotConfigured(t *testing.T) {
	service := NewAnalyticsService(nil, nil, nil, nil, nil)

	orgID := uuid.New()
	date := time.Now()

	_, err := service.SyncBufferData(context.Background(), orgID, date)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sync use case not configured")
}
