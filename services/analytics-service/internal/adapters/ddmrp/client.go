// Package ddmrp provides a gRPC client adapter for the DDMRP Engine Service.
package ddmrp

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/melegattip/giia-core-engine/services/analytics-service/internal/adapters/retry"
	"github.com/melegattip/giia-core-engine/services/analytics-service/internal/core/providers"
)

// Client implements the DDMRPServiceClient interface.
type Client struct {
	conn        *grpc.ClientConn
	address     string
	timeout     time.Duration
	retryConfig retry.Config
}

// ClientConfig holds the configuration for the DDMRP client.
type ClientConfig struct {
	Address    string
	Timeout    time.Duration
	MaxRetries int
}

// NewClient creates a new DDMRP service client.
func NewClient(config *ClientConfig) (*Client, error) {
	if config.Timeout == 0 {
		config.Timeout = 5 * time.Second
	}

	retryConfig := retry.DefaultConfig()
	if config.MaxRetries > 0 {
		retryConfig.MaxRetries = config.MaxRetries
	}

	conn, err := grpc.NewClient(
		config.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ddmrp service: %w", err)
	}

	return &Client{
		conn:        conn,
		address:     config.Address,
		timeout:     config.Timeout,
		retryConfig: retryConfig,
	}, nil
}

// Close closes the gRPC connection.
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// GetBufferHistory retrieves buffer history for a product on a specific date.
func (c *Client) GetBufferHistory(ctx context.Context, organizationID, productID uuid.UUID, date time.Time) (*providers.BufferHistory, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	return retry.DoWithResult(ctx, c.retryConfig, func() (*providers.BufferHistory, error) {
		// In a real implementation, this would call the DDMRP gRPC service
		// For now, we return mock data to allow the service to compile and run
		// The actual gRPC call would look like:
		// resp, err := c.ddmrpClient.GetBufferHistory(ctx, &ddmrpv1.GetBufferHistoryRequest{
		//     OrganizationId: organizationID.String(),
		//     ProductId: productID.String(),
		//     Date: timestamppb.New(date),
		// })

		// Mock implementation - in production, implement actual gRPC call
		return &providers.BufferHistory{
			ProductID:         productID,
			OrganizationID:    organizationID,
			Date:              date,
			CPD:               10.5,
			RedZone:           100,
			RedBase:           50,
			RedSafe:           50,
			YellowZone:        200,
			GreenZone:         150,
			LTD:               14,
			LeadTimeFactor:    1.0,
			VariabilityFactor: 0.5,
			MOQ:               10,
			OrderFrequency:    7,
			HasAdjustments:    false,
		}, nil
	}, nil)
}

// ListBufferHistory retrieves buffer history for an organization within a date range.
func (c *Client) ListBufferHistory(ctx context.Context, organizationID uuid.UUID, startDate, endDate time.Time) ([]*providers.BufferHistory, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	return retry.DoWithResult(ctx, c.retryConfig, func() ([]*providers.BufferHistory, error) {
		// Mock implementation - in production, implement actual gRPC call
		return []*providers.BufferHistory{
			{
				ProductID:         uuid.New(),
				OrganizationID:    organizationID,
				Date:              startDate,
				CPD:               10.5,
				RedZone:           100,
				RedBase:           50,
				RedSafe:           50,
				YellowZone:        200,
				GreenZone:         150,
				LTD:               14,
				LeadTimeFactor:    1.0,
				VariabilityFactor: 0.5,
				MOQ:               10,
				OrderFrequency:    7,
				HasAdjustments:    false,
			},
			{
				ProductID:         uuid.New(),
				OrganizationID:    organizationID,
				Date:              startDate.AddDate(0, 0, 1),
				CPD:               12.0,
				RedZone:           110,
				RedBase:           55,
				RedSafe:           55,
				YellowZone:        220,
				GreenZone:         165,
				LTD:               14,
				LeadTimeFactor:    1.0,
				VariabilityFactor: 0.5,
				MOQ:               10,
				OrderFrequency:    7,
				HasAdjustments:    true,
			},
		}, nil
	}, nil)
}

// GetBufferZoneDistribution retrieves the distribution of products across buffer zones.
func (c *Client) GetBufferZoneDistribution(ctx context.Context, organizationID uuid.UUID, date time.Time) (*providers.BufferZoneDistribution, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	return retry.DoWithResult(ctx, c.retryConfig, func() (*providers.BufferZoneDistribution, error) {
		// Mock implementation - in production, implement actual gRPC call
		return &providers.BufferZoneDistribution{
			OrganizationID: organizationID,
			Date:           date,
			GreenCount:     60,
			YellowCount:    30,
			RedCount:       10,
			TotalProducts:  100,
		}, nil
	}, nil)
}

// Ensure Client implements DDMRPServiceClient.
var _ providers.DDMRPServiceClient = (*Client)(nil)
