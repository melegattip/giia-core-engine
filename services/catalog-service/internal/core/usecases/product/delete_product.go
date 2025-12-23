package product

import (
	"context"

	"github.com/melegattip/giia-core-engine/pkg/errors"
	"github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/catalog-service/internal/core/providers"
	"github.com/google/uuid"
)

type DeleteProductUseCase struct {
	productRepo    providers.ProductRepository
	eventPublisher providers.EventPublisher
	logger         logger.Logger
}

func NewDeleteProductUseCase(
	productRepo providers.ProductRepository,
	eventPublisher providers.EventPublisher,
	logger logger.Logger,
) *DeleteProductUseCase {
	return &DeleteProductUseCase{
		productRepo:    productRepo,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

func (uc *DeleteProductUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return errors.NewBadRequest("product ID is required")
	}

	product, err := uc.productRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := uc.productRepo.Delete(ctx, id); err != nil {
		uc.logger.Error(ctx, err, "Failed to delete product", logger.Tags{
			"product_id": id.String(),
		})
		return err
	}

	uc.logger.Info(ctx, "Product deleted successfully", logger.Tags{
		"product_id": id.String(),
		"sku":        product.SKU,
	})

	if err := uc.eventPublisher.PublishProductDeleted(ctx, product); err != nil {
		uc.logger.Warn(ctx, "Failed to publish product deleted event", logger.Tags{
			"product_id": product.ID.String(),
			"error":      err.Error(),
		})
	}

	return nil
}
