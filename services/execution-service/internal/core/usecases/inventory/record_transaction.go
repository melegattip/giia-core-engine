package inventory

import (
	"context"

	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/providers"
	"github.com/google/uuid"
)

type RecordTransactionUseCase struct {
	inventoryTxnRepo     providers.InventoryTransactionRepository
	inventoryBalanceRepo providers.InventoryBalanceRepository
	ddmrpClient          providers.DDMRPServiceClient
	eventPublisher       providers.EventPublisher
}

func NewRecordTransactionUseCase(
	inventoryTxnRepo providers.InventoryTransactionRepository,
	inventoryBalanceRepo providers.InventoryBalanceRepository,
	ddmrpClient providers.DDMRPServiceClient,
	publisher providers.EventPublisher,
) *RecordTransactionUseCase {
	return &RecordTransactionUseCase{
		inventoryTxnRepo:     inventoryTxnRepo,
		inventoryBalanceRepo: inventoryBalanceRepo,
		ddmrpClient:          ddmrpClient,
		eventPublisher:       publisher,
	}
}

type RecordTransactionInput struct {
	OrganizationID uuid.UUID
	ProductID      uuid.UUID
	LocationID     uuid.UUID
	Type           domain.TransactionType
	Quantity       float64
	UnitCost       float64
	ReferenceType  string
	ReferenceID    uuid.UUID
	Reason         string
	CreatedBy      uuid.UUID
}

func (uc *RecordTransactionUseCase) Execute(ctx context.Context, input *RecordTransactionInput) (*domain.InventoryTransaction, error) {
	if input == nil {
		return nil, domain.NewValidationError("input cannot be nil")
	}

	txn, err := domain.NewInventoryTransaction(
		input.OrganizationID,
		input.ProductID,
		input.LocationID,
		input.CreatedBy,
		input.Type,
		input.Quantity,
		input.UnitCost,
		input.ReferenceType,
		input.ReferenceID,
		input.Reason,
	)
	if err != nil {
		return nil, err
	}

	if err := uc.inventoryTxnRepo.Create(ctx, txn); err != nil {
		return nil, err
	}

	if err := uc.inventoryBalanceRepo.UpdateOnHand(ctx, input.OrganizationID, input.ProductID, input.LocationID, input.Quantity); err != nil {
		return nil, err
	}

	uc.ddmrpClient.UpdateNetFlowPosition(ctx, input.OrganizationID, input.ProductID)
	uc.eventPublisher.PublishInventoryUpdated(ctx, txn)

	return txn, nil
}