package sales_order

import (
	"context"

	"github.com/giia/giia-core-engine/services/execution-service/internal/core/domain"
	"github.com/giia/giia-core-engine/services/execution-service/internal/core/providers"
	"github.com/google/uuid"
)

type IssueDeliveryNoteUseCase struct {
	soRepo               providers.SalesOrderRepository
	inventoryTxnRepo     providers.InventoryTransactionRepository
	inventoryBalanceRepo providers.InventoryBalanceRepository
	ddmrpClient          providers.DDMRPServiceClient
	eventPublisher       providers.EventPublisher
}

func NewIssueDeliveryNoteUseCase(
	soRepo providers.SalesOrderRepository,
	inventoryTxnRepo providers.InventoryTransactionRepository,
	inventoryBalanceRepo providers.InventoryBalanceRepository,
	ddmrpClient providers.DDMRPServiceClient,
	publisher providers.EventPublisher,
) *IssueDeliveryNoteUseCase {
	return &IssueDeliveryNoteUseCase{
		soRepo:               soRepo,
		inventoryTxnRepo:     inventoryTxnRepo,
		inventoryBalanceRepo: inventoryBalanceRepo,
		ddmrpClient:          ddmrpClient,
		eventPublisher:       publisher,
	}
}

type IssueDeliveryNoteInput struct {
	SOID               uuid.UUID
	OrganizationID     uuid.UUID
	LocationID         uuid.UUID
	DeliveryNoteNumber string
	IssuedBy           uuid.UUID
}

func (uc *IssueDeliveryNoteUseCase) Execute(ctx context.Context, input *IssueDeliveryNoteInput) (*domain.SalesOrder, error) {
	if input == nil {
		return nil, domain.NewValidationError("input cannot be nil")
	}
	if input.SOID == uuid.Nil {
		return nil, domain.NewValidationError("so_id is required")
	}
	if input.OrganizationID == uuid.Nil {
		return nil, domain.NewValidationError("organization_id is required")
	}
	if input.LocationID == uuid.Nil {
		return nil, domain.NewValidationError("location_id is required")
	}
	if input.DeliveryNoteNumber == "" {
		return nil, domain.NewValidationError("delivery_note_number is required")
	}
	if input.IssuedBy == uuid.Nil {
		return nil, domain.NewValidationError("issued_by is required")
	}

	so, err := uc.soRepo.GetByID(ctx, input.SOID, input.OrganizationID)
	if err != nil {
		return nil, err
	}

	if err := so.IssueDeliveryNote(input.DeliveryNoteNumber); err != nil {
		return nil, err
	}

	for _, lineItem := range so.LineItems {
		txn, err := domain.NewInventoryTransaction(
			input.OrganizationID,
			lineItem.ProductID,
			input.LocationID,
			input.IssuedBy,
			domain.TransactionIssue,
			-lineItem.Quantity,
			0,
			"sales_order",
			so.ID,
			"SO delivery: "+so.SONumber,
		)
		if err != nil {
			return nil, err
		}

		if err := uc.inventoryTxnRepo.Create(ctx, txn); err != nil {
			return nil, err
		}

		if err := uc.inventoryBalanceRepo.UpdateOnHand(ctx, input.OrganizationID, lineItem.ProductID, input.LocationID, -lineItem.Quantity); err != nil {
			return nil, err
		}

		uc.ddmrpClient.UpdateNetFlowPosition(ctx, input.OrganizationID, lineItem.ProductID)
		uc.eventPublisher.PublishInventoryUpdated(ctx, txn)
	}

	if err := uc.soRepo.Update(ctx, so); err != nil {
		return nil, err
	}

	uc.eventPublisher.PublishDeliveryNoteIssued(ctx, so)

	return so, nil
}