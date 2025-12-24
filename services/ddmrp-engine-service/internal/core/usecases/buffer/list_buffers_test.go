package buffer

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/melegattip/giia-core-engine/services/ddmrp-engine-service/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListBuffers_Execute_Success_NoFilters(t *testing.T) {
	mockBufferRepo := new(MockBufferRepository)

	orgID := uuid.New()
	expectedBuffers := []domain.Buffer{
		{ID: uuid.New(), ProductID: uuid.New(), OrganizationID: orgID, Zone: domain.ZoneGreen},
		{ID: uuid.New(), ProductID: uuid.New(), OrganizationID: orgID, Zone: domain.ZoneYellow},
	}

	mockBufferRepo.On("List", mock.Anything, orgID, 10, 0).
		Return(expectedBuffers, nil)

	useCase := NewListBuffersUseCase(mockBufferRepo)
	result, err := useCase.Execute(context.Background(), ListBuffersInput{
		OrganizationID: orgID,
		Limit:          10,
		Offset:         0,
	})

	assert.NoError(t, err)
	assert.Len(t, result, 2)

	mockBufferRepo.AssertExpectations(t)
}

func TestListBuffers_Execute_Success_FilterByZone(t *testing.T) {
	mockBufferRepo := new(MockBufferRepository)

	orgID := uuid.New()
	expectedBuffers := []domain.Buffer{
		{ID: uuid.New(), ProductID: uuid.New(), OrganizationID: orgID, Zone: domain.ZoneRed},
	}

	mockBufferRepo.On("ListByZone", mock.Anything, orgID, domain.ZoneRed).
		Return(expectedBuffers, nil)

	useCase := NewListBuffersUseCase(mockBufferRepo)
	result, err := useCase.Execute(context.Background(), ListBuffersInput{
		OrganizationID: orgID,
		Zone:           domain.ZoneRed,
	})

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, domain.ZoneRed, result[0].Zone)

	mockBufferRepo.AssertExpectations(t)
}

func TestListBuffers_Execute_Success_FilterByAlertLevel(t *testing.T) {
	mockBufferRepo := new(MockBufferRepository)

	orgID := uuid.New()
	expectedBuffers := []domain.Buffer{
		{ID: uuid.New(), ProductID: uuid.New(), OrganizationID: orgID, AlertLevel: domain.AlertCritical},
	}

	mockBufferRepo.On("ListByAlertLevel", mock.Anything, orgID, domain.AlertCritical).
		Return(expectedBuffers, nil)

	useCase := NewListBuffersUseCase(mockBufferRepo)
	result, err := useCase.Execute(context.Background(), ListBuffersInput{
		OrganizationID: orgID,
		AlertLevel:     domain.AlertCritical,
	})

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, domain.AlertCritical, result[0].AlertLevel)

	mockBufferRepo.AssertExpectations(t)
}

func TestListBuffers_Execute_NilOrganizationID(t *testing.T) {
	mockBufferRepo := new(MockBufferRepository)

	useCase := NewListBuffersUseCase(mockBufferRepo)
	result, err := useCase.Execute(context.Background(), ListBuffersInput{
		OrganizationID: uuid.Nil,
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "organization_id is required")
}

func TestListBuffers_Execute_EmptyResult(t *testing.T) {
	mockBufferRepo := new(MockBufferRepository)

	orgID := uuid.New()
	mockBufferRepo.On("List", mock.Anything, orgID, 10, 0).
		Return([]domain.Buffer{}, nil)

	useCase := NewListBuffersUseCase(mockBufferRepo)
	result, err := useCase.Execute(context.Background(), ListBuffersInput{
		OrganizationID: orgID,
		Limit:          10,
		Offset:         0,
	})

	assert.NoError(t, err)
	assert.Empty(t, result)

	mockBufferRepo.AssertExpectations(t)
}

func TestListBuffers_Execute_RepositoryError(t *testing.T) {
	mockBufferRepo := new(MockBufferRepository)

	orgID := uuid.New()
	mockBufferRepo.On("List", mock.Anything, orgID, 10, 0).
		Return([]domain.Buffer{}, domain.NewInternalError("database error"))

	useCase := NewListBuffersUseCase(mockBufferRepo)
	result, err := useCase.Execute(context.Background(), ListBuffersInput{
		OrganizationID: orgID,
		Limit:          10,
		Offset:         0,
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to list buffers")

	mockBufferRepo.AssertExpectations(t)
}

func TestListBuffers_Execute_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		input       ListBuffersInput
		setupMock   func(m *MockBufferRepository)
		wantLen     int
		wantErr     bool
		errContains string
	}{
		{
			name: "success - no filters",
			input: ListBuffersInput{
				OrganizationID: uuid.New(),
				Limit:          10,
				Offset:         0,
			},
			setupMock: func(m *MockBufferRepository) {
				m.On("List", mock.Anything, mock.Anything, 10, 0).
					Return([]domain.Buffer{{ID: uuid.New()}, {ID: uuid.New()}}, nil)
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "success - filter by zone",
			input: ListBuffersInput{
				OrganizationID: uuid.New(),
				Zone:           domain.ZoneYellow,
			},
			setupMock: func(m *MockBufferRepository) {
				m.On("ListByZone", mock.Anything, mock.Anything, domain.ZoneYellow).
					Return([]domain.Buffer{{ID: uuid.New(), Zone: domain.ZoneYellow}}, nil)
			},
			wantLen: 1,
			wantErr: false,
		},
		{
			name: "success - filter by alert level",
			input: ListBuffersInput{
				OrganizationID: uuid.New(),
				AlertLevel:     domain.AlertReplenish,
			},
			setupMock: func(m *MockBufferRepository) {
				m.On("ListByAlertLevel", mock.Anything, mock.Anything, domain.AlertReplenish).
					Return([]domain.Buffer{{ID: uuid.New(), AlertLevel: domain.AlertReplenish}}, nil)
			},
			wantLen: 1,
			wantErr: false,
		},
		{
			name: "nil organization_id",
			input: ListBuffersInput{
				OrganizationID: uuid.Nil,
			},
			setupMock:   func(m *MockBufferRepository) {},
			wantErr:     true,
			errContains: "organization_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBufferRepo := new(MockBufferRepository)
			tt.setupMock(mockBufferRepo)

			useCase := NewListBuffersUseCase(mockBufferRepo)
			result, err := useCase.Execute(context.Background(), tt.input)

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
