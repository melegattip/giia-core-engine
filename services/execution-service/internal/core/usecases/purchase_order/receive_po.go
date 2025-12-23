package purchase_order

import (
	"context"

	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/providers"
	"github.com/google/uuid"
)

type ReceivePOUseCase struct {
	poRepo               providers.PurchaseOrderRepository
	inventoryTxnRepo     providers.InventoryTransactionRepository
	inventoryBalanceRepo providers.InventoryBalanceRepository
	ddmrpClient          providers.DDMRPServiceClient
	eventPublisher       providers.EventPublisher
}

func NewReceivePOUseCase(
	poRepo providers.PurchaseOrderRepository,
	inventoryTxnRepo providers.InventoryTransactionRepository,
	inventoryBalanceRepo providers.InventoryBalanceRepository,
	ddmrpClient providers.DDMRPServiceClient,
	publisher providers.EventPublisher,
) *ReceivePOUseCase {
	return &ReceivePOUseCase{
		poRepo:               poRepo,
		inventoryTxnRepo:     inventoryTxnRepo,
		inventoryBalanceRepo: inventoryBalanceRepo,
		ddmrpClient:          ddmrpClient,
		eventPublisher:       publisher,
	}
}

type ReceiveLineItem struct {
	LineItemID  uuid.UUID
	ReceivedQty float64
}

type ReceivePOInput struct {
	POID           uuid.UUID
	OrganizationID uuid.UUID
	LocationID     uuid.UUID
	Receipts       []ReceiveLineItem
	ReceivedBy     uuid.UUID
}

func (uc *ReceivePOUseCase) Execute(ctx context.Context, input *ReceivePOInput) (*domain.PurchaseOrder, error) {
	if input == nil {
		return nil, domain.NewValidationError("input cannot be nil")
	}
	if input.POID == uuid.Nil {
		return nil, domain.NewValidationError("po_id is required")
	}
	if input.OrganizationID == uuid.Nil {
		return nil, domain.NewValidationError("organization_id is required")
	}
	if input.LocationID == uuid.Nil {
		return nil, domain.NewValidationError("location_id is required")
	}
	if input.ReceivedBy == uuid.Nil {
		return nil, domain.NewValidationError("received_by is required")
	}
	if len(input.Receipts) == 0 {
		return nil, domain.NewValidationError("at least one receipt is required")
	}

	po, err := uc.poRepo.GetByID(ctx, input.POID, input.OrganizationID)
	if err != nil {
		return nil, err
	}

	if po.Status != domain.POStatusConfirmed && po.Status != domain.POStatusPartial {
		return nil, domain.NewValidationError("can only receive confirmed or partially received purchase orders")
	}

	for _, receipt := range input.Receipts {
		if receipt.ReceivedQty <= 0 {
			return nil, domain.NewValidationError("received quantity must be greater than zero")
		}

		var lineItem *domain.POLineItem
		for i := range po.LineItems {
			if po.LineItems[i].ID == receipt.LineItemID {
				lineItem = &po.LineItems[i]
				break
			}
		}

		if lineItem == nil {
			return nil, domain.NewValidationError("invalid line_item_id")
		}

		if lineItem.ReceivedQty+receipt.ReceivedQty > lineItem.Quantity {
			return nil, domain.NewValidationError("received quantity exceeds ordered quantity")
		}

		txn, err := domain.NewInventoryTransaction(
			input.OrganizationID,
			lineItem.ProductID,
			input.LocationID,
			input.ReceivedBy,
			domain.TransactionReceipt,
			receipt.ReceivedQty,
			lineItem.UnitCost,
			"purchase_order",
			po.ID,
			"PO receipt: "+po.PONumber,
		)
		if err != nil {
			return nil, err
		}

		if err := uc.inventoryTxnRepo.Create(ctx, txn); err != nil {
			return nil, err
		}

		if err := uc.inventoryBalanceRepo.UpdateOnHand(ctx, input.OrganizationID, lineItem.ProductID, input.LocationID, receipt.ReceivedQty); err != nil {
			return nil, err
		}

		lineItem.ReceivedQty += receipt.ReceivedQty

		uc.ddmrpClient.UpdateNetFlowPosition(ctx, input.OrganizationID, lineItem.ProductID)
		uc.eventPublisher.PublishInventoryUpdated(ctx, txn)
	}

	po.UpdateReceiptStatus()

	if err := uc.poRepo.Update(ctx, po); err != nil {
		return nil, err
	}

	uc.eventPublisher.PublishPOReceived(ctx, po)

	return po, nil
}