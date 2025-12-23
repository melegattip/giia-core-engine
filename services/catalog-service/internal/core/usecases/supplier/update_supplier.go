package supplier

import (
	"context"

	"github.com/giia/giia-core-engine/pkg/errors"
	"github.com/giia/giia-core-engine/pkg/logger"
	"github.com/giia/giia-core-engine/services/catalog-service/internal/core/domain"
	"github.com/giia/giia-core-engine/services/catalog-service/internal/core/providers"
	"github.com/google/uuid"
)

type UpdateSupplierRequest struct {
	ID                uuid.UUID `json:"id" validate:"required"`
	Name              string    `json:"name" validate:"required,max=255"`
	LeadTimeDays      int       `json:"lead_time_days" validate:"gte=0"`
	ReliabilityRating int       `json:"reliability_rating" validate:"gte=0,lte=100"`
	ContactInfo       string    `json:"contact_info,omitempty"`
	Status            string    `json:"status,omitempty"`
}

type UpdateSupplierUseCase struct {
	supplierRepo   providers.SupplierRepository
	eventPublisher providers.EventPublisher
	logger         logger.Logger
}

func NewUpdateSupplierUseCase(
	supplierRepo providers.SupplierRepository,
	eventPublisher providers.EventPublisher,
	logger logger.Logger,
) *UpdateSupplierUseCase {
	return &UpdateSupplierUseCase{
		supplierRepo:   supplierRepo,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

func (uc *UpdateSupplierUseCase) Execute(ctx context.Context, req *UpdateSupplierRequest) (*domain.Supplier, error) {
	if req == nil {
		return nil, errors.NewBadRequest("request cannot be nil")
	}

	if req.ID == uuid.Nil {
		return nil, errors.NewBadRequest("supplier ID is required")
	}

	orgID, ok := ctx.Value("organization_id").(uuid.UUID)
	if !ok || orgID == uuid.Nil {
		return nil, errors.NewBadRequest("organization ID is required in context")
	}

	existingSupplier, err := uc.supplierRepo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	existingSupplier.Name = req.Name
	existingSupplier.LeadTimeDays = req.LeadTimeDays
	existingSupplier.ReliabilityRating = req.ReliabilityRating

	if req.Status != "" {
		existingSupplier.Status = domain.SupplierStatus(req.Status)
	}

	if err := existingSupplier.Validate(); err != nil {
		uc.logger.Warn(ctx, "Supplier validation failed", logger.Tags{
			"supplier_id": req.ID.String(),
			"error":       err.Error(),
		})
		return nil, err
	}

	if err := uc.supplierRepo.Update(ctx, existingSupplier); err != nil {
		uc.logger.Error(ctx, err, "Failed to update supplier", logger.Tags{
			"supplier_id": req.ID.String(),
		})
		return nil, err
	}

	uc.logger.Info(ctx, "Supplier updated successfully", logger.Tags{
		"supplier_id": existingSupplier.ID.String(),
	})

	if err := uc.eventPublisher.PublishSupplierUpdated(ctx, existingSupplier); err != nil {
		uc.logger.Warn(ctx, "Failed to publish supplier updated event", logger.Tags{
			"supplier_id": existingSupplier.ID.String(),
			"error":       err.Error(),
		})
	}

	return existingSupplier, nil
}
