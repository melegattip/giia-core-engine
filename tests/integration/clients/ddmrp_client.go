// Package clients provides HTTP and gRPC clients for integration testing.
package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// DDMRPClient provides methods to interact with the DDMRP Engine Service.
type DDMRPClient struct {
	httpBaseURL string
	grpcURL     string
	httpClient  *http.Client
	grpcConn    *grpc.ClientConn
}

// NewDDMRPClient creates a new DDMRPClient.
func NewDDMRPClient(httpBaseURL, grpcURL string) *DDMRPClient {
	return &DDMRPClient{
		httpBaseURL: httpBaseURL,
		grpcURL:     grpcURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Buffer represents a DDMRP buffer.
type Buffer struct {
	ID                 string    `json:"id"`
	ProductID          string    `json:"product_id"`
	OrganizationID     string    `json:"organization_id"`
	BufferProfileID    string    `json:"buffer_profile_id"`
	CPD                float64   `json:"cpd"`
	LTD                int       `json:"ltd"`
	RedBase            float64   `json:"red_base"`
	RedSafe            float64   `json:"red_safe"`
	RedZone            float64   `json:"red_zone"`
	YellowZone         float64   `json:"yellow_zone"`
	GreenZone          float64   `json:"green_zone"`
	TopOfRed           float64   `json:"top_of_red"`
	TopOfYellow        float64   `json:"top_of_yellow"`
	TopOfGreen         float64   `json:"top_of_green"`
	OnHand             float64   `json:"on_hand"`
	OnOrder            float64   `json:"on_order"`
	QualifiedDemand    float64   `json:"qualified_demand"`
	NetFlowPosition    float64   `json:"net_flow_position"`
	BufferPenetration  float64   `json:"buffer_penetration"`
	Zone               string    `json:"zone"`
	AlertLevel         string    `json:"alert_level"`
	LastRecalculatedAt time.Time `json:"last_recalculated_at"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// BufferProfile represents a DDMRP buffer profile.
type BufferProfile struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	OrganizationID    string  `json:"organization_id"`
	LeadTimeFactor    float64 `json:"lead_time_factor"`
	VariabilityFactor float64 `json:"variability_factor"`
}

// DemandAdjustment represents a Flow Adjustment Demand (FAD).
type DemandAdjustment struct {
	ID             string    `json:"id"`
	ProductID      string    `json:"product_id"`
	OrganizationID string    `json:"organization_id"`
	StartDate      time.Time `json:"start_date"`
	EndDate        time.Time `json:"end_date"`
	AdjustmentType string    `json:"adjustment_type"`
	Factor         float64   `json:"factor"`
	Reason         string    `json:"reason"`
	CreatedAt      time.Time `json:"created_at"`
	CreatedBy      string    `json:"created_by"`
}

// CalculateBufferRequest represents a request to calculate a buffer.
type CalculateBufferRequest struct {
	ProductID      string `json:"product_id"`
	OrganizationID string `json:"organization_id"`
}

// UpdateNFPRequest represents a request to update Net Flow Position.
type UpdateNFPRequest struct {
	ProductID       string  `json:"product_id"`
	OrganizationID  string  `json:"organization_id"`
	OnHand          float64 `json:"on_hand"`
	OnOrder         float64 `json:"on_order"`
	QualifiedDemand float64 `json:"qualified_demand"`
}

// CreateFADRequest represents a request to create a FAD.
type CreateFADRequest struct {
	ProductID      string    `json:"product_id"`
	OrganizationID string    `json:"organization_id"`
	StartDate      time.Time `json:"start_date"`
	EndDate        time.Time `json:"end_date"`
	AdjustmentType string    `json:"adjustment_type"`
	Factor         float64   `json:"factor"`
	Reason         string    `json:"reason"`
	CreatedBy      string    `json:"created_by"`
}

// Connect establishes a gRPC connection.
func (c *DDMRPClient) Connect(ctx context.Context) error {
	conn, err := grpc.DialContext(ctx, c.grpcURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to DDMRP gRPC server: %w", err)
	}
	c.grpcConn = conn
	return nil
}

// Close closes the gRPC connection.
func (c *DDMRPClient) Close() error {
	if c.grpcConn != nil {
		return c.grpcConn.Close()
	}
	return nil
}

// GetBuffer gets a buffer by product ID (via HTTP).
func (c *DDMRPClient) GetBuffer(ctx context.Context, productID, organizationID, accessToken string) (*Buffer, error) {
	url := fmt.Sprintf("%s/api/v1/buffers/%s?organization_id=%s", c.httpBaseURL, productID, organizationID)

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
		return nil, fmt.Errorf("get buffer failed with status %d: %v", resp.StatusCode, errResp)
	}

	var result struct {
		Buffer Buffer `json:"buffer"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result.Buffer, nil
}

// ListBuffers lists all buffers for an organization.
func (c *DDMRPClient) ListBuffers(ctx context.Context, organizationID, accessToken string) ([]Buffer, error) {
	url := fmt.Sprintf("%s/api/v1/buffers?organization_id=%s", c.httpBaseURL, organizationID)

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
		return nil, fmt.Errorf("list buffers failed with status %d: %v", resp.StatusCode, errResp)
	}

	var result struct {
		Buffers []Buffer `json:"buffers"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Buffers, nil
}

// CalculateBuffer triggers buffer calculation.
func (c *DDMRPClient) CalculateBuffer(ctx context.Context, req CalculateBufferRequest, accessToken string) (*Buffer, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.httpBaseURL+"/api/v1/buffers/calculate", bytes.NewBuffer(body))
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

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("calculate buffer failed with status %d: %v", resp.StatusCode, errResp)
	}

	var result struct {
		Buffer Buffer `json:"buffer"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result.Buffer, nil
}

// UpdateNFP updates Net Flow Position.
func (c *DDMRPClient) UpdateNFP(ctx context.Context, req UpdateNFPRequest, accessToken string) (*Buffer, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.httpBaseURL+"/api/v1/buffers/nfp", bytes.NewBuffer(body))
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
		return nil, fmt.Errorf("update NFP failed with status %d: %v", resp.StatusCode, errResp)
	}

	var result struct {
		Buffer Buffer `json:"buffer"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result.Buffer, nil
}

// CreateFAD creates a Flow Adjustment Demand.
func (c *DDMRPClient) CreateFAD(ctx context.Context, req CreateFADRequest, accessToken string) (*DemandAdjustment, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.httpBaseURL+"/api/v1/fads", bytes.NewBuffer(body))
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

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("create FAD failed with status %d: %v", resp.StatusCode, errResp)
	}

	var result struct {
		DemandAdjustment DemandAdjustment `json:"demand_adjustment"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result.DemandAdjustment, nil
}

// ListFADs lists FADs for a product.
func (c *DDMRPClient) ListFADs(ctx context.Context, productID, organizationID, accessToken string) ([]DemandAdjustment, error) {
	url := fmt.Sprintf("%s/api/v1/fads?product_id=%s&organization_id=%s", c.httpBaseURL, productID, organizationID)

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
		return nil, fmt.Errorf("list FADs failed with status %d: %v", resp.StatusCode, errResp)
	}

	var result struct {
		DemandAdjustments []DemandAdjustment `json:"demand_adjustments"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.DemandAdjustments, nil
}

// CheckReplenishment checks buffers needing replenishment.
func (c *DDMRPClient) CheckReplenishment(ctx context.Context, organizationID, accessToken string) ([]Buffer, error) {
	url := fmt.Sprintf("%s/api/v1/buffers/replenishment?organization_id=%s", c.httpBaseURL, organizationID)

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
		return nil, fmt.Errorf("check replenishment failed with status %d: %v", resp.StatusCode, errResp)
	}

	var result struct {
		Buffers []Buffer `json:"buffers"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Buffers, nil
}

// HealthCheck checks if the DDMRP service is healthy.
func (c *DDMRPClient) HealthCheck(ctx context.Context) error {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", c.httpBaseURL+"/health", nil)
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
