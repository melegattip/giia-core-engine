package alert_test

import (
	"context"
	"testing"

	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/usecases/alert"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAcknowledgeAlertUseCase_Execute_WithValidData_AcknowledgesAlert(t *testing.T) {
	givenAlertID := uuid.New()
	givenOrgID := uuid.New()
	givenUserID := uuid.New()
	givenAlert := &domain.Alert{
		ID:             givenAlertID,
		OrganizationID: givenOrgID,
		AlertType:      domain.AlertTypePODelayed,
		ResourceType:   "purchase_order",
		ResourceID:     uuid.New(),
	}

	mockAlertRepo := new(MockAlertRepository)

	mockAlertRepo.On("GetByID", mock.Anything, givenAlertID, givenOrgID).Return(givenAlert, nil)
	mockAlertRepo.On("Update", mock.Anything, mock.MatchedBy(func(alert *domain.Alert) bool {
		return alert.AcknowledgedAt != nil && alert.AcknowledgedBy != nil && *alert.AcknowledgedBy == givenUserID
	})).Return(nil)

	useCase := alert.NewAcknowledgeAlertUseCase(mockAlertRepo)

	input := &alert.AcknowledgeAlertInput{
		AlertID:        givenAlertID,
		OrganizationID: givenOrgID,
		UserID:         givenUserID,
	}

	result, err := useCase.Execute(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.AcknowledgedAt)
	assert.NotNil(t, result.AcknowledgedBy)
	assert.Equal(t, givenUserID, *result.AcknowledgedBy)
	mockAlertRepo.AssertExpectations(t)
}

func TestAcknowledgeAlertUseCase_Execute_WithNilInput_ReturnsError(t *testing.T) {
	mockAlertRepo := new(MockAlertRepository)

	useCase := alert.NewAcknowledgeAlertUseCase(mockAlertRepo)

	result, err := useCase.Execute(context.Background(), nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "input cannot be nil")
}

func TestAcknowledgeAlertUseCase_Execute_WithNilAlertID_ReturnsError(t *testing.T) {
	mockAlertRepo := new(MockAlertRepository)

	useCase := alert.NewAcknowledgeAlertUseCase(mockAlertRepo)

	input := &alert.AcknowledgeAlertInput{
		AlertID:        uuid.Nil,
		OrganizationID: uuid.New(),
		UserID:         uuid.New(),
	}

	result, err := useCase.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "alert_id is required")
}

func TestAcknowledgeAlertUseCase_Execute_WithNilOrganizationID_ReturnsError(t *testing.T) {
	mockAlertRepo := new(MockAlertRepository)

	useCase := alert.NewAcknowledgeAlertUseCase(mockAlertRepo)

	input := &alert.AcknowledgeAlertInput{
		AlertID:        uuid.New(),
		OrganizationID: uuid.Nil,
		UserID:         uuid.New(),
	}

	result, err := useCase.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "organization_id is required")
}

func TestAcknowledgeAlertUseCase_Execute_WithNilUserID_ReturnsError(t *testing.T) {
	mockAlertRepo := new(MockAlertRepository)

	useCase := alert.NewAcknowledgeAlertUseCase(mockAlertRepo)

	input := &alert.AcknowledgeAlertInput{
		AlertID:        uuid.New(),
		OrganizationID: uuid.New(),
		UserID:         uuid.Nil,
	}

	result, err := useCase.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user_id is required")
}

func TestAcknowledgeAlertUseCase_Execute_WhenAlertNotFound_ReturnsError(t *testing.T) {
	givenAlertID := uuid.New()
	givenOrgID := uuid.New()

	mockAlertRepo := new(MockAlertRepository)

	mockAlertRepo.On("GetByID", mock.Anything, givenAlertID, givenOrgID).Return((*domain.Alert)(nil), assert.AnError)

	useCase := alert.NewAcknowledgeAlertUseCase(mockAlertRepo)

	input := &alert.AcknowledgeAlertInput{
		AlertID:        givenAlertID,
		OrganizationID: givenOrgID,
		UserID:         uuid.New(),
	}

	result, err := useCase.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, result)
	mockAlertRepo.AssertExpectations(t)
}

func TestAcknowledgeAlertUseCase_Execute_WhenAlertAlreadyAcknowledged_ReturnsError(t *testing.T) {
	givenAlertID := uuid.New()
	givenOrgID := uuid.New()
	givenUserID := uuid.New()
	givenAlert := &domain.Alert{
		ID:             givenAlertID,
		OrganizationID: givenOrgID,
		AlertType:      domain.AlertTypePODelayed,
	}
	givenAlert.Acknowledge(uuid.New())

	mockAlertRepo := new(MockAlertRepository)

	mockAlertRepo.On("GetByID", mock.Anything, givenAlertID, givenOrgID).Return(givenAlert, nil)

	useCase := alert.NewAcknowledgeAlertUseCase(mockAlertRepo)

	input := &alert.AcknowledgeAlertInput{
		AlertID:        givenAlertID,
		OrganizationID: givenOrgID,
		UserID:         givenUserID,
	}

	result, err := useCase.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "alert already acknowledged")
	mockAlertRepo.AssertExpectations(t)
}

func TestAcknowledgeAlertUseCase_Execute_WhenUpdateFails_ReturnsError(t *testing.T) {
	givenAlertID := uuid.New()
	givenOrgID := uuid.New()
	givenUserID := uuid.New()
	givenAlert := &domain.Alert{
		ID:             givenAlertID,
		OrganizationID: givenOrgID,
		AlertType:      domain.AlertTypePODelayed,
	}

	mockAlertRepo := new(MockAlertRepository)

	mockAlertRepo.On("GetByID", mock.Anything, givenAlertID, givenOrgID).Return(givenAlert, nil)
	mockAlertRepo.On("Update", mock.Anything, mock.Anything).Return(assert.AnError)

	useCase := alert.NewAcknowledgeAlertUseCase(mockAlertRepo)

	input := &alert.AcknowledgeAlertInput{
		AlertID:        givenAlertID,
		OrganizationID: givenOrgID,
		UserID:         givenUserID,
	}

	result, err := useCase.Execute(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, result)
	mockAlertRepo.AssertExpectations(t)
}
