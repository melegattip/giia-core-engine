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

// CatalogClient provides methods to interact with the Catalog Service.
type CatalogClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewCatalogClient creates a new CatalogClient.
func NewCatalogClient(baseURL string) *CatalogClient {
	return &CatalogClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Product represents a product in the catalog.
type Product struct {
	ID              string    `json:"id"`
	OrganizationID  string    `json:"organization_id"`
	SKU             string    `json:"sku"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	Category        string    `json:"category"`
	UnitOfMeasure   string    `json:"unit_of_measure"`
	Status          string    `json:"status"`
	BufferProfileID string    `json:"buffer_profile_id,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// CreateProductRequest represents a request to create a product.
type CreateProductRequest struct {
	OrganizationID  string `json:"organization_id"`
	SKU             string `json:"sku"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	Category        string `json:"category"`
	UnitOfMeasure   string `json:"unit_of_measure"`
	BufferProfileID string `json:"buffer_profile_id,omitempty"`
}

// CreateProductResponse represents a response from creating a product.
type CreateProductResponse struct {
	Product Product `json:"product"`
}

// GetProductResponse represents a response from getting a product.
type GetProductResponse struct {
	Product Product `json:"product"`
}

// ListProductsResponse represents a response from listing products.
type ListProductsResponse struct {
	Products []Product `json:"products"`
	Total    int       `json:"total"`
	Page     int       `json:"page"`
	PageSize int       `json:"page_size"`
}

// UpdateProductRequest represents a request to update a product.
type UpdateProductRequest struct {
	ID              string `json:"id"`
	OrganizationID  string `json:"organization_id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	Category        string `json:"category"`
	UnitOfMeasure   string `json:"unit_of_measure"`
	Status          string `json:"status"`
	BufferProfileID string `json:"buffer_profile_id,omitempty"`
}

// UpdateProductResponse represents a response from updating a product.
type UpdateProductResponse struct {
	Product Product `json:"product"`
}

// CreateProduct creates a new product.
func (c *CatalogClient) CreateProduct(ctx context.Context, req CreateProductRequest, accessToken string) (*CreateProductResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v1/products", bytes.NewBuffer(body))
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
		return nil, fmt.Errorf("create product failed with status %d: %v", resp.StatusCode, errResp)
	}

	var result CreateProductResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetProduct gets a product by ID.
func (c *CatalogClient) GetProduct(ctx context.Context, productID, organizationID, accessToken string) (*GetProductResponse, error) {
	url := fmt.Sprintf("%s/api/v1/products/%s?organization_id=%s", c.baseURL, productID, organizationID)

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
		return nil, fmt.Errorf("get product failed with status %d: %v", resp.StatusCode, errResp)
	}

	var result GetProductResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// ListProducts lists products for an organization.
func (c *CatalogClient) ListProducts(ctx context.Context, organizationID, accessToken string, page, pageSize int) (*ListProductsResponse, error) {
	url := fmt.Sprintf("%s/api/v1/products?organization_id=%s&page=%d&page_size=%d", c.baseURL, organizationID, page, pageSize)

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
		return nil, fmt.Errorf("list products failed with status %d: %v", resp.StatusCode, errResp)
	}

	var result ListProductsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// UpdateProduct updates a product.
func (c *CatalogClient) UpdateProduct(ctx context.Context, req UpdateProductRequest, accessToken string) (*UpdateProductResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/products/%s", c.baseURL, req.ID)
	httpReq, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(body))
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
		return nil, fmt.Errorf("update product failed with status %d: %v", resp.StatusCode, errResp)
	}

	var result UpdateProductResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// DeleteProduct deletes a product.
func (c *CatalogClient) DeleteProduct(ctx context.Context, productID, organizationID, accessToken string) error {
	url := fmt.Sprintf("%s/api/v1/products/%s?organization_id=%s", c.baseURL, productID, organizationID)

	httpReq, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return fmt.Errorf("delete product failed with status %d: %v", resp.StatusCode, errResp)
	}

	return nil
}

// SearchProducts searches for products.
func (c *CatalogClient) SearchProducts(ctx context.Context, organizationID, query, accessToken string) (*ListProductsResponse, error) {
	url := fmt.Sprintf("%s/api/v1/products/search?organization_id=%s&query=%s", c.baseURL, organizationID, query)

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
		return nil, fmt.Errorf("search products failed with status %d: %v", resp.StatusCode, errResp)
	}

	var result ListProductsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// HealthCheck checks if the catalog service is healthy.
func (c *CatalogClient) HealthCheck(ctx context.Context) error {
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
