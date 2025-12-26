package events

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEventTypesConstants(t *testing.T) {
	// Verify all type constants are defined
	assert.Equal(t, "purchase_order.created", TypePOCreated)
	assert.Equal(t, "purchase_order.updated", TypePOUpdated)
	assert.Equal(t, "purchase_order.received", TypePOReceived)
	assert.Equal(t, "purchase_order.cancelled", TypePOCancelled)
	assert.Equal(t, "purchase_order.approved", TypePOApproved)

	assert.Equal(t, "sales_order.created", TypeSOCreated)
	assert.Equal(t, "sales_order.updated", TypeSOUpdated)
	assert.Equal(t, "sales_order.shipped", TypeSOShipped)
	assert.Equal(t, "sales_order.cancelled", TypeSOCancelled)
	assert.Equal(t, "sales_order.delivery_note_issued", TypeSODeliveryNoteIssued)

	assert.Equal(t, "inventory.updated", TypeInventoryUpdated)
	assert.Equal(t, "inventory.adjusted", TypeInventoryAdjusted)
	assert.Equal(t, "inventory.transferred", TypeInventoryTransferred)

	assert.Equal(t, "alert.created", TypeAlertCreated)
	assert.Equal(t, "alert.resolved", TypeAlertResolved)
}

func TestPurchaseOrderLineEvent_Marshal(t *testing.T) {
	event := PurchaseOrderLineEvent{
		ID:          "line-123",
		ProductID:   "prod-123",
		ProductSKU:  "SKU-001",
		Quantity:    100.0,
		UnitPrice:   25.00,
		ReceivedQty: 50.0,
		LineTotal:   2500.00,
	}

	assert.Equal(t, "line-123", event.ID)
	assert.Equal(t, 100.0, event.Quantity)
	assert.Equal(t, 2500.00, event.LineTotal)
}

func TestSalesOrderLineEvent_Marshal(t *testing.T) {
	event := SalesOrderLineEvent{
		ID:         "line-456",
		ProductID:  "prod-456",
		ProductSKU: "SKU-002",
		Quantity:   50.0,
		UnitPrice:  15.00,
		ShippedQty: 25.0,
		LineTotal:  750.00,
	}

	assert.Equal(t, "line-456", event.ID)
	assert.Equal(t, 50.0, event.Quantity)
	assert.Equal(t, 750.00, event.LineTotal)
}

func TestInventoryBalanceEvent_Marshal(t *testing.T) {
	event := InventoryBalanceEvent{
		OrganizationID: "org-123",
		ProductID:      "prod-123",
		LocationID:     "loc-123",
		OnHandQty:      500.0,
		AllocatedQty:   100.0,
		AvailableQty:   400.0,
		InTransitQty:   50.0,
		MinimumQty:     100.0,
		MaximumQty:     1000.0,
		ReorderPoint:   200.0,
		UpdatedAt:      time.Now(),
	}

	assert.Equal(t, 500.0, event.OnHandQty)
	assert.Equal(t, 400.0, event.AvailableQty)
	assert.Equal(t, 200.0, event.ReorderPoint)
}

func TestInventoryBalanceAlertEvent_Marshal(t *testing.T) {
	event := InventoryBalanceAlertEvent{
		AlertEvent: AlertEvent{
			ID:             "alert-123",
			OrganizationID: "org-123",
			AlertType:      "low_stock",
			Severity:       "high",
			Status:         "active",
			Message:        "Low stock warning",
		},
		ProductID:    "prod-123",
		LocationID:   "loc-123",
		CurrentQty:   50.0,
		ThresholdQty: 100.0,
		BufferZone:   "red",
	}

	assert.Equal(t, "prod-123", event.ProductID)
	assert.Equal(t, 50.0, event.CurrentQty)
	assert.Equal(t, "red", event.BufferZone)
}

func TestDeliveryNoteEvent_Marshal(t *testing.T) {
	event := DeliveryNoteEvent{
		SalesOrderEvent: SalesOrderEvent{
			ID:                 "so-123",
			OrganizationID:     "org-123",
			SONumber:           "SO-2024-001",
			CustomerID:         "cust-123",
			Status:             "shipped",
			DeliveryNoteNumber: "DN-001",
		},
		ShippingAddress:  "123 Main St",
		Carrier:          "FedEx",
		TrackingNumber:   "TRACK123",
		EstimatedArrival: time.Now().Add(48 * time.Hour),
	}

	assert.Equal(t, "DN-001", event.DeliveryNoteNumber)
	assert.Equal(t, "FedEx", event.Carrier)
	assert.Equal(t, "TRACK123", event.TrackingNumber)
}

func TestNewPublisher_WithNilConnection(t *testing.T) {
	publisher, err := NewPublisher(nil, nil)

	assert.NoError(t, err)
	assert.NotNil(t, publisher)
	assert.False(t, publisher.IsEnabled())
}

func TestPublisher_CreateEnvelope(t *testing.T) {
	publisher := NewNoOpPublisher()

	// Test that the publisher can be created and used
	err := publisher.Publish(context.Background(), "test.subject", "test.type", "org-123", map[string]string{"key": "value"})
	assert.NoError(t, err)
}

func TestPublisher_ContextCancellation(t *testing.T) {
	publisher := NewNoOpPublisher()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Should still work with NoOp publisher
	err := publisher.Publish(ctx, "test.subject", "test.type", "org-123", nil)
	assert.NoError(t, err)
}

func TestPublisherConfig_Custom(t *testing.T) {
	config := &PublisherConfig{
		MaxRetries:     5,
		InitialBackoff: 200 * time.Millisecond,
		MaxBackoff:     5 * time.Second,
		AsyncMode:      true,
	}

	assert.Equal(t, 5, config.MaxRetries)
	assert.Equal(t, 200*time.Millisecond, config.InitialBackoff)
	assert.Equal(t, 5*time.Second, config.MaxBackoff)
	assert.True(t, config.AsyncMode)
}
