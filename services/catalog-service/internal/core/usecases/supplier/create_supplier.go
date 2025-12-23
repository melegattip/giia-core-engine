package supplier

import (
	"context"

	"github.com/melegattip/giia-core-engine/pkg/errors"
	"github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/catalog-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/catalog-service/internal/core/providers"
	"github.com/google/uuid"
)

type CreateSupplierRequest struct {
	Code              string `json:"code" validate:"required,max=100"`
	Name              string `json:"name" validate:"required,max=255"`
	LeadTimeDays      int    `json:"lead_time_days" validate:"gte=0"`
	ReliabilityRating int    `json:"reliability_rating" validate:"gte=0,lte=100"`
	ContactInfo       string `json:"contact_info,omitempty"`
}

type CreateSupplierUseCase struct {
	supplierRepo   providers.SupplierRepository
	eventPublisher providers.EventPublisher
	logger         logger.Logger
}

func NewCreateSupplierUseCase(
	supplierRepo providers.SupplierRepository,
	eventPublisher providers.EventPublisher,
	logger logger.Logger,
) *CreateSupplierUseCase {
	return &CreateSupplierUseCase{
		supplierRepo:   supplierRepo,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

func (uc *CreateSupplierUseCase) Execute(ctx context.Context, req *CreateSupplierRequest) (*domain.Supplier, error) {
	if req == nil {
		return nil, errors.NewBadRequest("request cannot be nil")
	}

	orgID, ok := ctx.Value("organization_id").(uuid.UUID)
	if !ok || orgID == uuid.Nil {
		return nil, errors.NewBadRequest("organization ID is required in context")
	}

	existing, _ := uc.supplierRepo.GetByCode(ctx, req.Code)
	if existing != nil {
		return nil, errors.NewConflict("supplier with this code already exists")
	}

	supplier := &domain.Supplier{
		Code:              req.Code,
		Name:              req.Name,
		LeadTimeDays:      req.LeadTimeDays,
		ReliabilityRating: req.ReliabilityRating,
		Status:            domain.SupplierStatusActive,
		OrganizationID:    orgID,
	}

	if err := supplier.Validate(); err != nil {
		uc.logger.Warn(ctx, "Supplier validation failed", logger.Tags{
			"code":  req.Code,
			"error": err.Error(),
		})
		return nil, err
	}

	if err := uc.supplierRepo.Create(ctx, supplier); err != nil {
		uc.logger.Error(ctx, err, "Failed to create supplier", logger.Tags{
			"code": req.Code,
		})
		return nil, err
	}

	uc.logger.Info(ctx, "Supplier created successfully", logger.Tags{
		"supplier_id": supplier.ID.String(),
		"code":        supplier.Code,
	})

	if err := uc.eventPublisher.PublishSupplierCreated(ctx, supplier); err != nil {
		uc.logger.Warn(ctx, "Failed to publish supplier created event", logger.Tags{
			"supplier_id": supplier.ID.String(),
			"error":       err.Error(),
		})
	}

	return supplier, nil
}
