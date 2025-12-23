package domain

import (
	"time"

	"github.com/google/uuid"
)

type SalesOrder struct {
	ID                 uuid.UUID
	OrganizationID     uuid.UUID
	SONumber           string
	CustomerID         uuid.UUID
	Status             SOStatus
	OrderDate          time.Time
	DueDate            time.Time
	ShipDate           *time.Time
	DeliveryNoteIssued bool
	DeliveryNoteNumber string
	DeliveryNoteDate   *time.Time
	TotalAmount        float64
	LineItems          []SOLineItem
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type SOLineItem struct {
	ID           uuid.UUID
	SalesOrderID uuid.UUID
	ProductID    uuid.UUID
	Quantity     float64
	UnitPrice    float64
	LineTotal    float64
}

type SOStatus string

const (
	SOStatusPending   SOStatus = "pending"
	SOStatusConfirmed SOStatus = "confirmed"
	SOStatusPicking   SOStatus = "picking"
	SOStatusPacked    SOStatus = "packed"
	SOStatusShipped   SOStatus = "shipped"
	SOStatusDelivered SOStatus = "delivered"
	SOStatusCancelled SOStatus = "cancelled"
)

func (s SOStatus) IsValid() bool {
	switch s {
	case SOStatusPending, SOStatusConfirmed, SOStatusPicking,
		SOStatusPacked, SOStatusShipped, SOStatusDelivered, SOStatusCancelled:
		return true
	}
	return false
}

func NewSalesOrder(
	orgID, customerID uuid.UUID,
	soNumber string,
	orderDate, dueDate time.Time,
	lineItems []SOLineItem,
) (*SalesOrder, error) {
	if orgID == uuid.Nil {
		return nil, NewValidationError("organization_id is required")
	}
	if customerID == uuid.Nil {
		return nil, NewValidationError("customer_id is required")
	}
	if soNumber == "" {
		return nil, NewValidationError("so_number is required")
	}
	if len(lineItems) == 0 {
		return nil, NewValidationError("at least one line item is required")
	}

	totalAmount := 0.0
	for _, item := range lineItems {
		if item.ProductID == uuid.Nil {
			return nil, NewValidationError("product_id is required for all line items")
		}
		if item.Quantity <= 0 {
			return nil, NewValidationError("quantity must be greater than zero")
		}
		if item.UnitPrice < 0 {
			return nil, NewValidationError("unit_price cannot be negative")
		}
		totalAmount += item.LineTotal
	}

	now := time.Now()
	return &SalesOrder{
		ID:                 uuid.New(),
		OrganizationID:     orgID,
		SONumber:           soNumber,
		CustomerID:         customerID,
		Status:             SOStatusPending,
		OrderDate:          orderDate,
		DueDate:            dueDate,
		TotalAmount:        totalAmount,
		LineItems:          lineItems,
		CreatedAt:          now,
		UpdatedAt:          now,
		DeliveryNoteIssued: false,
	}, nil
}

func (so *SalesOrder) IsQualifiedDemand() bool {
	return so.Status == SOStatusConfirmed && !so.DeliveryNoteIssued
}

func (so *SalesOrder) IssueDeliveryNote(noteNumber string) error {
	if so.DeliveryNoteIssued {
		return NewValidationError("delivery note already issued")
	}
	if noteNumber == "" {
		return NewValidationError("delivery note number is required")
	}
	if so.Status != SOStatusConfirmed && so.Status != SOStatusPicking && so.Status != SOStatusPacked {
		return NewValidationError("can only issue delivery note for confirmed, picking, or packed orders")
	}

	now := time.Now()
	so.DeliveryNoteIssued = true
	so.DeliveryNoteNumber = noteNumber
	so.DeliveryNoteDate = &now
	so.UpdatedAt = now

	return nil
}

func (so *SalesOrder) Confirm() error {
	if so.Status != SOStatusPending {
		return NewValidationError("can only confirm pending sales orders")
	}
	so.Status = SOStatusConfirmed
	so.UpdatedAt = time.Now()
	return nil
}

func (so *SalesOrder) Cancel() error {
	if so.Status == SOStatusShipped || so.Status == SOStatusDelivered || so.Status == SOStatusCancelled {
		return NewValidationError("cannot cancel shipped, delivered, or already cancelled sales orders")
	}
	so.Status = SOStatusCancelled
	so.UpdatedAt = time.Now()
	return nil
}

func (so *SalesOrder) MarkAsShipped() error {
	if so.Status != SOStatusPacked {
		return NewValidationError("can only ship packed orders")
	}
	now := time.Now()
	so.Status = SOStatusShipped
	so.ShipDate = &now
	so.UpdatedAt = now
	return nil
}