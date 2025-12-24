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

func TestDeleteFAD_Execute_Success(t *testing.T) {
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
		Reason:         "Test reason",
		CreatedBy:      uuid.New(),
	}

	mockRepo.On("GetByID", mock.Anything, fadID).Return(existingFAD, nil)
	mockRepo.On("Delete", mock.Anything, fadID).Return(nil)
	mockEventPublisher.On("PublishFADDeleted", mock.Anything, fadID.String()).Return(nil)

	useCase := NewDeleteFADUseCase(mockRepo, mockEventPublisher)
	err := useCase.Execute(context.Background(), fadID)

	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockEventPublisher.AssertExpectations(t)
}

func TestDeleteFAD_Execute_NilID(t *testing.T) {
	mockRepo := new(MockDemandAdjustmentRepository)
	mockEventPublisher := new(MockEventPublisher)

	useCase := NewDeleteFADUseCase(mockRepo, mockEventPublisher)
	err := useCase.Execute(context.Background(), uuid.Nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "id is required")
}

func TestDeleteFAD_Execute_FADNotFound(t *testing.T) {
	mockRepo := new(MockDemandAdjustmentRepository)
	mockEventPublisher := new(MockEventPublisher)

	fadID := uuid.New()

	mockRepo.On("GetByID", mock.Anything, fadID).
		Return(nil, domain.NewNotFoundError("demand adjustment not found"))

	useCase := NewDeleteFADUseCase(mockRepo, mockEventPublisher)
	err := useCase.Execute(context.Background(), fadID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "demand adjustment not found")

	mockRepo.AssertExpectations(t)
}

func TestDeleteFAD_Execute_DeleteError(t *testing.T) {
	mockRepo := new(MockDemandAdjustmentRepository)
	mockEventPublisher := new(MockEventPublisher)

	fadID := uuid.New()
	existingFAD := &domain.DemandAdjustment{
		ID:             fadID,
		ProductID:      uuid.New(),
		OrganizationID: uuid.New(),
	}

	mockRepo.On("GetByID", mock.Anything, fadID).Return(existingFAD, nil)
	mockRepo.On("Delete", mock.Anything, fadID).
		Return(domain.NewInternalError("database error"))

	useCase := NewDeleteFADUseCase(mockRepo, mockEventPublisher)
	err := useCase.Execute(context.Background(), fadID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete demand adjustment")

	mockRepo.AssertExpectations(t)
}

func TestDeleteFAD_Execute_EventPublishError(t *testing.T) {
	mockRepo := new(MockDemandAdjustmentRepository)
	mockEventPublisher := new(MockEventPublisher)

	fadID := uuid.New()
	existingFAD := &domain.DemandAdjustment{
		ID:             fadID,
		ProductID:      uuid.New(),
		OrganizationID: uuid.New(),
	}

	mockRepo.On("GetByID", mock.Anything, fadID).Return(existingFAD, nil)
	mockRepo.On("Delete", mock.Anything, fadID).Return(nil)
	mockEventPublisher.On("PublishFADDeleted", mock.Anything, fadID.String()).
		Return(domain.NewInternalError("event publish error"))

	useCase := NewDeleteFADUseCase(mockRepo, mockEventPublisher)
	err := useCase.Execute(context.Background(), fadID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to publish FAD deleted event")

	mockRepo.AssertExpectations(t)
	mockEventPublisher.AssertExpectations(t)
}

func TestDeleteFAD_Execute_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		fadID       uuid.UUID
		setupMock   func(repo *MockDemandAdjustmentRepository, eventPub *MockEventPublisher)
		wantErr     bool
		errContains string
	}{
		{
			name:  "success",
			fadID: uuid.New(),
			setupMock: func(repo *MockDemandAdjustmentRepository, eventPub *MockEventPublisher) {
				repo.On("GetByID", mock.Anything, mock.Anything).
					Return(&domain.DemandAdjustment{ID: uuid.New()}, nil)
				repo.On("Delete", mock.Anything, mock.Anything).Return(nil)
				eventPub.On("PublishFADDeleted", mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: false,
		},
		{
			name:        "nil id",
			fadID:       uuid.Nil,
			setupMock:   func(repo *MockDemandAdjustmentRepository, eventPub *MockEventPublisher) {},
			wantErr:     true,
			errContains: "id is required",
		},
		{
			name:  "not found",
			fadID: uuid.New(),
			setupMock: func(repo *MockDemandAdjustmentRepository, eventPub *MockEventPublisher) {
				repo.On("GetByID", mock.Anything, mock.Anything).
					Return(nil, domain.NewNotFoundError("not found"))
			},
			wantErr:     true,
			errContains: "demand adjustment not found",
		},
		{
			name:  "delete error",
			fadID: uuid.New(),
			setupMock: func(repo *MockDemandAdjustmentRepository, eventPub *MockEventPublisher) {
				repo.On("GetByID", mock.Anything, mock.Anything).
					Return(&domain.DemandAdjustment{ID: uuid.New()}, nil)
				repo.On("Delete", mock.Anything, mock.Anything).
					Return(domain.NewInternalError("db error"))
			},
			wantErr:     true,
			errContains: "failed to delete demand adjustment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockDemandAdjustmentRepository)
			mockEventPublisher := new(MockEventPublisher)
			tt.setupMock(mockRepo, mockEventPublisher)

			useCase := NewDeleteFADUseCase(mockRepo, mockEventPublisher)
			err := useCase.Execute(context.Background(), tt.fadID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
			mockEventPublisher.AssertExpectations(t)
		})
	}
}
