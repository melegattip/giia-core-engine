// Package execution provides a gRPC client adapter for the Execution Service.
package execution

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

// Client implements the ExecutionServiceClient interface.
type Client struct {
	conn        *grpc.ClientConn
	address     string
	timeout     time.Duration
	retryConfig retry.Config
}

// ClientConfig holds the configuration for the Execution client.
type ClientConfig struct {
	Address    string
	Timeout    time.Duration
	MaxRetries int
}

// NewClient creates a new Execution service client.
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
		return nil, fmt.Errorf("failed to connect to execution service: %w", err)
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

// GetSalesData retrieves aggregated sales data for an organization within a date range.
func (c *Client) GetSalesData(ctx context.Context, organizationID uuid.UUID, startDate, endDate time.Time) (*providers.SalesData, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	return retry.DoWithResult(ctx, c.retryConfig, func() (*providers.SalesData, error) {
		// In a real implementation, this would call the Execution gRPC service
		// For now, we return mock data to allow the service to compile and run
		// The actual gRPC call would look like:
		// resp, err := c.executionClient.GetSalesData(ctx, &executionv1.GetSalesDataRequest{
		//     OrganizationId: organizationID.String(),
		//     StartDate: timestamppb.New(startDate),
		//     EndDate: timestamppb.New(endDate),
		// })

		// Mock implementation - in production, implement actual gRPC call
		return &providers.SalesData{
			OrganizationID: organizationID,
			StartDate:      startDate,
			EndDate:        endDate,
			TotalValue:     150000.00,
			OrderCount:     250,
		}, nil
	}, nil)
}

// GetInventorySnapshots retrieves inventory snapshots for an organization within a date range.
func (c *Client) GetInventorySnapshots(ctx context.Context, organizationID uuid.UUID, startDate, endDate time.Time) ([]*providers.InventorySnapshot, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	return retry.DoWithResult(ctx, c.retryConfig, func() ([]*providers.InventorySnapshot, error) {
		// Mock implementation - in production, implement actual gRPC call
		days := int(endDate.Sub(startDate).Hours() / 24)
		if days <= 0 {
			days = 1
		}

		snapshots := make([]*providers.InventorySnapshot, days)
		baseValue := 200000.00

		for i := 0; i < days; i++ {
			date := startDate.AddDate(0, 0, i)
			// Simulate some variation in inventory value
			variation := float64(i%10) * 1000
			snapshots[i] = &providers.InventorySnapshot{
				Date:       date,
				TotalValue: baseValue + variation,
				ProductID:  nil, // Total inventory, not per-product
			}
		}

		return snapshots, nil
	}, nil)
}

// GetProductSales retrieves per-product sales data for an organization within a date range.
func (c *Client) GetProductSales(ctx context.Context, organizationID uuid.UUID, startDate, endDate time.Time) ([]*providers.ProductSales, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	return retry.DoWithResult(ctx, c.retryConfig, func() ([]*providers.ProductSales, error) {
		// Mock implementation - in production, implement actual gRPC call
		return []*providers.ProductSales{
			{
				ProductID:     uuid.New(),
				SKU:           "TOP-SELLER-001",
				Name:          "Top Selling Product 1",
				Sales30Days:   25000.00,
				AvgStockValue: 5000.00,
			},
			{
				ProductID:     uuid.New(),
				SKU:           "TOP-SELLER-002",
				Name:          "Top Selling Product 2",
				Sales30Days:   20000.00,
				AvgStockValue: 4000.00,
			},
			{
				ProductID:     uuid.New(),
				SKU:           "SLOW-MOVER-001",
				Name:          "Slow Moving Product 1",
				Sales30Days:   500.00,
				AvgStockValue: 10000.00,
			},
			{
				ProductID:     uuid.New(),
				SKU:           "SLOW-MOVER-002",
				Name:          "Slow Moving Product 2",
				Sales30Days:   300.00,
				AvgStockValue: 8000.00,
			},
		}, nil
	}, nil)
}

// Ensure Client implements ExecutionServiceClient.
var _ providers.ExecutionServiceClient = (*Client)(nil)
