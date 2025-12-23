package nfp

import (
	"context"

	"github.com/giia/giia-core-engine/pkg/errors"
	"github.com/giia/giia-core-engine/services/ddmrp-engine-service/internal/core/domain"
	"github.com/giia/giia-core-engine/services/ddmrp-engine-service/internal/core/providers"
	"github.com/google/uuid"
)

type UpdateNFPUseCase struct {
	bufferRepo     providers.BufferRepository
	eventPublisher providers.EventPublisher
}

func NewUpdateNFPUseCase(
	bufferRepo providers.BufferRepository,
	eventPublisher providers.EventPublisher,
) *UpdateNFPUseCase {
	return &UpdateNFPUseCase{
		bufferRepo:     bufferRepo,
		eventPublisher: eventPublisher,
	}
}

type UpdateNFPInput struct {
	ProductID       uuid.UUID
	OrganizationID  uuid.UUID
	OnHand          float64
	OnOrder         float64
	QualifiedDemand float64
}

func (uc *UpdateNFPUseCase) Execute(ctx context.Context, input UpdateNFPInput) (*domain.Buffer, error) {
	if input.ProductID == uuid.Nil {
		return nil, errors.NewBadRequest("product_id is required")
	}
	if input.OrganizationID == uuid.Nil {
		return nil, errors.NewBadRequest("organization_id is required")
	}

	buffer, err := uc.bufferRepo.GetByProduct(ctx, input.ProductID, input.OrganizationID)
	if err != nil {
		return nil, errors.NewNotFound("buffer not found for product")
	}

	oldZone := buffer.Zone

	buffer.OnHand = input.OnHand
	buffer.OnOrder = input.OnOrder
	buffer.QualifiedDemand = input.QualifiedDemand

	buffer.DetermineZone()

	if err := uc.bufferRepo.Save(ctx, buffer); err != nil {
		return nil, errors.NewInternalServerError("failed to update buffer NFP")
	}

	if oldZone != buffer.Zone {
		if err := uc.eventPublisher.PublishBufferStatusChanged(ctx, buffer, oldZone); err != nil {
			return nil, errors.NewInternalServerError("failed to publish status changed event")
		}
	}

	if buffer.AlertLevel == domain.AlertReplenish || buffer.AlertLevel == domain.AlertCritical {
		if err := uc.eventPublisher.PublishBufferAlertTriggered(ctx, buffer); err != nil {
			return nil, errors.NewInternalServerError("failed to publish alert event")
		}
	}

	return buffer, nil
}
