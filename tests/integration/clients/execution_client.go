// Package clients provides HTTP and gRPC clients for integration testing.
package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// ExecutionClient provides methods to interact with the Execution Service.
type ExecutionClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewExecutionClient creates a new ExecutionClient.
func NewExecutionClient(baseURL string) *ExecutionClient {
	return &ExecutionClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// PurchaseOrder represents a purchase order.
type PurchaseOrder struct {
	ID             string              `json:"id"`
	OrderNumber    string              `json:"order_number"`
	OrganizationID string              `json:"organization_id"`
	SupplierID     string              `json:"supplier_id"`
	Status         string              `json:"status"`
	Items          []PurchaseOrderItem `json:"items"`
	TotalAmount    float64             `json:"total_amount"`
	ExpectedDate   time.Time           `json:"expected_date"`
	ReceivedDate   *time.Time          `json:"received_date,omitempty"`
	Notes          string              `json:"notes"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
}

// PurchaseOrderItem represents a purchase order line item.
type PurchaseOrderItem struct {
	ID               string  `json:"id"`
	ProductID        string  `json:"product_id"`
	SKU              string  `json:"sku"`
	Quantity         float64 `json:"quantity"`
	ReceivedQuantity float64 `json:"received_quantity"`
	UnitPrice        float64 `json:"unit_price"`
}

// SalesOrder represents a sales order.
type SalesOrder struct {
	ID             string           `json:"id"`
	OrderNumber    string           `json:"order_number"`
	OrganizationID string           `json:"organization_id"`
	CustomerID     string           `json:"customer_id"`
	Status         string           `json:"status"`
	Items          []SalesOrderItem `json:"items"`
	TotalAmount    float64          `json:"total_amount"`
	ShippedDate    *time.Time       `json:"shipped_date,omitempty"`
	Notes          string           `json:"notes"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
}

// SalesOrderItem represents a sales order line item.
type SalesOrderItem struct {
	ID              string  `json:"id"`
	ProductID       string  `json:"product_id"`
	SKU             string  `json:"sku"`
	Quantity        float64 `json:"quantity"`
	ShippedQuantity float64 `json:"shipped_quantity"`
	UnitPrice       float64 `json:"unit_price"`
}

// InventoryBalance represents inventory balance for a product.
type InventoryBalance struct {
	ProductID      string    `json:"product_id"`
	OrganizationID string    `json:"organization_id"`
	OnHand         float64   `json:"on_hand"`
	Reserved       float64   `json:"reserved"`
	Available      float64   `json:"available"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// InventoryTransaction represents an inventory transaction.
type InventoryTransaction struct {
	ID              string    `json:"id"`
	ProductID       string    `json:"product_id"`
	OrganizationID  string    `json:"organization_id"`
	TransactionType string    `json:"transaction_type"`
	Quantity        float64   `json:"quantity"`
	ReferenceType   string    `json:"reference_type"`
	ReferenceID     string    `json:"reference_id"`
	Notes           string    `json:"notes"`
	CreatedAt       time.Time `json:"created_at"`
}

// CreateOrderItemRequest represents an order item in a create request.
type CreateOrderItemRequest struct {
	ProductID string  `json:"product_id"`
	SKU       string  `json:"sku"`
	Quantity  float64 `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
}

// CreatePurchaseOrderRequest represents a request to create a purchase order.
type CreatePurchaseOrderRequest struct {
	OrganizationID string                   `json:"organization_id"`
	SupplierID     string                   `json:"supplier_id"`
	Items          []CreateOrderItemRequest `json:"items"`
	ExpectedDate   time.Time                `json:"expected_date,omitempty"`
	Notes          string                   `json:"notes"`
}

// CreatePurchaseOrderResponse represents a response from creating a purchase order.
type CreatePurchaseOrderResponse struct {
	Order PurchaseOrder `json:"order"`
}

// CreateSalesOrderRequest represents a request to create a sales order.
type CreateSalesOrderRequest struct {
	OrganizationID string                   `json:"organization_id"`
	CustomerID     string                   `json:"customer_id"`
	Items          []CreateOrderItemRequest `json:"items"`
	Notes          string                   `json:"notes"`
}

// CreateSalesOrderResponse represents a response from creating a sales order.
type CreateSalesOrderResponse struct {
	Order SalesOrder `json:"order"`
}

// ReceiveGoodsRequest represents a request to receive goods from a purchase order.
type ReceiveGoodsRequest struct {
	Items []ReceiveItemRequest `json:"items"`
}

// ReceiveItemRequest represents an item being received.
type ReceiveItemRequest struct {
	ProductID string  `json:"product_id"`
	Quantity  float64 `json:"quantity"`
}

// ShipOrderRequest represents a request to ship a sales order.
type ShipOrderRequest struct {
	Items []ShipItemRequest `json:"items"`
}

// ShipItemRequest represents an item being shipped.
type ShipItemRequest struct {
	ProductID string  `json:"product_id"`
	Quantity  float64 `json:"quantity"`
}

// CreatePurchaseOrder creates a new purchase order.
func (c *ExecutionClient) CreatePurchaseOrder(ctx context.Context, req CreatePurchaseOrderRequest, accessToken string) (*CreatePurchaseOrderResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v1/purchase-orders", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("create purchase order failed with status %d: %v", resp.StatusCode, errResp)
	}

	var result CreatePurchaseOrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetPurchaseOrder gets a purchase order by ID.
func (c *ExecutionClient) GetPurchaseOrder(ctx context.Context, orderID, organizationID, accessToken string) (*PurchaseOrder, error) {
	url := fmt.Sprintf("%s/api/v1/purchase-orders/%s?organization_id=%s", c.baseURL, orderID, organizationID)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("get purchase order failed with status %d: %v", resp.StatusCode, errResp)
	}

	var result struct {
		Order PurchaseOrder `json:"order"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result.Order, nil
}

// ReceiveGoods receives goods for a purchase order.
func (c *ExecutionClient) ReceiveGoods(ctx context.Context, orderID string, req ReceiveGoodsRequest, accessToken string) (*PurchaseOrder, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/purchase-orders/%s/receive", c.baseURL, orderID)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("receive goods failed with status %d: %v", resp.StatusCode, errResp)
	}

	var result struct {
		Order PurchaseOrder `json:"order"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result.Order, nil
}

// CancelPurchaseOrder cancels a purchase order.
func (c *ExecutionClient) CancelPurchaseOrder(ctx context.Context, orderID, accessToken string) error {
	url := fmt.Sprintf("%s/api/v1/purchase-orders/%s/cancel", c.baseURL, orderID)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return fmt.Errorf("cancel purchase order failed with status %d: %v", resp.StatusCode, errResp)
	}

	return nil
}

// CreateSalesOrder creates a new sales order.
func (c *ExecutionClient) CreateSalesOrder(ctx context.Context, req CreateSalesOrderRequest, accessToken string) (*CreateSalesOrderResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v1/sales-orders", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("create sales order failed with status %d: %v", resp.StatusCode, errResp)
	}

	var result CreateSalesOrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetSalesOrder gets a sales order by ID.
func (c *ExecutionClient) GetSalesOrder(ctx context.Context, orderID, organizationID, accessToken string) (*SalesOrder, error) {
	url := fmt.Sprintf("%s/api/v1/sales-orders/%s?organization_id=%s", c.baseURL, orderID, organizationID)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("get sales order failed with status %d: %v", resp.StatusCode, errResp)
	}

	var result struct {
		Order SalesOrder `json:"order"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result.Order, nil
}

// ShipOrder ships items from a sales order.
func (c *ExecutionClient) ShipOrder(ctx context.Context, orderID string, req ShipOrderRequest, accessToken string) (*SalesOrder, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/sales-orders/%s/ship", c.baseURL, orderID)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("ship order failed with status %d: %v", resp.StatusCode, errResp)
	}

	var result struct {
		Order SalesOrder `json:"order"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result.Order, nil
}

// CancelSalesOrder cancels a sales order.
func (c *ExecutionClient) CancelSalesOrder(ctx context.Context, orderID, accessToken string) error {
	url := fmt.Sprintf("%s/api/v1/sales-orders/%s/cancel", c.baseURL, orderID)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return fmt.Errorf("cancel sales order failed with status %d: %v", resp.StatusCode, errResp)
	}

	return nil
}

// GetInventoryBalances gets inventory balances for an organization.
func (c *ExecutionClient) GetInventoryBalances(ctx context.Context, organizationID, accessToken string) ([]InventoryBalance, error) {
	url := fmt.Sprintf("%s/api/v1/inventory/balances?organization_id=%s", c.baseURL, organizationID)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("get inventory balances failed with status %d: %v", resp.StatusCode, errResp)
	}

	var result struct {
		Balances []InventoryBalance `json:"balances"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Balances, nil
}

// GetInventoryTransactions gets inventory transactions for an organization.
func (c *ExecutionClient) GetInventoryTransactions(ctx context.Context, organizationID, accessToken string) ([]InventoryTransaction, error) {
	url := fmt.Sprintf("%s/api/v1/inventory/transactions?organization_id=%s", c.baseURL, organizationID)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("get inventory transactions failed with status %d: %v", resp.StatusCode, errResp)
	}

	var result struct {
		Transactions []InventoryTransaction `json:"transactions"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Transactions, nil
}

// HealthCheck checks if the execution service is healthy.
func (c *ExecutionClient) HealthCheck(ctx context.Context) error {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/health", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status %d", resp.StatusCode)
	}

	return nil
}
