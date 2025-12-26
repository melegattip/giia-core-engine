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

// AnalyticsClient provides methods to interact with the Analytics Service.
type AnalyticsClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewAnalyticsClient creates a new AnalyticsClient.
func NewAnalyticsClient(baseURL string) *AnalyticsClient {
	return &AnalyticsClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// DaysInInventoryKPI represents Days in Inventory KPI.
type DaysInInventoryKPI struct {
	ProductID      string    `json:"product_id"`
	OrganizationID string    `json:"organization_id"`
	DaysInStock    float64   `json:"days_in_stock"`
	AverageDemand  float64   `json:"average_demand"`
	CurrentOnHand  float64   `json:"current_on_hand"`
	CalculatedAt   time.Time `json:"calculated_at"`
}

// ImmobilizedInventoryKPI represents Immobilized Inventory KPI.
type ImmobilizedInventoryKPI struct {
	ProductID          string    `json:"product_id"`
	OrganizationID     string    `json:"organization_id"`
	ImmobilizedValue   float64   `json:"immobilized_value"`
	ImmobilizedPercent float64   `json:"immobilized_percent"`
	DaysSinceMovement  int       `json:"days_since_movement"`
	CalculatedAt       time.Time `json:"calculated_at"`
}

// InventoryRotationKPI represents Inventory Rotation KPI.
type InventoryRotationKPI struct {
	ProductID      string    `json:"product_id"`
	OrganizationID string    `json:"organization_id"`
	RotationIndex  float64   `json:"rotation_index"`
	TurnoverRate   float64   `json:"turnover_rate"`
	Period         string    `json:"period"`
	CalculatedAt   time.Time `json:"calculated_at"`
}

// BufferAnalytics represents buffer analytics data.
type BufferAnalytics struct {
	ProductID         string    `json:"product_id"`
	OrganizationID    string    `json:"organization_id"`
	BufferPenetration float64   `json:"buffer_penetration"`
	Zone              string    `json:"zone"`
	AlertLevel        string    `json:"alert_level"`
	NetFlowPosition   float64   `json:"net_flow_position"`
	ProjectedDemand   float64   `json:"projected_demand"`
	CalculatedAt      time.Time `json:"calculated_at"`
}

// AnalyticsSnapshot represents a point-in-time analytics snapshot.
type AnalyticsSnapshot struct {
	ID                       string    `json:"id"`
	OrganizationID           string    `json:"organization_id"`
	TotalProducts            int       `json:"total_products"`
	ProductsInRed            int       `json:"products_in_red"`
	ProductsInYellow         int       `json:"products_in_yellow"`
	ProductsInGreen          int       `json:"products_in_green"`
	AverageBufferPenetration float64   `json:"average_buffer_penetration"`
	TotalInventoryValue      float64   `json:"total_inventory_value"`
	ImmobilizedValue         float64   `json:"immobilized_value"`
	SnapshotTime             time.Time `json:"snapshot_time"`
	CreatedAt                time.Time `json:"created_at"`
}

// GetDaysInInventory gets Days in Inventory KPI.
func (c *AnalyticsClient) GetDaysInInventory(ctx context.Context, organizationID, accessToken string) ([]DaysInInventoryKPI, error) {
	url := fmt.Sprintf("%s/api/v1/analytics/days-in-inventory?organization_id=%s", c.baseURL, organizationID)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("X-Organization-ID", organizationID)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("get days in inventory failed with status %d: %v", resp.StatusCode, errResp)
	}

	var result struct {
		Metrics []DaysInInventoryKPI `json:"metrics"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Metrics, nil
}

// GetImmobilizedInventory gets Immobilized Inventory KPI.
func (c *AnalyticsClient) GetImmobilizedInventory(ctx context.Context, organizationID, accessToken string) ([]ImmobilizedInventoryKPI, error) {
	url := fmt.Sprintf("%s/api/v1/analytics/immobilized-inventory?organization_id=%s", c.baseURL, organizationID)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("X-Organization-ID", organizationID)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("get immobilized inventory failed with status %d: %v", resp.StatusCode, errResp)
	}

	var result struct {
		Metrics []ImmobilizedInventoryKPI `json:"metrics"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Metrics, nil
}

// GetInventoryRotation gets Inventory Rotation KPI.
func (c *AnalyticsClient) GetInventoryRotation(ctx context.Context, organizationID, accessToken string) ([]InventoryRotationKPI, error) {
	url := fmt.Sprintf("%s/api/v1/analytics/inventory-rotation?organization_id=%s", c.baseURL, organizationID)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("X-Organization-ID", organizationID)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("get inventory rotation failed with status %d: %v", resp.StatusCode, errResp)
	}

	var result struct {
		Metrics []InventoryRotationKPI `json:"metrics"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Metrics, nil
}

// GetBufferAnalytics gets buffer analytics.
func (c *AnalyticsClient) GetBufferAnalytics(ctx context.Context, organizationID, accessToken string) ([]BufferAnalytics, error) {
	url := fmt.Sprintf("%s/api/v1/analytics/buffer-analytics?organization_id=%s", c.baseURL, organizationID)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("X-Organization-ID", organizationID)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("get buffer analytics failed with status %d: %v", resp.StatusCode, errResp)
	}

	var result struct {
		Analytics []BufferAnalytics `json:"analytics"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Analytics, nil
}

// GetSnapshot gets the latest analytics snapshot.
func (c *AnalyticsClient) GetSnapshot(ctx context.Context, organizationID, accessToken string) (*AnalyticsSnapshot, error) {
	url := fmt.Sprintf("%s/api/v1/analytics/snapshot?organization_id=%s", c.baseURL, organizationID)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("X-Organization-ID", organizationID)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("get snapshot failed with status %d: %v", resp.StatusCode, errResp)
	}

	var result struct {
		Snapshot AnalyticsSnapshot `json:"snapshot"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result.Snapshot, nil
}

// SyncBufferData triggers synchronization of buffer data.
func (c *AnalyticsClient) SyncBufferData(ctx context.Context, organizationID, accessToken string) error {
	body, err := json.Marshal(map[string]string{
		"organization_id": organizationID,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v1/analytics/sync-buffer", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("X-Organization-ID", organizationID)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return fmt.Errorf("sync buffer data failed with status %d: %v", resp.StatusCode, errResp)
	}

	return nil
}

// HealthCheck checks if the analytics service is healthy.
func (c *AnalyticsClient) HealthCheck(ctx context.Context) error {
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
