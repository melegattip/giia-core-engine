package alert

import (
	"context"

	"github.com/giia/giia-core-engine/services/execution-service/internal/core/domain"
	"github.com/giia/giia-core-engine/services/execution-service/internal/core/providers"
	"github.com/google/uuid"
)

type AcknowledgeAlertUseCase struct {
	alertRepo providers.AlertRepository
}

func NewAcknowledgeAlertUseCase(
	alertRepo providers.AlertRepository,
) *AcknowledgeAlertUseCase {
	return &AcknowledgeAlertUseCase{
		alertRepo: alertRepo,
	}
}

type AcknowledgeAlertInput struct {
	AlertID        uuid.UUID
	OrganizationID uuid.UUID
	UserID         uuid.UUID
}

func (uc *AcknowledgeAlertUseCase) Execute(ctx context.Context, input *AcknowledgeAlertInput) (*domain.Alert, error) {
	if input == nil {
		return nil, domain.NewValidationError("input cannot be nil")
	}
	if input.AlertID == uuid.Nil {
		return nil, domain.NewValidationError("alert_id is required")
	}
	if input.OrganizationID == uuid.Nil {
		return nil, domain.NewValidationError("organization_id is required")
	}
	if input.UserID == uuid.Nil {
		return nil, domain.NewValidationError("user_id is required")
	}

	alert, err := uc.alertRepo.GetByID(ctx, input.AlertID, input.OrganizationID)
	if err != nil {
		return nil, err
	}

	if err := alert.Acknowledge(input.UserID); err != nil {
		return nil, err
	}

	if err := uc.alertRepo.Update(ctx, alert); err != nil {
		return nil, err
	}

	return alert, nil
}