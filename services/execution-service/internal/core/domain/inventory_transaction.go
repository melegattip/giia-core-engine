package domain

import (
	"time"

	"github.com/google/uuid"
)

type InventoryTransaction struct {
	ID              uuid.UUID
	OrganizationID  uuid.UUID
	ProductID       uuid.UUID
	LocationID      uuid.UUID
	Type            TransactionType
	Quantity        float64
	UnitCost        float64
	ReferenceType   string
	ReferenceID     uuid.UUID
	Reason          string
	TransactionDate time.Time
	CreatedBy       uuid.UUID
	CreatedAt       time.Time
}

type TransactionType string

const (
	TransactionReceipt    TransactionType = "receipt"
	TransactionIssue      TransactionType = "issue"
	TransactionTransfer   TransactionType = "transfer"
	TransactionAdjustment TransactionType = "adjustment"
)

func (t TransactionType) IsValid() bool {
	switch t {
	case TransactionReceipt, TransactionIssue, TransactionTransfer, TransactionAdjustment:
		return true
	}
	return false
}

func NewInventoryTransaction(
	orgID, productID, locationID, createdBy uuid.UUID,
	txnType TransactionType,
	quantity, unitCost float64,
	referenceType string,
	referenceID uuid.UUID,
	reason string,
) (*InventoryTransaction, error) {
	if orgID == uuid.Nil {
		return nil, NewValidationError("organization_id is required")
	}
	if productID == uuid.Nil {
		return nil, NewValidationError("product_id is required")
	}
	if locationID == uuid.Nil {
		return nil, NewValidationError("location_id is required")
	}
	if createdBy == uuid.Nil {
		return nil, NewValidationError("created_by is required")
	}
	if !txnType.IsValid() {
		return nil, NewValidationError("invalid transaction type")
	}
	if quantity == 0 {
		return nil, NewValidationError("quantity cannot be zero")
	}
	if unitCost < 0 {
		return nil, NewValidationError("unit_cost cannot be negative")
	}

	now := time.Now()
	return &InventoryTransaction{
		ID:              uuid.New(),
		OrganizationID:  orgID,
		ProductID:       productID,
		LocationID:      locationID,
		Type:            txnType,
		Quantity:        quantity,
		UnitCost:        unitCost,
		ReferenceType:   referenceType,
		ReferenceID:     referenceID,
		Reason:          reason,
		TransactionDate: now,
		CreatedBy:       createdBy,
		CreatedAt:       now,
	}, nil
}