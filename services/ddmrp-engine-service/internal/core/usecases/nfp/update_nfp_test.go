package nfp

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/melegattip/giia-core-engine/services/ddmrp-engine-service/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockBufferRepository is a mock implementation of BufferRepository
type MockBufferRepository struct {
	mock.Mock
}

func (m *MockBufferRepository) Create(ctx context.Context, buffer *domain.Buffer) error {
	args := m.Called(ctx, buffer)
	return args.Error(0)
}

func (m *MockBufferRepository) Save(ctx context.Context, buffer *domain.Buffer) error {
	args := m.Called(ctx, buffer)
	return args.Error(0)
}

func (m *MockBufferRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Buffer, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Buffer), args.Error(1)
}

func (m *MockBufferRepository) GetByProduct(ctx context.Context, productID, organizationID uuid.UUID) (*domain.Buffer, error) {
	args := m.Called(ctx, productID, organizationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Buffer), args.Error(1)
}

func (m *MockBufferRepository) List(ctx context.Context, organizationID uuid.UUID, limit, offset int) ([]domain.Buffer, error) {
	args := m.Called(ctx, organizationID, limit, offset)
	return args.Get(0).([]domain.Buffer), args.Error(1)
}

func (m *MockBufferRepository) ListByZone(ctx context.Context, organizationID uuid.UUID, zone domain.ZoneType) ([]domain.Buffer, error) {
	args := m.Called(ctx, organizationID, zone)
	return args.Get(0).([]domain.Buffer), args.Error(1)
}

func (m *MockBufferRepository) ListByAlertLevel(ctx context.Context, organizationID uuid.UUID, alertLevel domain.AlertLevel) ([]domain.Buffer, error) {
	args := m.Called(ctx, organizationID, alertLevel)
	return args.Get(0).([]domain.Buffer), args.Error(1)
}

func (m *MockBufferRepository) ListAll(ctx context.Context, organizationID uuid.UUID) ([]domain.Buffer, error) {
	args := m.Called(ctx, organizationID)
	return args.Get(0).([]domain.Buffer), args.Error(1)
}

func (m *MockBufferRepository) UpdateNFP(ctx context.Context, bufferID uuid.UUID, onHand, onOrder, qualifiedDemand float64) error {
	args := m.Called(ctx, bufferID, onHand, onOrder, qualifiedDemand)
	return args.Error(0)
}

func (m *MockBufferRepository) Delete(ctx context.Context, id uuid.UUID) error {
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

func TestUpdateNFP_Execute_Success_NoZoneChange(t *testing.T) {
	mockBufferRepo := new(MockBufferRepository)
	mockEventPublisher := new(MockEventPublisher)

	productID := uuid.New()
	orgID := uuid.New()
	existingBuffer := &domain.Buffer{
		ID:              uuid.New(),
		ProductID:       productID,
		OrganizationID:  orgID,
		RedZone:         100,
		YellowZone:      200,
		GreenZone:       300,
		TopOfRed:        100,
		TopOfYellow:     300,
		TopOfGreen:      600,
		OnHand:          500,
		OnOrder:         100,
		QualifiedDemand: 50,
		Zone:            domain.ZoneGreen,
		AlertLevel:      domain.AlertNormal,
	}

	mockBufferRepo.On("GetByProduct", mock.Anything, productID, orgID).
		Return(existingBuffer, nil)
	mockBufferRepo.On("Save", mock.Anything, mock.AnythingOfType("*domain.Buffer")).
		Return(nil)

	useCase := NewUpdateNFPUseCase(mockBufferRepo, mockEventPublisher)
	result, err := useCase.Execute(context.Background(), UpdateNFPInput{
		ProductID:       productID,
		OrganizationID:  orgID,
		OnHand:          500,
		OnOrder:         100,
		QualifiedDemand: 50,
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 500.0, result.OnHand)
	assert.Equal(t, 100.0, result.OnOrder)
	assert.Equal(t, 50.0, result.QualifiedDemand)

	mockBufferRepo.AssertExpectations(t)
}

func TestUpdateNFP_Execute_Success_ZoneChangeTriggersEvent(t *testing.T) {
	mockBufferRepo := new(MockBufferRepository)
	mockEventPublisher := new(MockEventPublisher)

	productID := uuid.New()
	orgID := uuid.New()
	existingBuffer := &domain.Buffer{
		ID:              uuid.New(),
		ProductID:       productID,
		OrganizationID:  orgID,
		RedZone:         100,
		YellowZone:      200,
		GreenZone:       300,
		TopOfRed:        100,
		TopOfYellow:     300,
		TopOfGreen:      600,
		OnHand:          500,
		OnOrder:         100,
		QualifiedDemand: 50,
		Zone:            domain.ZoneGreen,
		AlertLevel:      domain.AlertNormal,
	}

	mockBufferRepo.On("GetByProduct", mock.Anything, productID, orgID).
		Return(existingBuffer, nil)
	mockBufferRepo.On("Save", mock.Anything, mock.AnythingOfType("*domain.Buffer")).
		Return(nil)
	mockEventPublisher.On("PublishBufferStatusChanged", mock.Anything, mock.AnythingOfType("*domain.Buffer"), domain.ZoneGreen).
		Return(nil)
	mockEventPublisher.On("PublishBufferAlertTriggered", mock.Anything, mock.AnythingOfType("*domain.Buffer")).
		Return(nil)

	useCase := NewUpdateNFPUseCase(mockBufferRepo, mockEventPublisher)
	result, err := useCase.Execute(context.Background(), UpdateNFPInput{
		ProductID:       productID,
		OrganizationID:  orgID,
		OnHand:          50, // Low on hand triggers red zone
		OnOrder:         0,
		QualifiedDemand: 100,
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, domain.ZoneBelowRed, result.Zone)

	mockBufferRepo.AssertExpectations(t)
	mockEventPublisher.AssertExpectations(t)
}

func TestUpdateNFP_Execute_BufferNotFound(t *testing.T) {
	mockBufferRepo := new(MockBufferRepository)
	mockEventPublisher := new(MockEventPublisher)

	productID := uuid.New()
	orgID := uuid.New()

	mockBufferRepo.On("GetByProduct", mock.Anything, productID, orgID).
		Return(nil, domain.NewNotFoundError("buffer not found"))

	useCase := NewUpdateNFPUseCase(mockBufferRepo, mockEventPublisher)
	result, err := useCase.Execute(context.Background(), UpdateNFPInput{
		ProductID:       productID,
		OrganizationID:  orgID,
		OnHand:          100,
		OnOrder:         50,
		QualifiedDemand: 25,
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "buffer not found")

	mockBufferRepo.AssertExpectations(t)
}

func TestUpdateNFP_Execute_NilProductID(t *testing.T) {
	mockBufferRepo := new(MockBufferRepository)
	mockEventPublisher := new(MockEventPublisher)

	useCase := NewUpdateNFPUseCase(mockBufferRepo, mockEventPublisher)
	result, err := useCase.Execute(context.Background(), UpdateNFPInput{
		ProductID:       uuid.Nil,
		OrganizationID:  uuid.New(),
		OnHand:          100,
		OnOrder:         50,
		QualifiedDemand: 25,
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "product_id is required")
}

func TestUpdateNFP_Execute_NilOrganizationID(t *testing.T) {
	mockBufferRepo := new(MockBufferRepository)
	mockEventPublisher := new(MockEventPublisher)

	useCase := NewUpdateNFPUseCase(mockBufferRepo, mockEventPublisher)
	result, err := useCase.Execute(context.Background(), UpdateNFPInput{
		ProductID:       uuid.New(),
		OrganizationID:  uuid.Nil,
		OnHand:          100,
		OnOrder:         50,
		QualifiedDemand: 25,
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "organization_id is required")
}

func TestUpdateNFP_Execute_SaveFailure(t *testing.T) {
	mockBufferRepo := new(MockBufferRepository)
	mockEventPublisher := new(MockEventPublisher)

	productID := uuid.New()
	orgID := uuid.New()
	existingBuffer := &domain.Buffer{
		ID:             uuid.New(),
		ProductID:      productID,
		OrganizationID: orgID,
		RedZone:        100,
		YellowZone:     200,
		GreenZone:      300,
		Zone:           domain.ZoneGreen,
	}

	mockBufferRepo.On("GetByProduct", mock.Anything, productID, orgID).
		Return(existingBuffer, nil)
	mockBufferRepo.On("Save", mock.Anything, mock.AnythingOfType("*domain.Buffer")).
		Return(domain.NewInternalError("database error"))

	useCase := NewUpdateNFPUseCase(mockBufferRepo, mockEventPublisher)
	result, err := useCase.Execute(context.Background(), UpdateNFPInput{
		ProductID:       productID,
		OrganizationID:  orgID,
		OnHand:          100,
		OnOrder:         50,
		QualifiedDemand: 25,
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to update buffer NFP")

	mockBufferRepo.AssertExpectations(t)
}

func TestUpdateNFP_Execute_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		input       UpdateNFPInput
		setupMock   func(bufferRepo *MockBufferRepository, eventPub *MockEventPublisher)
		wantErr     bool
		errContains string
	}{
		{
			name: "success - updates NFP values",
			input: UpdateNFPInput{
				ProductID:       uuid.New(),
				OrganizationID:  uuid.New(),
				OnHand:          500,
				OnOrder:         100,
				QualifiedDemand: 50,
			},
			setupMock: func(bufferRepo *MockBufferRepository, eventPub *MockEventPublisher) {
				bufferRepo.On("GetByProduct", mock.Anything, mock.Anything, mock.Anything).
					Return(&domain.Buffer{
						ID:          uuid.New(),
						RedZone:     100,
						YellowZone:  200,
						GreenZone:   300,
						TopOfRed:    100,
						TopOfYellow: 300,
						TopOfGreen:  600,
						Zone:        domain.ZoneGreen,
					}, nil)
				bufferRepo.On("Save", mock.Anything, mock.AnythingOfType("*domain.Buffer")).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "nil product_id",
			input: UpdateNFPInput{
				ProductID:       uuid.Nil,
				OrganizationID:  uuid.New(),
				OnHand:          100,
				OnOrder:         50,
				QualifiedDemand: 25,
			},
			setupMock:   func(bufferRepo *MockBufferRepository, eventPub *MockEventPublisher) {},
			wantErr:     true,
			errContains: "product_id is required",
		},
		{
			name: "nil organization_id",
			input: UpdateNFPInput{
				ProductID:       uuid.New(),
				OrganizationID:  uuid.Nil,
				OnHand:          100,
				OnOrder:         50,
				QualifiedDemand: 25,
			},
			setupMock:   func(bufferRepo *MockBufferRepository, eventPub *MockEventPublisher) {},
			wantErr:     true,
			errContains: "organization_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBufferRepo := new(MockBufferRepository)
			mockEventPublisher := new(MockEventPublisher)
			tt.setupMock(mockBufferRepo, mockEventPublisher)

			useCase := NewUpdateNFPUseCase(mockBufferRepo, mockEventPublisher)
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

			mockBufferRepo.AssertExpectations(t)
			mockEventPublisher.AssertExpectations(t)
		})
	}
}
