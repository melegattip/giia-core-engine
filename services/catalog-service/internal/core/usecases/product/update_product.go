package product

import (
	"context"

	"github.com/giia/giia-core-engine/pkg/errors"
	"github.com/giia/giia-core-engine/pkg/logger"
	"github.com/giia/giia-core-engine/services/catalog-service/internal/core/domain"
	"github.com/giia/giia-core-engine/services/catalog-service/internal/core/providers"
	"github.com/google/uuid"
)

type UpdateProductRequest struct {
	ID              uuid.UUID  `json:"id"`
	Name            string     `json:"name,omitempty" validate:"max=255"`
	Description     string     `json:"description"`
	Category        string     `json:"category,omitempty" validate:"max=100"`
	UnitOfMeasure   string     `json:"unit_of_measure,omitempty" validate:"max=50"`
	Status          string     `json:"status,omitempty"`
	BufferProfileID *uuid.UUID `json:"buffer_profile_id,omitempty"`
}

type UpdateProductUseCase struct {
	productRepo    providers.ProductRepository
	eventPublisher providers.EventPublisher
	logger         logger.Logger
}

func NewUpdateProductUseCase(
	productRepo providers.ProductRepository,
	eventPublisher providers.EventPublisher,
	logger logger.Logger,
) *UpdateProductUseCase {
	return &UpdateProductUseCase{
		productRepo:    productRepo,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

func (uc *UpdateProductUseCase) Execute(ctx context.Context, req *UpdateProductRequest) (*domain.Product, error) {
	if req == nil {
		return nil, errors.NewBadRequest("request cannot be nil")
	}

	if req.ID == uuid.Nil {
		return nil, errors.NewBadRequest("product ID is required")
	}

	product, err := uc.productRepo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		product.Name = req.Name
	}
	if req.Description != "" {
		product.Description = req.Description
	}
	if req.Category != "" {
		product.Category = req.Category
	}
	if req.UnitOfMeasure != "" {
		product.UnitOfMeasure = req.UnitOfMeasure
	}
	if req.Status != "" {
		product.Status = domain.ProductStatus(req.Status)
	}
	if req.BufferProfileID != nil {
		product.BufferProfileID = req.BufferProfileID
	}

	if err := product.Validate(); err != nil {
		uc.logger.Warn(ctx, "Product validation failed", logger.Tags{
			"product_id": product.ID.String(),
			"error":      err.Error(),
		})
		return nil, err
	}

	if err := uc.productRepo.Update(ctx, product); err != nil {
		uc.logger.Error(ctx, err, "Failed to update product", logger.Tags{
			"product_id": product.ID.String(),
		})
		return nil, err
	}

	uc.logger.Info(ctx, "Product updated successfully", logger.Tags{
		"product_id": product.ID.String(),
		"sku":        product.SKU,
	})

	if err := uc.eventPublisher.PublishProductUpdated(ctx, product); err != nil {
		uc.logger.Warn(ctx, "Failed to publish product updated event", logger.Tags{
			"product_id": product.ID.String(),
			"error":      err.Error(),
		})
	}

	return product, nil
}
