package nfp

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/melegattip/giia-core-engine/services/ddmrp-engine-service/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCheckReplenishment_Execute_Success_HasBuffersNeedingReplenishment(t *testing.T) {
	mockBufferRepo := new(MockBufferRepository)

	orgID := uuid.New()
	replenishBuffers := []domain.Buffer{
		{ID: uuid.New(), OrganizationID: orgID, AlertLevel: domain.AlertReplenish, Zone: domain.ZoneRed},
	}
	criticalBuffers := []domain.Buffer{
		{ID: uuid.New(), OrganizationID: orgID, AlertLevel: domain.AlertCritical, Zone: domain.ZoneBelowRed},
		{ID: uuid.New(), OrganizationID: orgID, AlertLevel: domain.AlertCritical, Zone: domain.ZoneBelowRed},
	}

	mockBufferRepo.On("ListByAlertLevel", mock.Anything, orgID, domain.AlertReplenish).
		Return(replenishBuffers, nil)
	mockBufferRepo.On("ListByAlertLevel", mock.Anything, orgID, domain.AlertCritical).
		Return(criticalBuffers, nil)

	useCase := NewCheckReplenishmentUseCase(mockBufferRepo)
	result, err := useCase.Execute(context.Background(), orgID)

	assert.NoError(t, err)
	assert.Len(t, result, 3) // 2 critical + 1 replenish

	mockBufferRepo.AssertExpectations(t)
}

func TestCheckReplenishment_Execute_Success_NoBuffersNeedingReplenishment(t *testing.T) {
	mockBufferRepo := new(MockBufferRepository)

	orgID := uuid.New()

	mockBufferRepo.On("ListByAlertLevel", mock.Anything, orgID, domain.AlertReplenish).
		Return([]domain.Buffer{}, nil)
	mockBufferRepo.On("ListByAlertLevel", mock.Anything, orgID, domain.AlertCritical).
		Return([]domain.Buffer{}, nil)

	useCase := NewCheckReplenishmentUseCase(mockBufferRepo)
	result, err := useCase.Execute(context.Background(), orgID)

	assert.NoError(t, err)
	assert.Empty(t, result)

	mockBufferRepo.AssertExpectations(t)
}

func TestCheckReplenishment_Execute_NilOrganizationID(t *testing.T) {
	mockBufferRepo := new(MockBufferRepository)

	useCase := NewCheckReplenishmentUseCase(mockBufferRepo)
	result, err := useCase.Execute(context.Background(), uuid.Nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "organization_id is required")
}

func TestCheckReplenishment_Execute_ReplenishQueryError(t *testing.T) {
	mockBufferRepo := new(MockBufferRepository)

	orgID := uuid.New()

	mockBufferRepo.On("ListByAlertLevel", mock.Anything, orgID, domain.AlertReplenish).
		Return([]domain.Buffer{}, domain.NewInternalError("database error"))

	useCase := NewCheckReplenishmentUseCase(mockBufferRepo)
	result, err := useCase.Execute(context.Background(), orgID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to list replenish buffers")

	mockBufferRepo.AssertExpectations(t)
}

func TestCheckReplenishment_Execute_CriticalQueryError(t *testing.T) {
	mockBufferRepo := new(MockBufferRepository)

	orgID := uuid.New()

	mockBufferRepo.On("ListByAlertLevel", mock.Anything, orgID, domain.AlertReplenish).
		Return([]domain.Buffer{}, nil)
	mockBufferRepo.On("ListByAlertLevel", mock.Anything, orgID, domain.AlertCritical).
		Return([]domain.Buffer{}, domain.NewInternalError("database error"))

	useCase := NewCheckReplenishmentUseCase(mockBufferRepo)
	result, err := useCase.Execute(context.Background(), orgID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to list critical buffers")

	mockBufferRepo.AssertExpectations(t)
}

func TestCheckReplenishment_Execute_OnlyCriticalBuffers(t *testing.T) {
	mockBufferRepo := new(MockBufferRepository)

	orgID := uuid.New()
	criticalBuffers := []domain.Buffer{
		{ID: uuid.New(), OrganizationID: orgID, AlertLevel: domain.AlertCritical},
	}

	mockBufferRepo.On("ListByAlertLevel", mock.Anything, orgID, domain.AlertReplenish).
		Return([]domain.Buffer{}, nil)
	mockBufferRepo.On("ListByAlertLevel", mock.Anything, orgID, domain.AlertCritical).
		Return(criticalBuffers, nil)

	useCase := NewCheckReplenishmentUseCase(mockBufferRepo)
	result, err := useCase.Execute(context.Background(), orgID)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, domain.AlertCritical, result[0].AlertLevel)

	mockBufferRepo.AssertExpectations(t)
}

func TestCheckReplenishment_Execute_OnlyReplenishBuffers(t *testing.T) {
	mockBufferRepo := new(MockBufferRepository)

	orgID := uuid.New()
	replenishBuffers := []domain.Buffer{
		{ID: uuid.New(), OrganizationID: orgID, AlertLevel: domain.AlertReplenish},
	}

	mockBufferRepo.On("ListByAlertLevel", mock.Anything, orgID, domain.AlertReplenish).
		Return(replenishBuffers, nil)
	mockBufferRepo.On("ListByAlertLevel", mock.Anything, orgID, domain.AlertCritical).
		Return([]domain.Buffer{}, nil)

	useCase := NewCheckReplenishmentUseCase(mockBufferRepo)
	result, err := useCase.Execute(context.Background(), orgID)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, domain.AlertReplenish, result[0].AlertLevel)

	mockBufferRepo.AssertExpectations(t)
}

func TestCheckReplenishment_Execute_TableDriven(t *testing.T) {
	tests := []struct {
		name           string
		organizationID uuid.UUID
		setupMock      func(m *MockBufferRepository)
		wantLen        int
		wantErr        bool
		errContains    string
	}{
		{
			name:           "success - combined buffers",
			organizationID: uuid.New(),
			setupMock: func(m *MockBufferRepository) {
				m.On("ListByAlertLevel", mock.Anything, mock.Anything, domain.AlertReplenish).
					Return([]domain.Buffer{{ID: uuid.New()}}, nil)
				m.On("ListByAlertLevel", mock.Anything, mock.Anything, domain.AlertCritical).
					Return([]domain.Buffer{{ID: uuid.New()}, {ID: uuid.New()}}, nil)
			},
			wantLen: 3,
			wantErr: false,
		},
		{
			name:           "success - empty results",
			organizationID: uuid.New(),
			setupMock: func(m *MockBufferRepository) {
				m.On("ListByAlertLevel", mock.Anything, mock.Anything, domain.AlertReplenish).
					Return([]domain.Buffer{}, nil)
				m.On("ListByAlertLevel", mock.Anything, mock.Anything, domain.AlertCritical).
					Return([]domain.Buffer{}, nil)
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name:           "nil organization_id",
			organizationID: uuid.Nil,
			setupMock:      func(m *MockBufferRepository) {},
			wantErr:        true,
			errContains:    "organization_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBufferRepo := new(MockBufferRepository)
			tt.setupMock(mockBufferRepo)

			useCase := NewCheckReplenishmentUseCase(mockBufferRepo)
			result, err := useCase.Execute(context.Background(), tt.organizationID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, tt.wantLen)
			}

			mockBufferRepo.AssertExpectations(t)
		})
	}
}
