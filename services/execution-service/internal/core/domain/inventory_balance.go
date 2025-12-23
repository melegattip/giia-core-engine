package domain

import (
	"time"

	"github.com/google/uuid"
)

type InventoryBalance struct {
	ID             uuid.UUID
	OrganizationID uuid.UUID
	ProductID      uuid.UUID
	LocationID     uuid.UUID
	OnHand         float64
	Reserved       float64
	Available      float64
	UpdatedAt      time.Time
}

func NewInventoryBalance(
	orgID, productID, locationID uuid.UUID,
) (*InventoryBalance, error) {
	if orgID == uuid.Nil {
		return nil, NewValidationError("organization_id is required")
	}
	if productID == uuid.Nil {
		return nil, NewValidationError("product_id is required")
	}
	if locationID == uuid.Nil {
		return nil, NewValidationError("location_id is required")
	}

	return &InventoryBalance{
		ID:             uuid.New(),
		OrganizationID: orgID,
		ProductID:      productID,
		LocationID:     locationID,
		OnHand:         0,
		Reserved:       0,
		Available:      0,
		UpdatedAt:      time.Now(),
	}, nil
}

func (ib *InventoryBalance) UpdateOnHand(quantity float64) {
	ib.OnHand += quantity
	ib.CalculateAvailable()
	ib.UpdatedAt = time.Now()
}

func (ib *InventoryBalance) UpdateReserved(quantity float64) error {
	newReserved := ib.Reserved + quantity
	if newReserved < 0 {
		return NewValidationError("reserved quantity cannot be negative")
	}
	if newReserved > ib.OnHand {
		return NewValidationError("cannot reserve more than on-hand quantity")
	}
	ib.Reserved = newReserved
	ib.CalculateAvailable()
	ib.UpdatedAt = time.Now()
	return nil
}

func (ib *InventoryBalance) CalculateAvailable() {
	ib.Available = ib.OnHand - ib.Reserved
	if ib.Available < 0 {
		ib.Available = 0
	}
}