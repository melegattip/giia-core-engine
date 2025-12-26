// Package main demonstrates a complete workflow with the GIIA Platform APIs.
// This example shows: authentication, product creation, purchase order workflow,
// buffer checking, and notification subscription.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Configuration
var (
	baseURL     = getEnv("GIIA_API_URL", "http://localhost")
	authPort    = ":8081"
	catalogPort = ":8082"
	ddmrpPort   = ":8083"
	execPort    = ":8084"
)

// GIIAClient is the main client for GIIA APIs
type GIIAClient struct {
	accessToken string
	orgID       string
	httpClient  *http.Client
}

// LoginResponse represents the login response
type LoginResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	User        struct {
		ID             string   `json:"id"`
		Email          string   `json:"email"`
		OrganizationID string   `json:"organization_id"`
		Roles          []string `json:"roles"`
	} `json:"user"`
}

// Product represents a product
type Product struct {
	ID            string `json:"id"`
	SKU           string `json:"sku"`
	Name          string `json:"name"`
	Category      string `json:"category"`
	UnitOfMeasure string `json:"unit_of_measure"`
	Status        string `json:"status"`
}

// Buffer represents DDMRP buffer status
type Buffer struct {
	ProductID         string  `json:"product_id"`
	Zone              string  `json:"zone"`
	NetFlowPosition   float64 `json:"net_flow_position"`
	BufferPenetration float64 `json:"buffer_penetration"`
	RedZone           float64 `json:"red_zone"`
	YellowZone        float64 `json:"yellow_zone"`
	GreenZone         float64 `json:"green_zone"`
}

func main() {
	ctx := context.Background()

	// Step 1: Authenticate
	fmt.Println("=== Step 1: Authentication ===")
	client, err := authenticate(ctx, getEnv("GIIA_EMAIL", ""), getEnv("GIIA_PASSWORD", ""))
	if err != nil {
		fmt.Printf("Authentication failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Logged in as user, org: %s\n", client.orgID)

	// Step 2: Create a product
	fmt.Println("\n=== Step 2: Create Product ===")
	product, err := client.CreateProduct(ctx, Product{
		SKU:           fmt.Sprintf("DEMO-%d", time.Now().Unix()),
		Name:          "Demo Widget",
		Category:      "Electronics",
		UnitOfMeasure: "units",
	})
	if err != nil {
		fmt.Printf("Failed to create product: %v\n", err)
	} else {
		fmt.Printf("✓ Created product: %s (%s)\n", product.Name, product.ID)
	}

	// Step 3: List products
	fmt.Println("\n=== Step 3: List Products ===")
	products, err := client.ListProducts(ctx)
	if err != nil {
		fmt.Printf("Failed to list products: %v\n", err)
	} else {
		fmt.Printf("✓ Found %d products\n", len(products))
		for _, p := range products[:min(3, len(products))] {
			fmt.Printf("  - %s: %s\n", p.SKU, p.Name)
		}
	}

	// Step 4: Check buffer status
	fmt.Println("\n=== Step 4: Check Buffer Status ===")
	if len(products) > 0 {
		buffer, err := client.GetBuffer(ctx, products[0].ID)
		if err != nil {
			fmt.Printf("Failed to get buffer: %v\n", err)
		} else {
			fmt.Printf("✓ Buffer for %s:\n", products[0].SKU)
			fmt.Printf("  Zone: %s\n", buffer.Zone)
			fmt.Printf("  NFP: %.2f\n", buffer.NetFlowPosition)
			fmt.Printf("  Penetration: %.1f%%\n", buffer.BufferPenetration*100)
		}
	}

	// Step 5: Create purchase order
	fmt.Println("\n=== Step 5: Create Purchase Order ===")
	po, err := client.CreatePurchaseOrder(ctx, map[string]interface{}{
		"po_number":             fmt.Sprintf("PO-%d", time.Now().Unix()),
		"supplier_id":           "00000000-0000-0000-0000-000000000001", // Demo supplier
		"order_date":            time.Now().Format("2006-01-02"),
		"expected_arrival_date": time.Now().AddDate(0, 0, 14).Format("2006-01-02"),
		"line_items": []map[string]interface{}{
			{
				"product_id": products[0].ID,
				"quantity":   100,
				"unit_cost":  25.50,
			},
		},
	})
	if err != nil {
		fmt.Printf("Failed to create PO: %v\n", err)
	} else {
		fmt.Printf("✓ Created PO: %v\n", po["po_number"])
	}

	fmt.Println("\n=== Workflow Complete ===")
}

func authenticate(ctx context.Context, email, password string) (*GIIAClient, error) {
	payload := map[string]string{
		"email":    email,
		"password": password,
	}
	body, _ := json.Marshal(payload)

	resp, err := http.Post(
		baseURL+authPort+"/api/v1/auth/login",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("login failed: %s", string(body))
	}

	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return nil, err
	}

	return &GIIAClient{
		accessToken: loginResp.AccessToken,
		orgID:       loginResp.User.OrganizationID,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (c *GIIAClient) doRequest(ctx context.Context, method, url string, body interface{}) (map[string]interface{}, error) {
	var reqBody io.Reader
	if body != nil {
		data, _ := json.Marshal(body)
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("X-Organization-ID", c.orgID)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error %d: %v", resp.StatusCode, result)
	}

	return result, nil
}

func (c *GIIAClient) CreateProduct(ctx context.Context, product Product) (*Product, error) {
	result, err := c.doRequest(ctx, "POST", baseURL+catalogPort+"/api/v1/products", product)
	if err != nil {
		return nil, err
	}

	data, _ := json.Marshal(result)
	var p Product
	json.Unmarshal(data, &p)
	return &p, nil
}

func (c *GIIAClient) ListProducts(ctx context.Context) ([]Product, error) {
	result, err := c.doRequest(ctx, "GET", baseURL+catalogPort+"/api/v1/products", nil)
	if err != nil {
		return nil, err
	}

	productsData, ok := result["products"].([]interface{})
	if !ok {
		return []Product{}, nil
	}

	products := make([]Product, 0, len(productsData))
	for _, p := range productsData {
		data, _ := json.Marshal(p)
		var product Product
		json.Unmarshal(data, &product)
		products = append(products, product)
	}
	return products, nil
}

func (c *GIIAClient) GetBuffer(ctx context.Context, productID string) (*Buffer, error) {
	result, err := c.doRequest(ctx, "GET", baseURL+ddmrpPort+"/api/v1/buffers/"+productID, nil)
	if err != nil {
		return nil, err
	}

	bufferData, ok := result["buffer"].(map[string]interface{})
	if !ok {
		return &Buffer{}, nil
	}

	data, _ := json.Marshal(bufferData)
	var buffer Buffer
	json.Unmarshal(data, &buffer)
	return &buffer, nil
}

func (c *GIIAClient) CreatePurchaseOrder(ctx context.Context, po map[string]interface{}) (map[string]interface{}, error) {
	return c.doRequest(ctx, "POST", baseURL+execPort+"/api/v1/purchase-orders", po)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
