// Package events provides event types for the Execution Service.
package events

import (
	"time"
)

// Event type constants.
const (
	TypePOCreated   = "purchase_order.created"
	TypePOUpdated   = "purchase_order.updated"
	TypePOReceived  = "purchase_order.received"
	TypePOCancelled = "purchase_order.cancelled"
	TypePOApproved  = "purchase_order.approved"

	TypeSOCreated            = "sales_order.created"
	TypeSOUpdated            = "sales_order.updated"
	TypeSOShipped            = "sales_order.shipped"
	TypeSOCancelled          = "sales_order.cancelled"
	TypeSODeliveryNoteIssued = "sales_order.delivery_note_issued"

	TypeInventoryUpdated     = "inventory.updated"
	TypeInventoryAdjusted    = "inventory.adjusted"
	TypeInventoryTransferred = "inventory.transferred"

	TypeAlertCreated  = "alert.created"
	TypeAlertResolved = "alert.resolved"
)

// PurchaseOrderEvent represents a purchase order event payload.
type PurchaseOrderEvent struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	PONumber       string    `json:"po_number"`
	SupplierID     string    `json:"supplier_id"`
	SupplierName   string    `json:"supplier_name,omitempty"`
	LocationID     string    `json:"location_id,omitempty"`
	Status         string    `json:"status"`
	TotalAmount    float64   `json:"total_amount"`
	Currency       string    `json:"currency,omitempty"`
	OrderedAt      time.Time `json:"ordered_at,omitempty"`
	ExpectedAt     time.Time `json:"expected_at,omitempty"`
	ReceivedAt     time.Time `json:"received_at,omitempty"`
	ItemCount      int       `json:"item_count,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// PurchaseOrderLineEvent represents a PO line item.
type PurchaseOrderLineEvent struct {
	ID          string  `json:"id"`
	ProductID   string  `json:"product_id"`
	ProductSKU  string  `json:"product_sku,omitempty"`
	Quantity    float64 `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	ReceivedQty float64 `json:"received_qty"`
	LineTotal   float64 `json:"line_total"`
}

// SalesOrderEvent represents a sales order event payload.
type SalesOrderEvent struct {
	ID                 string    `json:"id"`
	OrganizationID     string    `json:"organization_id"`
	SONumber           string    `json:"so_number"`
	CustomerID         string    `json:"customer_id"`
	CustomerName       string    `json:"customer_name,omitempty"`
	LocationID         string    `json:"location_id,omitempty"`
	Status             string    `json:"status"`
	TotalAmount        float64   `json:"total_amount"`
	Currency           string    `json:"currency,omitempty"`
	OrderedAt          time.Time `json:"ordered_at,omitempty"`
	ShippedAt          time.Time `json:"shipped_at,omitempty"`
	DeliveredAt        time.Time `json:"delivered_at,omitempty"`
	DeliveryNoteNumber string    `json:"delivery_note_number,omitempty"`
	ItemCount          int       `json:"item_count,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// SalesOrderLineEvent represents a SO line item.
type SalesOrderLineEvent struct {
	ID         string  `json:"id"`
	ProductID  string  `json:"product_id"`
	ProductSKU string  `json:"product_sku,omitempty"`
	Quantity   float64 `json:"quantity"`
	UnitPrice  float64 `json:"unit_price"`
	ShippedQty float64 `json:"shipped_qty"`
	LineTotal  float64 `json:"line_total"`
}

// InventoryEvent represents an inventory change event payload.
type InventoryEvent struct {
	ID              string    `json:"id"`
	OrganizationID  string    `json:"organization_id"`
	ProductID       string    `json:"product_id"`
	ProductSKU      string    `json:"product_sku,omitempty"`
	ProductName     string    `json:"product_name,omitempty"`
	LocationID      string    `json:"location_id"`
	LocationName    string    `json:"location_name,omitempty"`
	TransactionType string    `json:"transaction_type"`
	Quantity        float64   `json:"quantity"`
	PreviousBalance float64   `json:"previous_balance,omitempty"`
	NewBalance      float64   `json:"new_balance,omitempty"`
	ReferenceType   string    `json:"reference_type,omitempty"`
	ReferenceID     string    `json:"reference_id,omitempty"`
	Reason          string    `json:"reason,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

// InventoryBalanceEvent represents an inventory balance snapshot.
type InventoryBalanceEvent struct {
	OrganizationID string    `json:"organization_id"`
	ProductID      string    `json:"product_id"`
	LocationID     string    `json:"location_id"`
	OnHandQty      float64   `json:"on_hand_qty"`
	AllocatedQty   float64   `json:"allocated_qty"`
	AvailableQty   float64   `json:"available_qty"`
	InTransitQty   float64   `json:"in_transit_qty,omitempty"`
	MinimumQty     float64   `json:"minimum_qty,omitempty"`
	MaximumQty     float64   `json:"maximum_qty,omitempty"`
	ReorderPoint   float64   `json:"reorder_point,omitempty"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// AlertEvent represents an alert event payload.
type AlertEvent struct {
	ID             string            `json:"id"`
	OrganizationID string            `json:"organization_id"`
	AlertType      string            `json:"alert_type"`
	Severity       string            `json:"severity"`
	Status         string            `json:"status"`
	Title          string            `json:"title,omitempty"`
	Message        string            `json:"message"`
	ResourceType   string            `json:"resource_type,omitempty"`
	ResourceID     string            `json:"resource_id,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
	AcknowledgedAt time.Time         `json:"acknowledged_at,omitempty"`
	ResolvedAt     time.Time         `json:"resolved_at,omitempty"`
}

// InventoryBalanceAlertEvent represents a low stock or overstock alert.
type InventoryBalanceAlertEvent struct {
	AlertEvent
	ProductID    string  `json:"product_id"`
	LocationID   string  `json:"location_id"`
	CurrentQty   float64 `json:"current_qty"`
	ThresholdQty float64 `json:"threshold_qty"`
	BufferZone   string  `json:"buffer_zone,omitempty"`
}

// DeliveryNoteEvent represents a delivery note issuance event.
type DeliveryNoteEvent struct {
	SalesOrderEvent
	ShippingAddress  string    `json:"shipping_address,omitempty"`
	Carrier          string    `json:"carrier,omitempty"`
	TrackingNumber   string    `json:"tracking_number,omitempty"`
	EstimatedArrival time.Time `json:"estimated_arrival,omitempty"`
}
