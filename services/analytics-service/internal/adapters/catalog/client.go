// Package catalog provides a gRPC client adapter for the Catalog Service.
package catalog

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

// Client implements the CatalogServiceClient interface.
type Client struct {
	conn        *grpc.ClientConn
	address     string
	timeout     time.Duration
	retryConfig retry.Config
}

// ClientConfig holds the configuration for the Catalog client.
type ClientConfig struct {
	Address    string
	Timeout    time.Duration
	MaxRetries int
}

// NewClient creates a new Catalog service client.
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
		return nil, fmt.Errorf("failed to connect to catalog service: %w", err)
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

// ListProductsWithInventory retrieves all products with inventory data for an organization.
func (c *Client) ListProductsWithInventory(ctx context.Context, organizationID uuid.UUID) ([]*providers.ProductWithInventory, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	return retry.DoWithResult(ctx, c.retryConfig, func() ([]*providers.ProductWithInventory, error) {
		// In a real implementation, this would call the Catalog gRPC service
		// For now, we return mock data to allow the service to compile and run
		// The actual gRPC call would look like:
		// resp, err := c.catalogClient.ListProducts(ctx, &catalogv1.ListProductsRequest{
		//     OrganizationId: organizationID.String(),
		// })

		// Mock implementation - in production, implement actual gRPC call
		return []*providers.ProductWithInventory{
			{
				ProductID:        uuid.New(),
				SKU:              "MOCK-SKU-001",
				Name:             "Mock Product 1",
				Category:         "Category A",
				Quantity:         100,
				StandardCost:     25.50,
				LastPurchaseDate: ptrTime(time.Now().AddDate(0, -6, 0)),
				LastSaleDate:     ptrTime(time.Now().AddDate(0, -1, 0)),
			},
			{
				ProductID:        uuid.New(),
				SKU:              "MOCK-SKU-002",
				Name:             "Mock Product 2",
				Category:         "Category B",
				Quantity:         50,
				StandardCost:     45.00,
				LastPurchaseDate: ptrTime(time.Now().AddDate(-2, 0, 0)),
				LastSaleDate:     ptrTime(time.Now().AddDate(-1, 0, 0)),
			},
		}, nil
	}, nil)
}

// GetProduct retrieves a single product by ID.
func (c *Client) GetProduct(ctx context.Context, productID uuid.UUID) (*providers.ProductWithInventory, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	return retry.DoWithResult(ctx, c.retryConfig, func() (*providers.ProductWithInventory, error) {
		// Mock implementation - in production, implement actual gRPC call
		return &providers.ProductWithInventory{
			ProductID:        productID,
			SKU:              "MOCK-SKU",
			Name:             "Mock Product",
			Category:         "Category A",
			Quantity:         100,
			StandardCost:     25.50,
			LastPurchaseDate: ptrTime(time.Now().AddDate(0, -6, 0)),
			LastSaleDate:     ptrTime(time.Now().AddDate(0, -1, 0)),
		}, nil
	}, nil)
}

// ptrTime returns a pointer to a time.Time.
func ptrTime(t time.Time) *time.Time {
	return &t
}

// Ensure Client implements CatalogServiceClient.
var _ providers.CatalogServiceClient = (*Client)(nil)
