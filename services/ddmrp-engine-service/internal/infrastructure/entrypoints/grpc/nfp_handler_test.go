package grpc

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/melegattip/giia-core-engine/services/ddmrp-engine-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/ddmrp-engine-service/internal/core/usecases/nfp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUpdateNFPUseCase is a mock for testing
type MockUpdateNFPUseCase struct {
	mock.Mock
}

func (m *MockUpdateNFPUseCase) Execute(ctx context.Context, input nfp.UpdateNFPInput) (*domain.Buffer, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Buffer), args.Error(1)
}

// MockCheckReplenishmentUseCase is a mock for testing
type MockCheckReplenishmentUseCase struct {
	mock.Mock
}

func (m *MockCheckReplenishmentUseCase) Execute(ctx context.Context, organizationID uuid.UUID) ([]domain.Buffer, error) {
	args := m.Called(ctx, organizationID)
	return args.Get(0).([]domain.Buffer), args.Error(1)
}

func TestNFPHandler_UpdateNFP_Success(t *testing.T) {
	mockUpdateUC := new(MockUpdateNFPUseCase)

	productID := uuid.New()
	orgID := uuid.New()

	expectedBuffer := &domain.Buffer{
		ID:              uuid.New(),
		ProductID:       productID,
		OrganizationID:  orgID,
		OnHand:          500.0,
		OnOrder:         100.0,
		QualifiedDemand: 50.0,
		NetFlowPosition: 550.0,
		Zone:            domain.ZoneGreen,
		AlertLevel:      domain.AlertNormal,
	}

	mockUpdateUC.On("Execute", mock.Anything, nfp.UpdateNFPInput{
		ProductID:       productID,
		OrganizationID:  orgID,
		OnHand:          500.0,
		OnOrder:         100.0,
		QualifiedDemand: 50.0,
	}).Return(expectedBuffer, nil)

	result, err := mockUpdateUC.Execute(context.Background(), nfp.UpdateNFPInput{
		ProductID:       productID,
		OrganizationID:  orgID,
		OnHand:          500.0,
		OnOrder:         100.0,
		QualifiedDemand: 50.0,
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 500.0, result.OnHand)
	assert.Equal(t, 100.0, result.OnOrder)
	assert.Equal(t, 50.0, result.QualifiedDemand)
	assert.Equal(t, domain.ZoneGreen, result.Zone)

	mockUpdateUC.AssertExpectations(t)
}

func TestNFPHandler_UpdateNFP_ZoneChange(t *testing.T) {
	mockUpdateUC := new(MockUpdateNFPUseCase)

	productID := uuid.New()
	orgID := uuid.New()

	// Low on-hand triggers red zone
	expectedBuffer := &domain.Buffer{
		ID:              uuid.New(),
		ProductID:       productID,
		OrganizationID:  orgID,
		OnHand:          50.0,
		OnOrder:         0.0,
		QualifiedDemand: 100.0,
		NetFlowPosition: -50.0,
		Zone:            domain.ZoneBelowRed,
		AlertLevel:      domain.AlertCritical,
	}

	mockUpdateUC.On("Execute", mock.Anything, mock.Anything).Return(expectedBuffer, nil)

	result, err := mockUpdateUC.Execute(context.Background(), nfp.UpdateNFPInput{
		ProductID:       productID,
		OrganizationID:  orgID,
		OnHand:          50.0,
		OnOrder:         0.0,
		QualifiedDemand: 100.0,
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, domain.ZoneBelowRed, result.Zone)
	assert.Equal(t, domain.AlertCritical, result.AlertLevel)

	mockUpdateUC.AssertExpectations(t)
}

func TestNFPHandler_UpdateNFP_BufferNotFound(t *testing.T) {
	mockUpdateUC := new(MockUpdateNFPUseCase)

	productID := uuid.New()
	orgID := uuid.New()

	mockUpdateUC.On("Execute", mock.Anything, mock.Anything).
		Return(nil, domain.NewNotFoundError("buffer not found"))

	result, err := mockUpdateUC.Execute(context.Background(), nfp.UpdateNFPInput{
		ProductID:       productID,
		OrganizationID:  orgID,
		OnHand:          100.0,
		OnOrder:         50.0,
		QualifiedDemand: 25.0,
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "buffer not found")

	mockUpdateUC.AssertExpectations(t)
}

func TestNFPHandler_UpdateNFP_ValidationError(t *testing.T) {
	mockUpdateUC := new(MockUpdateNFPUseCase)

	mockUpdateUC.On("Execute", mock.Anything, mock.Anything).
		Return(nil, domain.NewValidationError("product_id is required"))

	result, err := mockUpdateUC.Execute(context.Background(), nfp.UpdateNFPInput{
		ProductID:       uuid.Nil,
		OrganizationID:  uuid.New(),
		OnHand:          100.0,
		OnOrder:         50.0,
		QualifiedDemand: 25.0,
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "product_id is required")

	mockUpdateUC.AssertExpectations(t)
}

func TestNFPHandler_CheckReplenishment_Success(t *testing.T) {
	mockCheckUC := new(MockCheckReplenishmentUseCase)

	orgID := uuid.New()

	expectedBuffers := []domain.Buffer{
		{ID: uuid.New(), OrganizationID: orgID, AlertLevel: domain.AlertCritical, Zone: domain.ZoneBelowRed},
		{ID: uuid.New(), OrganizationID: orgID, AlertLevel: domain.AlertCritical, Zone: domain.ZoneBelowRed},
		{ID: uuid.New(), OrganizationID: orgID, AlertLevel: domain.AlertReplenish, Zone: domain.ZoneRed},
	}

	mockCheckUC.On("Execute", mock.Anything, orgID).Return(expectedBuffers, nil)

	result, err := mockCheckUC.Execute(context.Background(), orgID)

	assert.NoError(t, err)
	assert.Len(t, result, 3)

	// Verify we have both critical and replenish buffers
	criticalCount := 0
	replenishCount := 0
	for _, buf := range result {
		if buf.AlertLevel == domain.AlertCritical {
			criticalCount++
		} else if buf.AlertLevel == domain.AlertReplenish {
			replenishCount++
		}
	}
	assert.Equal(t, 2, criticalCount)
	assert.Equal(t, 1, replenishCount)

	mockCheckUC.AssertExpectations(t)
}

func TestNFPHandler_CheckReplenishment_Empty(t *testing.T) {
	mockCheckUC := new(MockCheckReplenishmentUseCase)

	orgID := uuid.New()

	mockCheckUC.On("Execute", mock.Anything, orgID).Return([]domain.Buffer{}, nil)

	result, err := mockCheckUC.Execute(context.Background(), orgID)

	assert.NoError(t, err)
	assert.Empty(t, result)

	mockCheckUC.AssertExpectations(t)
}

func TestNFPHandler_CheckReplenishment_ValidationError(t *testing.T) {
	mockCheckUC := new(MockCheckReplenishmentUseCase)

	mockCheckUC.On("Execute", mock.Anything, uuid.Nil).
		Return([]domain.Buffer{}, domain.NewValidationError("organization_id is required"))

	result, err := mockCheckUC.Execute(context.Background(), uuid.Nil)

	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Contains(t, err.Error(), "organization_id is required")

	mockCheckUC.AssertExpectations(t)
}

func TestNFPHandler_CheckReplenishment_RepositoryError(t *testing.T) {
	mockCheckUC := new(MockCheckReplenishmentUseCase)

	orgID := uuid.New()

	mockCheckUC.On("Execute", mock.Anything, orgID).
		Return([]domain.Buffer{}, domain.NewInternalError("database error"))

	result, err := mockCheckUC.Execute(context.Background(), orgID)

	assert.Error(t, err)
	assert.Empty(t, result)

	mockCheckUC.AssertExpectations(t)
}

func TestNFPHandler_TableDriven(t *testing.T) {
	tests := []struct {
		name           string
		testType       string // "updateNFP" or "checkReplenishment"
		productID      uuid.UUID
		organizationID uuid.UUID
		setupMock      func(updateUC *MockUpdateNFPUseCase, checkUC *MockCheckReplenishmentUseCase)
		wantErr        bool
		errContains    string
	}{
		{
			name:           "updateNFP - success",
			testType:       "updateNFP",
			productID:      uuid.New(),
			organizationID: uuid.New(),
			setupMock: func(updateUC *MockUpdateNFPUseCase, checkUC *MockCheckReplenishmentUseCase) {
				updateUC.On("Execute", mock.Anything, mock.Anything).
					Return(&domain.Buffer{ID: uuid.New(), Zone: domain.ZoneGreen}, nil)
			},
			wantErr: false,
		},
		{
			name:           "updateNFP - not found",
			testType:       "updateNFP",
			productID:      uuid.New(),
			organizationID: uuid.New(),
			setupMock: func(updateUC *MockUpdateNFPUseCase, checkUC *MockCheckReplenishmentUseCase) {
				updateUC.On("Execute", mock.Anything, mock.Anything).
					Return(nil, domain.NewNotFoundError("buffer not found"))
			},
			wantErr:     true,
			errContains: "buffer not found",
		},
		{
			name:           "checkReplenishment - success with buffers",
			testType:       "checkReplenishment",
			organizationID: uuid.New(),
			setupMock: func(updateUC *MockUpdateNFPUseCase, checkUC *MockCheckReplenishmentUseCase) {
				checkUC.On("Execute", mock.Anything, mock.Anything).
					Return([]domain.Buffer{{ID: uuid.New(), AlertLevel: domain.AlertCritical}}, nil)
			},
			wantErr: false,
		},
		{
			name:           "checkReplenishment - empty",
			testType:       "checkReplenishment",
			organizationID: uuid.New(),
			setupMock: func(updateUC *MockUpdateNFPUseCase, checkUC *MockCheckReplenishmentUseCase) {
				checkUC.On("Execute", mock.Anything, mock.Anything).
					Return([]domain.Buffer{}, nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUpdateUC := new(MockUpdateNFPUseCase)
			mockCheckUC := new(MockCheckReplenishmentUseCase)
			tt.setupMock(mockUpdateUC, mockCheckUC)

			var err error

			if tt.testType == "updateNFP" {
				_, err = mockUpdateUC.Execute(context.Background(), nfp.UpdateNFPInput{
					ProductID:       tt.productID,
					OrganizationID:  tt.organizationID,
					OnHand:          100.0,
					OnOrder:         50.0,
					QualifiedDemand: 25.0,
				})
			} else {
				_, err = mockCheckUC.Execute(context.Background(), tt.organizationID)
			}

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}

			mockUpdateUC.AssertExpectations(t)
			mockCheckUC.AssertExpectations(t)
		})
	}
}
