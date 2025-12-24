// Package dto provides Data Transfer Objects for HTTP API requests and responses.
package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreatePurchaseOrderRequest represents the request body for creating a purchase order.
type CreatePurchaseOrderRequest struct {
	PONumber            string              `json:"po_number"`
	SupplierID          string              `json:"supplier_id"`
	OrderDate           time.Time           `json:"order_date"`
	ExpectedArrivalDate time.Time           `json:"expected_arrival_date"`
	LineItems           []POLineItemRequest `json:"line_items"`
}

// POLineItemRequest represents a line item in a purchase order request.
type POLineItemRequest struct {
	ProductID string  `json:"product_id"`
	Quantity  float64 `json:"quantity"`
	UnitCost  float64 `json:"unit_cost"`
}

// ReceivePurchaseOrderRequest represents the request body for receiving goods.
type ReceivePurchaseOrderRequest struct {
	LocationID string               `json:"location_id"`
	Receipts   []ReceiptLineRequest `json:"receipts"`
}

// ReceiptLineRequest represents a line item receipt.
type ReceiptLineRequest struct {
	LineItemID  string  `json:"line_item_id"`
	ReceivedQty float64 `json:"received_qty"`
}

// CreateSalesOrderRequest represents the request body for creating a sales order.
type CreateSalesOrderRequest struct {
	SONumber   string              `json:"so_number"`
	CustomerID string              `json:"customer_id"`
	OrderDate  time.Time           `json:"order_date"`
	DueDate    time.Time           `json:"due_date"`
	LineItems  []SOLineItemRequest `json:"line_items"`
}

// SOLineItemRequest represents a line item in a sales order request.
type SOLineItemRequest struct {
	ProductID string  `json:"product_id"`
	Quantity  float64 `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
}

// ShipSalesOrderRequest represents the request body for shipping a sales order.
type ShipSalesOrderRequest struct {
	LocationID         string `json:"location_id"`
	DeliveryNoteNumber string `json:"delivery_note_number"`
}

// ListQueryParams represents common query parameters for list endpoints.
type ListQueryParams struct {
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
	Filters  map[string]string `json:"filters,omitempty"`
}

// ParseQueryParams extracts pagination and filter parameters from query string.
func ParseQueryParams(page, pageSize int, status, supplierID, customerID, productID string) ListQueryParams {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	filters := make(map[string]string)
	if status != "" {
		filters["status"] = status
	}
	if supplierID != "" {
		filters["supplier_id"] = supplierID
	}
	if customerID != "" {
		filters["customer_id"] = customerID
	}
	if productID != "" {
		filters["product_id"] = productID
	}

	return ListQueryParams{
		Page:     page,
		PageSize: pageSize,
		Filters:  filters,
	}
}

// ConvertFiltersToInterface converts string filters to interface{} map.
func (q ListQueryParams) ConvertFiltersToInterface() map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range q.Filters {
		if v != "" {
			if uid, err := uuid.Parse(v); err == nil {
				result[k] = uid
			} else {
				result[k] = v
			}
		}
	}
	return result
}
