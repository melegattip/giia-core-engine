// Package nats provides an event publisher implementation using NATS.
package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/providers"
)

// EventSubjects for NATS messaging.
const (
	SubjectPOCreated        = "execution.purchase_order.created"
	SubjectPOUpdated        = "execution.purchase_order.updated"
	SubjectPOReceived       = "execution.purchase_order.received"
	SubjectPOCancelled      = "execution.purchase_order.cancelled"
	SubjectSOCreated        = "execution.sales_order.created"
	SubjectSOUpdated        = "execution.sales_order.updated"
	SubjectSOCancelled      = "execution.sales_order.cancelled"
	SubjectDeliveryNote     = "execution.sales_order.delivery_note_issued"
	SubjectInventoryUpdated = "execution.inventory.updated"
	SubjectAlertCreated     = "execution.alert.created"
)

// NATSClient interface for NATS connection.
type NATSClient interface {
	Publish(subject string, data []byte) error
	Close()
}

// EventPublisher implements providers.EventPublisher using NATS.
type EventPublisher struct {
	client NATSClient
}

// NewEventPublisher creates a new NATS event publisher.
func NewEventPublisher(client NATSClient) *EventPublisher {
	return &EventPublisher{
		client: client,
	}
}

// Event represents a base event structure.
type Event struct {
	Type      string      `json:"type"`
	Timestamp string      `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// publishEvent publishes an event to NATS.
func (p *EventPublisher) publishEvent(subject string, eventType string, data interface{}) error {
	event := Event{
		Type:      eventType,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Data:      data,
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	if p.client == nil {
		// Log and skip if no client configured (useful for testing)
		return nil
	}

	return p.client.Publish(subject, payload)
}

// PurchaseOrderEvent represents a PO event payload.
type PurchaseOrderEvent struct {
	ID             string  `json:"id"`
	OrganizationID string  `json:"organization_id"`
	PONumber       string  `json:"po_number"`
	SupplierID     string  `json:"supplier_id"`
	Status         string  `json:"status"`
	TotalAmount    float64 `json:"total_amount"`
}

// toPOEvent converts a PurchaseOrder to an event payload.
func toPOEvent(po *domain.PurchaseOrder) *PurchaseOrderEvent {
	return &PurchaseOrderEvent{
		ID:             po.ID.String(),
		OrganizationID: po.OrganizationID.String(),
		PONumber:       po.PONumber,
		SupplierID:     po.SupplierID.String(),
		Status:         string(po.Status),
		TotalAmount:    po.TotalAmount,
	}
}

// PublishPOCreated publishes a PO created event.
func (p *EventPublisher) PublishPOCreated(ctx context.Context, po *domain.PurchaseOrder) error {
	return p.publishEvent(SubjectPOCreated, "purchase_order.created", toPOEvent(po))
}

// PublishPOUpdated publishes a PO updated event.
func (p *EventPublisher) PublishPOUpdated(ctx context.Context, po *domain.PurchaseOrder) error {
	return p.publishEvent(SubjectPOUpdated, "purchase_order.updated", toPOEvent(po))
}

// PublishPOReceived publishes a PO received event.
func (p *EventPublisher) PublishPOReceived(ctx context.Context, po *domain.PurchaseOrder) error {
	return p.publishEvent(SubjectPOReceived, "purchase_order.received", toPOEvent(po))
}

// PublishPOCancelled publishes a PO cancelled event.
func (p *EventPublisher) PublishPOCancelled(ctx context.Context, po *domain.PurchaseOrder) error {
	return p.publishEvent(SubjectPOCancelled, "purchase_order.cancelled", toPOEvent(po))
}

// SalesOrderEvent represents a SO event payload.
type SalesOrderEvent struct {
	ID             string  `json:"id"`
	OrganizationID string  `json:"organization_id"`
	SONumber       string  `json:"so_number"`
	CustomerID     string  `json:"customer_id"`
	Status         string  `json:"status"`
	TotalAmount    float64 `json:"total_amount"`
}

// toSOEvent converts a SalesOrder to an event payload.
func toSOEvent(so *domain.SalesOrder) *SalesOrderEvent {
	return &SalesOrderEvent{
		ID:             so.ID.String(),
		OrganizationID: so.OrganizationID.String(),
		SONumber:       so.SONumber,
		CustomerID:     so.CustomerID.String(),
		Status:         string(so.Status),
		TotalAmount:    so.TotalAmount,
	}
}

// PublishSOCreated publishes a SO created event.
func (p *EventPublisher) PublishSOCreated(ctx context.Context, so *domain.SalesOrder) error {
	return p.publishEvent(SubjectSOCreated, "sales_order.created", toSOEvent(so))
}

// PublishSOUpdated publishes a SO updated event.
func (p *EventPublisher) PublishSOUpdated(ctx context.Context, so *domain.SalesOrder) error {
	return p.publishEvent(SubjectSOUpdated, "sales_order.updated", toSOEvent(so))
}

// PublishSOCancelled publishes a SO cancelled event.
func (p *EventPublisher) PublishSOCancelled(ctx context.Context, so *domain.SalesOrder) error {
	return p.publishEvent(SubjectSOCancelled, "sales_order.cancelled", toSOEvent(so))
}

// PublishDeliveryNoteIssued publishes a delivery note issued event.
func (p *EventPublisher) PublishDeliveryNoteIssued(ctx context.Context, so *domain.SalesOrder) error {
	event := struct {
		*SalesOrderEvent
		DeliveryNoteNumber string `json:"delivery_note_number"`
	}{
		SalesOrderEvent:    toSOEvent(so),
		DeliveryNoteNumber: so.DeliveryNoteNumber,
	}
	return p.publishEvent(SubjectDeliveryNote, "sales_order.delivery_note_issued", event)
}

// InventoryEvent represents an inventory event payload.
type InventoryEvent struct {
	ID             string  `json:"id"`
	OrganizationID string  `json:"organization_id"`
	ProductID      string  `json:"product_id"`
	LocationID     string  `json:"location_id"`
	Type           string  `json:"type"`
	Quantity       float64 `json:"quantity"`
	ReferenceType  string  `json:"reference_type"`
	ReferenceID    string  `json:"reference_id"`
}

// PublishInventoryUpdated publishes an inventory updated event.
func (p *EventPublisher) PublishInventoryUpdated(ctx context.Context, txn *domain.InventoryTransaction) error {
	event := &InventoryEvent{
		ID:             txn.ID.String(),
		OrganizationID: txn.OrganizationID.String(),
		ProductID:      txn.ProductID.String(),
		LocationID:     txn.LocationID.String(),
		Type:           string(txn.Type),
		Quantity:       txn.Quantity,
		ReferenceType:  txn.ReferenceType,
		ReferenceID:    txn.ReferenceID.String(),
	}
	return p.publishEvent(SubjectInventoryUpdated, "inventory.updated", event)
}

// AlertEvent represents an alert event payload.
type AlertEvent struct {
	ID             string `json:"id"`
	OrganizationID string `json:"organization_id"`
	Type           string `json:"type"`
	Severity       string `json:"severity"`
	Message        string `json:"message"`
}

// PublishAlertCreated publishes an alert created event.
func (p *EventPublisher) PublishAlertCreated(ctx context.Context, alert *domain.Alert) error {
	event := &AlertEvent{
		ID:             alert.ID.String(),
		OrganizationID: alert.OrganizationID.String(),
		Type:           string(alert.AlertType),
		Severity:       string(alert.Severity),
		Message:        alert.Message,
	}
	return p.publishEvent(SubjectAlertCreated, "alert.created", event)
}

// Ensure EventPublisher implements providers.EventPublisher.
var _ providers.EventPublisher = (*EventPublisher)(nil)
