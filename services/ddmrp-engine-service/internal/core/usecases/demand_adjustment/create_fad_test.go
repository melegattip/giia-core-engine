package demand_adjustment

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/melegattip/giia-core-engine/services/ddmrp-engine-service/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDemandAdjustmentRepository is a mock implementation of DemandAdjustmentRepository
type MockDemandAdjustmentRepository struct {
	mock.Mock
}

func (m *MockDemandAdjustmentRepository) Create(ctx context.Context, adjustment *domain.DemandAdjustment) error {
	args := m.Called(ctx, adjustment)
	return args.Error(0)
}

func (m *MockDemandAdjustmentRepository) Update(ctx context.Context, adjustment *domain.DemandAdjustment) error {
	args := m.Called(ctx, adjustment)
	return args.Error(0)
}

func (m *MockDemandAdjustmentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.DemandAdjustment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DemandAdjustment), args.Error(1)
}

func (m *MockDemandAdjustmentRepository) GetActiveForDate(ctx context.Context, productID, organizationID uuid.UUID, date time.Time) ([]domain.DemandAdjustment, error) {
	args := m.Called(ctx, productID, organizationID, mock.Anything)
	return args.Get(0).([]domain.DemandAdjustment), args.Error(1)
}

func (m *MockDemandAdjustmentRepository) ListByProduct(ctx context.Context, productID, organizationID uuid.UUID) ([]domain.DemandAdjustment, error) {
	args := m.Called(ctx, productID, organizationID)
	return args.Get(0).([]domain.DemandAdjustment), args.Error(1)
}

func (m *MockDemandAdjustmentRepository) ListByOrganization(ctx context.Context, organizationID uuid.UUID, limit, offset int) ([]domain.DemandAdjustment, error) {
	args := m.Called(ctx, organizationID, limit, offset)
	return args.Get(0).([]domain.DemandAdjustment), args.Error(1)
}

func (m *MockDemandAdjustmentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockEventPublisher is a mock implementation of EventPublisher
type MockEventPublisher struct {
	mock.Mock
}

func (m *MockEventPublisher) PublishBufferCalculated(ctx context.Context, buffer *domain.Buffer) error {
	args := m.Called(ctx, buffer)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishBufferStatusChanged(ctx context.Context, buffer *domain.Buffer, oldZone domain.ZoneType) error {
	args := m.Called(ctx, buffer, oldZone)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishBufferAlertTriggered(ctx context.Context, buffer *domain.Buffer) error {
	args := m.Called(ctx, buffer)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishFADCreated(ctx context.Context, fad *domain.DemandAdjustment) error {
	args := m.Called(ctx, fad)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishFADUpdated(ctx context.Context, fad *domain.DemandAdjustment) error {
	args := m.Called(ctx, fad)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishFADDeleted(ctx context.Context, fadID string) error {
	args := m.Called(ctx, fadID)
	return args.Error(0)
}

func TestCreateFAD_Execute_Success(t *testing.T) {
	mockRepo := new(MockDemandAdjustmentRepository)
	mockEventPublisher := new(MockEventPublisher)

	productID := uuid.New()
	orgID := uuid.New()
	createdBy := uuid.New()
	startDate := time.Now()
	endDate := startDate.AddDate(0, 1, 0)

	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.DemandAdjustment")).
		Return(nil)
	mockEventPublisher.On("PublishFADCreated", mock.Anything, mock.AnythingOfType("*domain.DemandAdjustment")).
		Return(nil)

	useCase := NewCreateFADUseCase(mockRepo, mockEventPublisher)
	result, err := useCase.Execute(context.Background(), CreateFADInput{
		ProductID:      productID,
		OrganizationID: orgID,
		StartDate:      startDate,
		EndDate:        endDate,
		AdjustmentType: domain.DemandAdjustmentFAD,
		Factor:         1.5,
		Reason:         "Holiday season adjustment",
		CreatedBy:      createdBy,
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, productID, result.ProductID)
	assert.Equal(t, orgID, result.OrganizationID)
	assert.Equal(t, 1.5, result.Factor)
	assert.Equal(t, domain.DemandAdjustmentFAD, result.AdjustmentType)

	mockRepo.AssertExpectations(t)
	mockEventPublisher.AssertExpectations(t)
}

func TestCreateFAD_Execute_ValidationError_NilProductID(t *testing.T) {
	mockRepo := new(MockDemandAdjustmentRepository)
	mockEventPublisher := new(MockEventPublisher)

	useCase := NewCreateFADUseCase(mockRepo, mockEventPublisher)
	result, err := useCase.Execute(context.Background(), CreateFADInput{
		ProductID:      uuid.Nil,
		OrganizationID: uuid.New(),
		StartDate:      time.Now(),
		EndDate:        time.Now().AddDate(0, 1, 0),
		AdjustmentType: domain.DemandAdjustmentFAD,
		Factor:         1.5,
		Reason:         "Test reason",
		CreatedBy:      uuid.New(),
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "product_id is required")
}

func TestCreateFAD_Execute_ValidationError_InvalidDateRange(t *testing.T) {
	mockRepo := new(MockDemandAdjustmentRepository)
	mockEventPublisher := new(MockEventPublisher)

	startDate := time.Now()
	endDate := startDate.AddDate(0, -1, 0) // End before start

	useCase := NewCreateFADUseCase(mockRepo, mockEventPublisher)
	result, err := useCase.Execute(context.Background(), CreateFADInput{
		ProductID:      uuid.New(),
		OrganizationID: uuid.New(),
		StartDate:      startDate,
		EndDate:        endDate,
		AdjustmentType: domain.DemandAdjustmentFAD,
		Factor:         1.5,
		Reason:         "Test reason",
		CreatedBy:      uuid.New(),
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "end_date must be >= start_date")
}

func TestCreateFAD_Execute_RepositoryError(t *testing.T) {
	mockRepo := new(MockDemandAdjustmentRepository)
	mockEventPublisher := new(MockEventPublisher)

	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.DemandAdjustment")).
		Return(domain.NewInternalError("database error"))

	useCase := NewCreateFADUseCase(mockRepo, mockEventPublisher)
	result, err := useCase.Execute(context.Background(), CreateFADInput{
		ProductID:      uuid.New(),
		OrganizationID: uuid.New(),
		StartDate:      time.Now(),
		EndDate:        time.Now().AddDate(0, 1, 0),
		AdjustmentType: domain.DemandAdjustmentFAD,
		Factor:         1.5,
		Reason:         "Test reason",
		CreatedBy:      uuid.New(),
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to create demand adjustment")

	mockRepo.AssertExpectations(t)
}

func TestCreateFAD_Execute_EventPublishError(t *testing.T) {
	mockRepo := new(MockDemandAdjustmentRepository)
	mockEventPublisher := new(MockEventPublisher)

	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.DemandAdjustment")).
		Return(nil)
	mockEventPublisher.On("PublishFADCreated", mock.Anything, mock.AnythingOfType("*domain.DemandAdjustment")).
		Return(domain.NewInternalError("event publish error"))

	useCase := NewCreateFADUseCase(mockRepo, mockEventPublisher)
	result, err := useCase.Execute(context.Background(), CreateFADInput{
		ProductID:      uuid.New(),
		OrganizationID: uuid.New(),
		StartDate:      time.Now(),
		EndDate:        time.Now().AddDate(0, 1, 0),
		AdjustmentType: domain.DemandAdjustmentFAD,
		Factor:         1.5,
		Reason:         "Test reason",
		CreatedBy:      uuid.New(),
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to publish FAD created event")

	mockRepo.AssertExpectations(t)
	mockEventPublisher.AssertExpectations(t)
}

func TestCreateFAD_Execute_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		input       CreateFADInput
		setupMock   func(repo *MockDemandAdjustmentRepository, eventPub *MockEventPublisher)
		wantErr     bool
		errContains string
	}{
		{
			name: "success",
			input: CreateFADInput{
				ProductID:      uuid.New(),
				OrganizationID: uuid.New(),
				StartDate:      time.Now(),
				EndDate:        time.Now().AddDate(0, 1, 0),
				AdjustmentType: domain.DemandAdjustmentFAD,
				Factor:         1.5,
				Reason:         "Valid reason",
				CreatedBy:      uuid.New(),
			},
			setupMock: func(repo *MockDemandAdjustmentRepository, eventPub *MockEventPublisher) {
				repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.DemandAdjustment")).
					Return(nil)
				eventPub.On("PublishFADCreated", mock.Anything, mock.AnythingOfType("*domain.DemandAdjustment")).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "nil product_id",
			input: CreateFADInput{
				ProductID:      uuid.Nil,
				OrganizationID: uuid.New(),
				StartDate:      time.Now(),
				EndDate:        time.Now().AddDate(0, 1, 0),
				AdjustmentType: domain.DemandAdjustmentFAD,
				Factor:         1.5,
				Reason:         "Valid reason",
				CreatedBy:      uuid.New(),
			},
			setupMock:   func(repo *MockDemandAdjustmentRepository, eventPub *MockEventPublisher) {},
			wantErr:     true,
			errContains: "product_id is required",
		},
		{
			name: "empty reason",
			input: CreateFADInput{
				ProductID:      uuid.New(),
				OrganizationID: uuid.New(),
				StartDate:      time.Now(),
				EndDate:        time.Now().AddDate(0, 1, 0),
				AdjustmentType: domain.DemandAdjustmentFAD,
				Factor:         1.5,
				Reason:         "",
				CreatedBy:      uuid.New(),
			},
			setupMock:   func(repo *MockDemandAdjustmentRepository, eventPub *MockEventPublisher) {},
			wantErr:     true,
			errContains: "reason is required",
		},
		{
			name: "negative factor",
			input: CreateFADInput{
				ProductID:      uuid.New(),
				OrganizationID: uuid.New(),
				StartDate:      time.Now(),
				EndDate:        time.Now().AddDate(0, 1, 0),
				AdjustmentType: domain.DemandAdjustmentFAD,
				Factor:         -1.0,
				Reason:         "Valid reason",
				CreatedBy:      uuid.New(),
			},
			setupMock:   func(repo *MockDemandAdjustmentRepository, eventPub *MockEventPublisher) {},
			wantErr:     true,
			errContains: "factor must be non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockDemandAdjustmentRepository)
			mockEventPublisher := new(MockEventPublisher)
			tt.setupMock(mockRepo, mockEventPublisher)

			useCase := NewCreateFADUseCase(mockRepo, mockEventPublisher)
			result, err := useCase.Execute(context.Background(), tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockRepo.AssertExpectations(t)
			mockEventPublisher.AssertExpectations(t)
		})
	}
}
