package supplier

import (
	"context"

	"github.com/melegattip/giia-core-engine/pkg/errors"
	"github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/catalog-service/internal/core/providers"
	"github.com/google/uuid"
)

type DeleteSupplierUseCase struct {
	supplierRepo   providers.SupplierRepository
	eventPublisher providers.EventPublisher
	logger         logger.Logger
}

func NewDeleteSupplierUseCase(
	supplierRepo providers.SupplierRepository,
	eventPublisher providers.EventPublisher,
	logger logger.Logger,
) *DeleteSupplierUseCase {
	return &DeleteSupplierUseCase{
		supplierRepo:   supplierRepo,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

func (uc *DeleteSupplierUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return errors.NewBadRequest("supplier ID is required")
	}

	orgID, ok := ctx.Value("organization_id").(uuid.UUID)
	if !ok || orgID == uuid.Nil {
		return errors.NewBadRequest("organization ID is required in context")
	}

	supplier, err := uc.supplierRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := uc.supplierRepo.Delete(ctx, id); err != nil {
		uc.logger.Error(ctx, err, "Failed to delete supplier", logger.Tags{
			"supplier_id": id.String(),
		})
		return err
	}

	uc.logger.Info(ctx, "Supplier deleted successfully", logger.Tags{
		"supplier_id": id.String(),
	})

	if err := uc.eventPublisher.PublishSupplierDeleted(ctx, supplier); err != nil {
		uc.logger.Warn(ctx, "Failed to publish supplier deleted event", logger.Tags{
			"supplier_id": id.String(),
			"error":       err.Error(),
		})
	}

	return nil
}
