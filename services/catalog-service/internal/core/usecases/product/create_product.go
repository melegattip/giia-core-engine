package product

import (
	"context"

	"github.com/giia/giia-core-engine/pkg/logger"
	"github.com/giia/giia-core-engine/services/catalog-service/internal/core/domain"
	"github.com/giia/giia-core-engine/services/catalog-service/internal/core/providers"
	"github.com/google/uuid"
)

type CreateProductRequest struct {
	SKU             string     `json:"sku" validate:"required,max=100"`
	Name            string     `json:"name" validate:"required,max=255"`
	Description     string     `json:"description"`
	Category        string     `json:"category,omitempty" validate:"max=100"`
	UnitOfMeasure   string     `json:"unit_of_measure" validate:"required,max=50"`
	BufferProfileID *uuid.UUID `json:"buffer_profile_id,omitempty"`
}

type CreateProductUseCase struct {
	productRepo    providers.ProductRepository
	eventPublisher providers.EventPublisher
	logger         logger.Logger
}

func NewCreateProductUseCase(
	productRepo providers.ProductRepository,
	eventPublisher providers.EventPublisher,
	logger logger.Logger,
) *CreateProductUseCase {
	return &CreateProductUseCase{
		productRepo:    productRepo,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

func (uc *CreateProductUseCase) Execute(ctx context.Context, req *CreateProductRequest) (*domain.Product, error) {
	if req == nil {
		return nil, domain.NewValidationError("request cannot be nil")
	}

	orgID, ok := ctx.Value("organization_id").(uuid.UUID)
	if !ok || orgID == uuid.Nil {
		return nil, domain.NewValidationError("organization ID is required in context")
	}

	product := &domain.Product{
		SKU:             req.SKU,
		Name:            req.Name,
		Description:     req.Description,
		Category:        req.Category,
		UnitOfMeasure:   req.UnitOfMeasure,
		Status:          domain.ProductStatusActive,
		BufferProfileID: req.BufferProfileID,
		OrganizationID:  orgID,
	}

	if err := product.Validate(); err != nil {
		uc.logger.Warn(ctx, "Product validation failed", logger.Tags{
			"sku":   req.SKU,
			"error": err.Error(),
		})
		return nil, err
	}

	if err := uc.productRepo.Create(ctx, product); err != nil {
		uc.logger.Error(ctx, err, "Failed to create product", logger.Tags{
			"sku": req.SKU,
		})
		return nil, err
	}

	uc.logger.Info(ctx, "Product created successfully", logger.Tags{
		"product_id": product.ID.String(),
		"sku":        product.SKU,
	})

	if err := uc.eventPublisher.PublishProductCreated(ctx, product); err != nil {
		uc.logger.Warn(ctx, "Failed to publish product created event", logger.Tags{
			"product_id": product.ID.String(),
			"error":      err.Error(),
		})
	}

	return product, nil
}
