package domain

import (
	"time"

	"github.com/google/uuid"
)

type PurchaseOrder struct {
	ID                  uuid.UUID
	OrganizationID      uuid.UUID
	PONumber            string
	SupplierID          uuid.UUID
	Status              POStatus
	OrderDate           time.Time
	ExpectedArrivalDate time.Time
	ActualArrivalDate   *time.Time
	DelayDays           int
	IsDelayed           bool
	TotalAmount         float64
	LineItems           []POLineItem
	CreatedBy           uuid.UUID
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type POLineItem struct {
	ID              uuid.UUID
	PurchaseOrderID uuid.UUID
	ProductID       uuid.UUID
	Quantity        float64
	ReceivedQty     float64
	UnitCost        float64
	LineTotal       float64
}

type POStatus string

const (
	POStatusDraft     POStatus = "draft"
	POStatusPending   POStatus = "pending"
	POStatusConfirmed POStatus = "confirmed"
	POStatusPartial   POStatus = "partial"
	POStatusReceived  POStatus = "received"
	POStatusClosed    POStatus = "closed"
	POStatusCancelled POStatus = "cancelled"
)

func (s POStatus) IsValid() bool {
	switch s {
	case POStatusDraft, POStatusPending, POStatusConfirmed, POStatusPartial,
		POStatusReceived, POStatusClosed, POStatusCancelled:
		return true
	}
	return false
}

func NewPurchaseOrder(
	orgID, supplierID, createdBy uuid.UUID,
	poNumber string,
	orderDate, expectedArrivalDate time.Time,
	lineItems []POLineItem,
) (*PurchaseOrder, error) {
	if orgID == uuid.Nil {
		return nil, NewValidationError("organization_id is required")
	}
	if supplierID == uuid.Nil {
		return nil, NewValidationError("supplier_id is required")
	}
	if poNumber == "" {
		return nil, NewValidationError("po_number is required")
	}
	if len(lineItems) == 0 {
		return nil, NewValidationError("at least one line item is required")
	}
	if createdBy == uuid.Nil {
		return nil, NewValidationError("created_by is required")
	}

	totalAmount := 0.0
	for _, item := range lineItems {
		if item.ProductID == uuid.Nil {
			return nil, NewValidationError("product_id is required for all line items")
		}
		if item.Quantity <= 0 {
			return nil, NewValidationError("quantity must be greater than zero")
		}
		if item.UnitCost < 0 {
			return nil, NewValidationError("unit_cost cannot be negative")
		}
		totalAmount += item.LineTotal
	}

	now := time.Now()
	return &PurchaseOrder{
		ID:                  uuid.New(),
		OrganizationID:      orgID,
		PONumber:            poNumber,
		SupplierID:          supplierID,
		Status:              POStatusDraft,
		OrderDate:           orderDate,
		ExpectedArrivalDate: expectedArrivalDate,
		TotalAmount:         totalAmount,
		LineItems:           lineItems,
		CreatedBy:           createdBy,
		CreatedAt:           now,
		UpdatedAt:           now,
		IsDelayed:           false,
		DelayDays:           0,
	}, nil
}

func (po *PurchaseOrder) CheckDelay() {
	if po.ActualArrivalDate != nil {
		po.DelayDays = int(po.ActualArrivalDate.Sub(po.ExpectedArrivalDate).Hours() / 24)
		po.IsDelayed = po.DelayDays > 0
	} else if time.Now().After(po.ExpectedArrivalDate) &&
		po.Status != POStatusReceived &&
		po.Status != POStatusClosed &&
		po.Status != POStatusCancelled {
		po.IsDelayed = true
		po.DelayDays = int(time.Since(po.ExpectedArrivalDate).Hours() / 24)
	}
}

func (po *PurchaseOrder) Confirm() error {
	if po.Status != POStatusDraft && po.Status != POStatusPending {
		return NewValidationError("can only confirm draft or pending purchase orders")
	}
	po.Status = POStatusConfirmed
	po.UpdatedAt = time.Now()
	return nil
}

func (po *PurchaseOrder) Cancel() error {
	if po.Status == POStatusReceived || po.Status == POStatusClosed || po.Status == POStatusCancelled {
		return NewValidationError("cannot cancel received, closed, or already cancelled purchase orders")
	}
	po.Status = POStatusCancelled
	po.UpdatedAt = time.Now()
	return nil
}

func (po *PurchaseOrder) UpdateReceiptStatus() {
	allReceived := true
	anyReceived := false

	for _, item := range po.LineItems {
		if item.ReceivedQty > 0 {
			anyReceived = true
		}
		if item.ReceivedQty < item.Quantity {
			allReceived = false
		}
	}

	if allReceived && anyReceived {
		po.Status = POStatusReceived
		if po.ActualArrivalDate == nil {
			now := time.Now()
			po.ActualArrivalDate = &now
		}
	} else if anyReceived {
		po.Status = POStatusPartial
	}

	po.CheckDelay()
	po.UpdatedAt = time.Now()
}