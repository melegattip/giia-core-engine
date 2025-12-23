package kpi_test

import (
	"context"
	"testing"
	"time"

	"github.com/melegattip/giia-core-engine/services/analytics-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/analytics-service/internal/core/providers"
	"github.com/melegattip/giia-core-engine/services/analytics-service/internal/core/usecases/kpi"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockDDMRPServiceClient struct {
	mock.Mock
}

func (m *MockDDMRPServiceClient) GetBufferHistory(ctx context.Context, organizationID, productID uuid.UUID, date time.Time) (*providers.BufferHistory, error) {
	args := m.Called(ctx, organizationID, productID, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*providers.BufferHistory), args.Error(1)
}

func (m *MockDDMRPServiceClient) ListBufferHistory(ctx context.Context, organizationID uuid.UUID, startDate, endDate time.Time) ([]*providers.BufferHistory, error) {
	args := m.Called(ctx, organizationID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*providers.BufferHistory), args.Error(1)
}

func (m *MockDDMRPServiceClient) GetBufferZoneDistribution(ctx context.Context, organizationID uuid.UUID, date time.Time) (*providers.BufferZoneDistribution, error) {
	args := m.Called(ctx, organizationID, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*providers.BufferZoneDistribution), args.Error(1)
}

func TestSyncBufferAnalyticsUseCase_Execute_WithValidData_SyncsAnalytics(t *testing.T) {
	givenOrgID := uuid.New()
	givenDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	productID1 := uuid.New()
	productID2 := uuid.New()

	givenBufferHistories := []*providers.BufferHistory{
		{
			ProductID:         productID1,
			OrganizationID:    givenOrgID,
			Date:              givenDate,
			CPD:               10.0,
			RedZone:           50.0,
			RedBase:           20.0,
			RedSafe:           30.0,
			YellowZone:        100.0,
			GreenZone:         150.0,
			LTD:               15,
			LeadTimeFactor:    1.5,
			VariabilityFactor: 1.2,
			MOQ:               10,
			OrderFrequency:    7,
			HasAdjustments:    false,
		},
		{
			ProductID:         productID2,
			OrganizationID:    givenOrgID,
			Date:              givenDate,
			CPD:               20.0,
			RedZone:           80.0,
			RedBase:           30.0,
			RedSafe:           50.0,
			YellowZone:        150.0,
			GreenZone:         200.0,
			LTD:               20,
			LeadTimeFactor:    1.8,
			VariabilityFactor: 1.3,
			MOQ:               15,
			OrderFrequency:    10,
			HasAdjustments:    true,
		},
	}

	mockKPIRepo := new(MockKPIRepository)
	mockDDMRPClient := new(MockDDMRPServiceClient)

	mockDDMRPClient.On("ListBufferHistory", mock.Anything, givenOrgID, givenDate, givenDate).Return(givenBufferHistories, nil)
	mockKPIRepo.On("SaveBufferAnalytics", mock.Anything, mock.MatchedBy(func(analytics *domain.BufferAnalytics) bool {
		return analytics.OrganizationID == givenOrgID
	})).Return(nil).Times(2)

	useCase := kpi.NewSyncBufferAnalyticsUseCase(mockKPIRepo, mockDDMRPClient)

	input := &kpi.SyncBufferAnalyticsInput{
		OrganizationID: givenOrgID,
		Date:           givenDate,
	}

	synced, err := useCase.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.Equal(t, 2, synced)
	mockKPIRepo.AssertExpectations(t)
	mockDDMRPClient.AssertExpectations(t)
}

func TestSyncBufferAnalyticsUseCase_Execute_WithNilInput_ReturnsError(t *testing.T) {
	mockKPIRepo := new(MockKPIRepository)
	mockDDMRPClient := new(MockDDMRPServiceClient)

	useCase := kpi.NewSyncBufferAnalyticsUseCase(mockKPIRepo, mockDDMRPClient)

	synced, err := useCase.Execute(context.Background(), nil)

	assert.Error(t, err)
	assert.Equal(t, 0, synced)
	assert.Contains(t, err.Error(), "input cannot be nil")
}

func TestSyncBufferAnalyticsUseCase_Execute_WithNilOrganizationID_ReturnsError(t *testing.T) {
	mockKPIRepo := new(MockKPIRepository)
	mockDDMRPClient := new(MockDDMRPServiceClient)

	useCase := kpi.NewSyncBufferAnalyticsUseCase(mockKPIRepo, mockDDMRPClient)

	input := &kpi.SyncBufferAnalyticsInput{
		OrganizationID: uuid.Nil,
		Date:           time.Now(),
	}

	synced, err := useCase.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Equal(t, 0, synced)
	assert.Contains(t, err.Error(), "organization_id is required")
}

func TestSyncBufferAnalyticsUseCase_Execute_WithZeroDate_ReturnsError(t *testing.T) {
	mockKPIRepo := new(MockKPIRepository)
	mockDDMRPClient := new(MockDDMRPServiceClient)

	useCase := kpi.NewSyncBufferAnalyticsUseCase(mockKPIRepo, mockDDMRPClient)

	input := &kpi.SyncBufferAnalyticsInput{
		OrganizationID: uuid.New(),
		Date:           time.Time{},
	}

	synced, err := useCase.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Equal(t, 0, synced)
	assert.Contains(t, err.Error(), "date is required")
}

func TestSyncBufferAnalyticsUseCase_Execute_WhenDDMRPClientFails_ReturnsError(t *testing.T) {
	givenOrgID := uuid.New()
	givenDate := time.Now()

	mockKPIRepo := new(MockKPIRepository)
	mockDDMRPClient := new(MockDDMRPServiceClient)

	mockDDMRPClient.On("ListBufferHistory", mock.Anything, givenOrgID, givenDate, givenDate).Return(nil, assert.AnError)

	useCase := kpi.NewSyncBufferAnalyticsUseCase(mockKPIRepo, mockDDMRPClient)

	input := &kpi.SyncBufferAnalyticsInput{
		OrganizationID: givenOrgID,
		Date:           givenDate,
	}

	synced, err := useCase.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Equal(t, 0, synced)
	mockDDMRPClient.AssertExpectations(t)
}

func TestSyncBufferAnalyticsUseCase_Execute_WithEmptyBufferHistories_ReturnsZeroSynced(t *testing.T) {
	givenOrgID := uuid.New()
	givenDate := time.Now()

	mockKPIRepo := new(MockKPIRepository)
	mockDDMRPClient := new(MockDDMRPServiceClient)

	mockDDMRPClient.On("ListBufferHistory", mock.Anything, givenOrgID, givenDate, givenDate).Return([]*providers.BufferHistory{}, nil)

	useCase := kpi.NewSyncBufferAnalyticsUseCase(mockKPIRepo, mockDDMRPClient)

	input := &kpi.SyncBufferAnalyticsInput{
		OrganizationID: givenOrgID,
		Date:           givenDate,
	}

	synced, err := useCase.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.Equal(t, 0, synced)
	mockDDMRPClient.AssertExpectations(t)
}

func TestSyncBufferAnalyticsUseCase_Execute_WhenSomeRecordsFail_ContinuesSyncing(t *testing.T) {
	givenOrgID := uuid.New()
	givenDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	productID1 := uuid.New()
	productID2 := uuid.New()
	productID3 := uuid.New()

	givenBufferHistories := []*providers.BufferHistory{
		{
			ProductID:         productID1,
			OrganizationID:    givenOrgID,
			Date:              givenDate,
			CPD:               10.0,
			RedZone:           50.0,
			RedBase:           20.0,
			RedSafe:           30.0,
			YellowZone:        100.0,
			GreenZone:         150.0,
			LTD:               15,
			LeadTimeFactor:    1.5,
			VariabilityFactor: 1.2,
			MOQ:               10,
			OrderFrequency:    7,
			HasAdjustments:    false,
		},
		{
			ProductID:         productID2,
			OrganizationID:    givenOrgID,
			Date:              givenDate,
			CPD:               -10.0,
			RedZone:           50.0,
			RedBase:           20.0,
			RedSafe:           30.0,
			YellowZone:        100.0,
			GreenZone:         150.0,
			LTD:               15,
			LeadTimeFactor:    1.5,
			VariabilityFactor: 1.2,
			MOQ:               10,
			OrderFrequency:    7,
			HasAdjustments:    false,
		},
		{
			ProductID:         productID3,
			OrganizationID:    givenOrgID,
			Date:              givenDate,
			CPD:               20.0,
			RedZone:           80.0,
			RedBase:           30.0,
			RedSafe:           50.0,
			YellowZone:        150.0,
			GreenZone:         200.0,
			LTD:               20,
			LeadTimeFactor:    1.8,
			VariabilityFactor: 1.3,
			MOQ:               15,
			OrderFrequency:    10,
			HasAdjustments:    true,
		},
	}

	mockKPIRepo := new(MockKPIRepository)
	mockDDMRPClient := new(MockDDMRPServiceClient)

	mockDDMRPClient.On("ListBufferHistory", mock.Anything, givenOrgID, givenDate, givenDate).Return(givenBufferHistories, nil)
	mockKPIRepo.On("SaveBufferAnalytics", mock.Anything, mock.MatchedBy(func(analytics *domain.BufferAnalytics) bool {
		return analytics.ProductID == productID1 || analytics.ProductID == productID3
	})).Return(nil).Times(2)

	useCase := kpi.NewSyncBufferAnalyticsUseCase(mockKPIRepo, mockDDMRPClient)

	input := &kpi.SyncBufferAnalyticsInput{
		OrganizationID: givenOrgID,
		Date:           givenDate,
	}

	synced, err := useCase.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.Equal(t, 2, synced)
	mockKPIRepo.AssertExpectations(t)
	mockDDMRPClient.AssertExpectations(t)
}
