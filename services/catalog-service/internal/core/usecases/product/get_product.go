package product

import (
	"context"

	"github.com/giia/giia-core-engine/pkg/errors"
	"github.com/giia/giia-core-engine/pkg/logger"
	"github.com/giia/giia-core-engine/services/catalog-service/internal/core/domain"
	"github.com/giia/giia-core-engine/services/catalog-service/internal/core/providers"
	"github.com/google/uuid"
)

type GetProductUseCase struct {
	productRepo providers.ProductRepository
	logger      logger.Logger
}

func NewGetProductUseCase(
	productRepo providers.ProductRepository,
	logger logger.Logger,
) *GetProductUseCase {
	return &GetProductUseCase{
		productRepo: productRepo,
		logger:      logger,
	}
}

func (uc *GetProductUseCase) Execute(ctx context.Context, id uuid.UUID, includeSuppliers bool) (*domain.Product, error) {
	if id == uuid.Nil {
		return nil, errors.NewBadRequest("product ID is required")
	}

	var product *domain.Product
	var err error

	if includeSuppliers {
		product, err = uc.productRepo.GetByIDWithSuppliers(ctx, id)
	} else {
		product, err = uc.productRepo.GetByID(ctx, id)
	}

	if err != nil {
		uc.logger.Warn(ctx, "Failed to retrieve product", logger.Tags{
			"product_id": id.String(),
			"error":      err.Error(),
		})
		return nil, err
	}

	return product, nil
}
