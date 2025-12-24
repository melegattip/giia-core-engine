// Package dto provides Data Transfer Objects for HTTP API requests and responses.
package dto

import (
	"time"

	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/domain"
)

// PurchaseOrderResponse represents a purchase order in API responses.
type PurchaseOrderResponse struct {
	ID                  string               `json:"id"`
	OrganizationID      string               `json:"organization_id"`
	PONumber            string               `json:"po_number"`
	SupplierID          string               `json:"supplier_id"`
	Status              string               `json:"status"`
	OrderDate           time.Time            `json:"order_date"`
	ExpectedArrivalDate time.Time            `json:"expected_arrival_date"`
	ActualArrivalDate   *time.Time           `json:"actual_arrival_date,omitempty"`
	DelayDays           int                  `json:"delay_days"`
	IsDelayed           bool                 `json:"is_delayed"`
	TotalAmount         float64              `json:"total_amount"`
	LineItems           []POLineItemResponse `json:"line_items"`
	CreatedBy           string               `json:"created_by"`
	CreatedAt           time.Time            `json:"created_at"`
	UpdatedAt           time.Time            `json:"updated_at"`
}

// POLineItemResponse represents a purchase order line item in API responses.
type POLineItemResponse struct {
	ID              string  `json:"id"`
	PurchaseOrderID string  `json:"purchase_order_id"`
	ProductID       string  `json:"product_id"`
	Quantity        float64 `json:"quantity"`
	ReceivedQty     float64 `json:"received_qty"`
	UnitCost        float64 `json:"unit_cost"`
	LineTotal       float64 `json:"line_total"`
}

// SalesOrderResponse represents a sales order in API responses.
type SalesOrderResponse struct {
	ID                 string               `json:"id"`
	OrganizationID     string               `json:"organization_id"`
	SONumber           string               `json:"so_number"`
	CustomerID         string               `json:"customer_id"`
	Status             string               `json:"status"`
	OrderDate          time.Time            `json:"order_date"`
	DueDate            time.Time            `json:"due_date"`
	ShipDate           *time.Time           `json:"ship_date,omitempty"`
	DeliveryNoteIssued bool                 `json:"delivery_note_issued"`
	DeliveryNoteNumber string               `json:"delivery_note_number,omitempty"`
	DeliveryNoteDate   *time.Time           `json:"delivery_note_date,omitempty"`
	TotalAmount        float64              `json:"total_amount"`
	LineItems          []SOLineItemResponse `json:"line_items"`
	CreatedAt          time.Time            `json:"created_at"`
	UpdatedAt          time.Time            `json:"updated_at"`
}

// SOLineItemResponse represents a sales order line item in API responses.
type SOLineItemResponse struct {
	ID           string  `json:"id"`
	SalesOrderID string  `json:"sales_order_id"`
	ProductID    string  `json:"product_id"`
	Quantity     float64 `json:"quantity"`
	UnitPrice    float64 `json:"unit_price"`
	LineTotal    float64 `json:"line_total"`
}

// InventoryBalanceResponse represents inventory balance in API responses.
type InventoryBalanceResponse struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	ProductID      string    `json:"product_id"`
	LocationID     string    `json:"location_id"`
	OnHand         float64   `json:"on_hand"`
	Reserved       float64   `json:"reserved"`
	Available      float64   `json:"available"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// InventoryTransactionResponse represents an inventory transaction in API responses.
type InventoryTransactionResponse struct {
	ID              string    `json:"id"`
	OrganizationID  string    `json:"organization_id"`
	ProductID       string    `json:"product_id"`
	LocationID      string    `json:"location_id"`
	Type            string    `json:"type"`
	Quantity        float64   `json:"quantity"`
	UnitCost        float64   `json:"unit_cost"`
	ReferenceType   string    `json:"reference_type"`
	ReferenceID     string    `json:"reference_id"`
	Reason          string    `json:"reason"`
	TransactionDate time.Time `json:"transaction_date"`
	CreatedBy       string    `json:"created_by"`
	CreatedAt       time.Time `json:"created_at"`
}

// PaginatedResponse represents a paginated list response.
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalCount int64       `json:"total_count"`
	TotalPages int         `json:"total_pages"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	ErrorCode string `json:"error_code,omitempty"`
	Message   string `json:"message"`
}

// MessageResponse represents a simple message response.
type MessageResponse struct {
	Message string `json:"message"`
}

// ToPurchaseOrderResponse converts a domain purchase order to API response.
func ToPurchaseOrderResponse(po *domain.PurchaseOrder) *PurchaseOrderResponse {
	if po == nil {
		return nil
	}

	lineItems := make([]POLineItemResponse, len(po.LineItems))
	for i, item := range po.LineItems {
		lineItems[i] = POLineItemResponse{
			ID:              item.ID.String(),
			PurchaseOrderID: item.PurchaseOrderID.String(),
			ProductID:       item.ProductID.String(),
			Quantity:        item.Quantity,
			ReceivedQty:     item.ReceivedQty,
			UnitCost:        item.UnitCost,
			LineTotal:       item.LineTotal,
		}
	}

	return &PurchaseOrderResponse{
		ID:                  po.ID.String(),
		OrganizationID:      po.OrganizationID.String(),
		PONumber:            po.PONumber,
		SupplierID:          po.SupplierID.String(),
		Status:              string(po.Status),
		OrderDate:           po.OrderDate,
		ExpectedArrivalDate: po.ExpectedArrivalDate,
		ActualArrivalDate:   po.ActualArrivalDate,
		DelayDays:           po.DelayDays,
		IsDelayed:           po.IsDelayed,
		TotalAmount:         po.TotalAmount,
		LineItems:           lineItems,
		CreatedBy:           po.CreatedBy.String(),
		CreatedAt:           po.CreatedAt,
		UpdatedAt:           po.UpdatedAt,
	}
}

// ToPurchaseOrderListResponse converts a slice of domain purchase orders to API responses.
func ToPurchaseOrderListResponse(orders []*domain.PurchaseOrder) []*PurchaseOrderResponse {
	result := make([]*PurchaseOrderResponse, len(orders))
	for i, po := range orders {
		result[i] = ToPurchaseOrderResponse(po)
	}
	return result
}

// ToSalesOrderResponse converts a domain sales order to API response.
func ToSalesOrderResponse(so *domain.SalesOrder) *SalesOrderResponse {
	if so == nil {
		return nil
	}

	lineItems := make([]SOLineItemResponse, len(so.LineItems))
	for i, item := range so.LineItems {
		lineItems[i] = SOLineItemResponse{
			ID:           item.ID.String(),
			SalesOrderID: item.SalesOrderID.String(),
			ProductID:    item.ProductID.String(),
			Quantity:     item.Quantity,
			UnitPrice:    item.UnitPrice,
			LineTotal:    item.LineTotal,
		}
	}

	return &SalesOrderResponse{
		ID:                 so.ID.String(),
		OrganizationID:     so.OrganizationID.String(),
		SONumber:           so.SONumber,
		CustomerID:         so.CustomerID.String(),
		Status:             string(so.Status),
		OrderDate:          so.OrderDate,
		DueDate:            so.DueDate,
		ShipDate:           so.ShipDate,
		DeliveryNoteIssued: so.DeliveryNoteIssued,
		DeliveryNoteNumber: so.DeliveryNoteNumber,
		DeliveryNoteDate:   so.DeliveryNoteDate,
		TotalAmount:        so.TotalAmount,
		LineItems:          lineItems,
		CreatedAt:          so.CreatedAt,
		UpdatedAt:          so.UpdatedAt,
	}
}

// ToSalesOrderListResponse converts a slice of domain sales orders to API responses.
func ToSalesOrderListResponse(orders []*domain.SalesOrder) []*SalesOrderResponse {
	result := make([]*SalesOrderResponse, len(orders))
	for i, so := range orders {
		result[i] = ToSalesOrderResponse(so)
	}
	return result
}

// ToInventoryBalanceResponse converts a domain inventory balance to API response.
func ToInventoryBalanceResponse(ib *domain.InventoryBalance) *InventoryBalanceResponse {
	if ib == nil {
		return nil
	}

	return &InventoryBalanceResponse{
		ID:             ib.ID.String(),
		OrganizationID: ib.OrganizationID.String(),
		ProductID:      ib.ProductID.String(),
		LocationID:     ib.LocationID.String(),
		OnHand:         ib.OnHand,
		Reserved:       ib.Reserved,
		Available:      ib.Available,
		UpdatedAt:      ib.UpdatedAt,
	}
}

// ToInventoryBalanceListResponse converts a slice of domain inventory balances to API responses.
func ToInventoryBalanceListResponse(balances []*domain.InventoryBalance) []*InventoryBalanceResponse {
	result := make([]*InventoryBalanceResponse, len(balances))
	for i, ib := range balances {
		result[i] = ToInventoryBalanceResponse(ib)
	}
	return result
}

// ToInventoryTransactionResponse converts a domain inventory transaction to API response.
func ToInventoryTransactionResponse(txn *domain.InventoryTransaction) *InventoryTransactionResponse {
	if txn == nil {
		return nil
	}

	return &InventoryTransactionResponse{
		ID:              txn.ID.String(),
		OrganizationID:  txn.OrganizationID.String(),
		ProductID:       txn.ProductID.String(),
		LocationID:      txn.LocationID.String(),
		Type:            string(txn.Type),
		Quantity:        txn.Quantity,
		UnitCost:        txn.UnitCost,
		ReferenceType:   txn.ReferenceType,
		ReferenceID:     txn.ReferenceID.String(),
		Reason:          txn.Reason,
		TransactionDate: txn.TransactionDate,
		CreatedBy:       txn.CreatedBy.String(),
		CreatedAt:       txn.CreatedAt,
	}
}

// ToInventoryTransactionListResponse converts a slice of domain inventory transactions to API responses.
func ToInventoryTransactionListResponse(transactions []*domain.InventoryTransaction) []*InventoryTransactionResponse {
	result := make([]*InventoryTransactionResponse, len(transactions))
	for i, txn := range transactions {
		result[i] = ToInventoryTransactionResponse(txn)
	}
	return result
}

// NewPaginatedResponse creates a new paginated response.
func NewPaginatedResponse(data interface{}, page, pageSize int, totalCount int64) *PaginatedResponse {
	totalPages := int(totalCount) / pageSize
	if int(totalCount)%pageSize > 0 {
		totalPages++
	}

	return &PaginatedResponse{
		Data:       data,
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
		TotalPages: totalPages,
	}
}
