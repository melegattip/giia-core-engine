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

func TestUpdateFAD_Execute_Success(t *testing.T) {
	mockRepo := new(MockDemandAdjustmentRepository)
	mockEventPublisher := new(MockEventPublisher)

	fadID := uuid.New()
	existingFAD := &domain.DemandAdjustment{
		ID:             fadID,
		ProductID:      uuid.New(),
		OrganizationID: uuid.New(),
		StartDate:      time.Now(),
		EndDate:        time.Now().AddDate(0, 1, 0),
		AdjustmentType: domain.DemandAdjustmentFAD,
		Factor:         1.2,
		Reason:         "Original reason",
		CreatedBy:      uuid.New(),
	}

	newStartDate := time.Now().AddDate(0, 0, 5)
	newEndDate := newStartDate.AddDate(0, 2, 0)

	mockRepo.On("GetByID", mock.Anything, fadID).Return(existingFAD, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.DemandAdjustment")).Return(nil)
	mockEventPublisher.On("PublishFADUpdated", mock.Anything, mock.AnythingOfType("*domain.DemandAdjustment")).Return(nil)

	useCase := NewUpdateFADUseCase(mockRepo, mockEventPublisher)
	result, err := useCase.Execute(context.Background(), UpdateFADInput{
		ID:             fadID,
		StartDate:      newStartDate,
		EndDate:        newEndDate,
		AdjustmentType: domain.DemandAdjustmentSeasonal,
		Factor:         1.8,
		Reason:         "Updated reason",
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1.8, result.Factor)
	assert.Equal(t, domain.DemandAdjustmentSeasonal, result.AdjustmentType)
	assert.Equal(t, "Updated reason", result.Reason)

	mockRepo.AssertExpectations(t)
	mockEventPublisher.AssertExpectations(t)
}

func TestUpdateFAD_Execute_FADNotFound(t *testing.T) {
	mockRepo := new(MockDemandAdjustmentRepository)
	mockEventPublisher := new(MockEventPublisher)

	fadID := uuid.New()

	mockRepo.On("GetByID", mock.Anything, fadID).
		Return(nil, domain.NewNotFoundError("demand adjustment not found"))

	useCase := NewUpdateFADUseCase(mockRepo, mockEventPublisher)
	result, err := useCase.Execute(context.Background(), UpdateFADInput{
		ID:             fadID,
		StartDate:      time.Now(),
		EndDate:        time.Now().AddDate(0, 1, 0),
		AdjustmentType: domain.DemandAdjustmentFAD,
		Factor:         1.5,
		Reason:         "Test reason",
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "demand adjustment not found")

	mockRepo.AssertExpectations(t)
}

func TestUpdateFAD_Execute_NilID(t *testing.T) {
	mockRepo := new(MockDemandAdjustmentRepository)
	mockEventPublisher := new(MockEventPublisher)

	useCase := NewUpdateFADUseCase(mockRepo, mockEventPublisher)
	result, err := useCase.Execute(context.Background(), UpdateFADInput{
		ID:             uuid.Nil,
		StartDate:      time.Now(),
		EndDate:        time.Now().AddDate(0, 1, 0),
		AdjustmentType: domain.DemandAdjustmentFAD,
		Factor:         1.5,
		Reason:         "Test reason",
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "id is required")
}

func TestUpdateFAD_Execute_ValidationError(t *testing.T) {
	mockRepo := new(MockDemandAdjustmentRepository)
	mockEventPublisher := new(MockEventPublisher)

	fadID := uuid.New()
	existingFAD := &domain.DemandAdjustment{
		ID:             fadID,
		ProductID:      uuid.New(),
		OrganizationID: uuid.New(),
		StartDate:      time.Now(),
		EndDate:        time.Now().AddDate(0, 1, 0),
		AdjustmentType: domain.DemandAdjustmentFAD,
		Factor:         1.2,
		Reason:         "Original reason",
		CreatedBy:      uuid.New(),
	}

	startDate := time.Now()
	endDate := startDate.AddDate(0, -1, 0) // End before start

	mockRepo.On("GetByID", mock.Anything, fadID).Return(existingFAD, nil)

	useCase := NewUpdateFADUseCase(mockRepo, mockEventPublisher)
	result, err := useCase.Execute(context.Background(), UpdateFADInput{
		ID:             fadID,
		StartDate:      startDate,
		EndDate:        endDate,
		AdjustmentType: domain.DemandAdjustmentFAD,
		Factor:         1.5,
		Reason:         "Test reason",
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "end_date must be >= start_date")

	mockRepo.AssertExpectations(t)
}

func TestUpdateFAD_Execute_RepositoryUpdateError(t *testing.T) {
	mockRepo := new(MockDemandAdjustmentRepository)
	mockEventPublisher := new(MockEventPublisher)

	fadID := uuid.New()
	existingFAD := &domain.DemandAdjustment{
		ID:             fadID,
		ProductID:      uuid.New(),
		OrganizationID: uuid.New(),
		StartDate:      time.Now(),
		EndDate:        time.Now().AddDate(0, 1, 0),
		AdjustmentType: domain.DemandAdjustmentFAD,
		Factor:         1.2,
		Reason:         "Original reason",
		CreatedBy:      uuid.New(),
	}

	mockRepo.On("GetByID", mock.Anything, fadID).Return(existingFAD, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.DemandAdjustment")).
		Return(domain.NewInternalError("database error"))

	useCase := NewUpdateFADUseCase(mockRepo, mockEventPublisher)
	result, err := useCase.Execute(context.Background(), UpdateFADInput{
		ID:             fadID,
		StartDate:      time.Now(),
		EndDate:        time.Now().AddDate(0, 1, 0),
		AdjustmentType: domain.DemandAdjustmentFAD,
		Factor:         1.5,
		Reason:         "Updated reason",
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to update demand adjustment")

	mockRepo.AssertExpectations(t)
}

func TestUpdateFAD_Execute_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		input       UpdateFADInput
		setupMock   func(repo *MockDemandAdjustmentRepository, eventPub *MockEventPublisher)
		wantErr     bool
		errContains string
	}{
		{
			name: "success",
			input: UpdateFADInput{
				ID:             uuid.New(),
				StartDate:      time.Now(),
				EndDate:        time.Now().AddDate(0, 1, 0),
				AdjustmentType: domain.DemandAdjustmentFAD,
				Factor:         1.5,
				Reason:         "Valid reason",
			},
			setupMock: func(repo *MockDemandAdjustmentRepository, eventPub *MockEventPublisher) {
				repo.On("GetByID", mock.Anything, mock.Anything).Return(&domain.DemandAdjustment{
					ID:             uuid.New(),
					ProductID:      uuid.New(),
					OrganizationID: uuid.New(),
					CreatedBy:      uuid.New(),
				}, nil)
				repo.On("Update", mock.Anything, mock.AnythingOfType("*domain.DemandAdjustment")).Return(nil)
				eventPub.On("PublishFADUpdated", mock.Anything, mock.AnythingOfType("*domain.DemandAdjustment")).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "nil id",
			input: UpdateFADInput{
				ID:             uuid.Nil,
				StartDate:      time.Now(),
				EndDate:        time.Now().AddDate(0, 1, 0),
				AdjustmentType: domain.DemandAdjustmentFAD,
				Factor:         1.5,
				Reason:         "Valid reason",
			},
			setupMock:   func(repo *MockDemandAdjustmentRepository, eventPub *MockEventPublisher) {},
			wantErr:     true,
			errContains: "id is required",
		},
		{
			name: "not found",
			input: UpdateFADInput{
				ID:             uuid.New(),
				StartDate:      time.Now(),
				EndDate:        time.Now().AddDate(0, 1, 0),
				AdjustmentType: domain.DemandAdjustmentFAD,
				Factor:         1.5,
				Reason:         "Valid reason",
			},
			setupMock: func(repo *MockDemandAdjustmentRepository, eventPub *MockEventPublisher) {
				repo.On("GetByID", mock.Anything, mock.Anything).
					Return(nil, domain.NewNotFoundError("not found"))
			},
			wantErr:     true,
			errContains: "demand adjustment not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockDemandAdjustmentRepository)
			mockEventPublisher := new(MockEventPublisher)
			tt.setupMock(mockRepo, mockEventPublisher)

			useCase := NewUpdateFADUseCase(mockRepo, mockEventPublisher)
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
