package purchase_order

import (
	"context"
	"time"

	"github.com/giia/giia-core-engine/services/execution-service/internal/core/domain"
	"github.com/giia/giia-core-engine/services/execution-service/internal/core/providers"
	"github.com/google/uuid"
)

type CreatePOUseCase struct {
	poRepo         providers.PurchaseOrderRepository
	catalogClient  providers.CatalogServiceClient
	eventPublisher providers.EventPublisher
}

func NewCreatePOUseCase(
	poRepo providers.PurchaseOrderRepository,
	catalogClient providers.CatalogServiceClient,
	publisher providers.EventPublisher,
) *CreatePOUseCase {
	return &CreatePOUseCase{
		poRepo:         poRepo,
		catalogClient:  catalogClient,
		eventPublisher: publisher,
	}
}

type CreatePOInput struct {
	OrganizationID      uuid.UUID
	PONumber            string
	SupplierID          uuid.UUID
	OrderDate           time.Time
	ExpectedArrivalDate time.Time
	LineItems           []domain.POLineItem
	CreatedBy           uuid.UUID
}

func (uc *CreatePOUseCase) Execute(ctx context.Context, input *CreatePOInput) (*domain.PurchaseOrder, error) {
	if input == nil {
		return nil, domain.NewValidationError("input cannot be nil")
	}

	existingPO, _ := uc.poRepo.GetByPONumber(ctx, input.PONumber, input.OrganizationID)
	if existingPO != nil {
		return nil, domain.NewConflictError("purchase order with this PO number already exists")
	}

	_, err := uc.catalogClient.GetSupplier(ctx, input.SupplierID)
	if err != nil {
		return nil, domain.NewValidationError("invalid supplier_id")
	}

	for _, item := range input.LineItems {
		_, err := uc.catalogClient.GetProduct(ctx, item.ProductID)
		if err != nil {
			return nil, domain.NewValidationError("invalid product_id: " + item.ProductID.String())
		}
	}

	po, err := domain.NewPurchaseOrder(
		input.OrganizationID,
		input.SupplierID,
		input.CreatedBy,
		input.PONumber,
		input.OrderDate,
		input.ExpectedArrivalDate,
		input.LineItems,
	)
	if err != nil {
		return nil, err
	}

	if err := uc.poRepo.Create(ctx, po); err != nil {
		return nil, err
	}

	uc.eventPublisher.PublishPOCreated(ctx, po)

	return po, nil
}