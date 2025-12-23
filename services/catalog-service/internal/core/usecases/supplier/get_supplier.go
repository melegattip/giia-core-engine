package supplier

import (
	"context"

	"github.com/melegattip/giia-core-engine/pkg/errors"
	"github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/catalog-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/catalog-service/internal/core/providers"
	"github.com/google/uuid"
)

type GetSupplierUseCase struct {
	supplierRepo providers.SupplierRepository
	logger       logger.Logger
}

func NewGetSupplierUseCase(
	supplierRepo providers.SupplierRepository,
	logger logger.Logger,
) *GetSupplierUseCase {
	return &GetSupplierUseCase{
		supplierRepo: supplierRepo,
		logger:       logger,
	}
}

func (uc *GetSupplierUseCase) Execute(ctx context.Context, id uuid.UUID) (*domain.Supplier, error) {
	if id == uuid.Nil {
		return nil, errors.NewBadRequest("supplier ID is required")
	}

	orgID, ok := ctx.Value("organization_id").(uuid.UUID)
	if !ok || orgID == uuid.Nil {
		return nil, errors.NewBadRequest("organization ID is required in context")
	}

	supplier, err := uc.supplierRepo.GetByID(ctx, id)
	if err != nil {
		uc.logger.Warn(ctx, "Supplier not found", logger.Tags{
			"supplier_id": id.String(),
		})
		return nil, err
	}

	return supplier, nil
}
