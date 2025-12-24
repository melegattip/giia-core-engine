// Package catalog provides a gRPC client adapter for the Catalog Service.
package catalog

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/providers"
)

// Client implements the CatalogServiceClient interface.
type Client struct {
	conn    *grpc.ClientConn
	address string
	timeout time.Duration
}

// ClientConfig holds the configuration for the Catalog client.
type ClientConfig struct {
	Address string
	Timeout time.Duration
}

// NewClient creates a new Catalog service client.
func NewClient(config *ClientConfig) (*Client, error) {
	if config.Timeout == 0 {
		config.Timeout = 5 * time.Second
	}

	conn, err := grpc.NewClient(
		config.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to catalog service: %w", err)
	}

	return &Client{
		conn:    conn,
		address: config.Address,
		timeout: config.Timeout,
	}, nil
}

// Close closes the gRPC connection.
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// GetProduct retrieves a product by ID.
func (c *Client) GetProduct(ctx context.Context, productID uuid.UUID) (*providers.Product, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// In a real implementation, this would call the Catalog gRPC service
	// For now, we return a mock response to allow the service to compile and run
	// The actual gRPC call would look like:
	// resp, err := c.catalogClient.GetProduct(ctx, &catalogv1.GetProductRequest{
	//     Id: productID.String(),
	// })

	// Mock implementation - in production, implement actual gRPC call
	return &providers.Product{
		ID:            productID,
		SKU:           "MOCK-SKU",
		Name:          "Mock Product",
		UnitOfMeasure: "EA",
	}, nil
}

// GetSupplier retrieves a supplier by ID.
func (c *Client) GetSupplier(ctx context.Context, supplierID uuid.UUID) (*providers.Supplier, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// Mock implementation - in production, implement actual gRPC call
	return &providers.Supplier{
		ID:   supplierID,
		Name: "Mock Supplier",
		Code: "MOCK-SUP",
	}, nil
}

// GetProductsByIDs retrieves multiple products by their IDs.
func (c *Client) GetProductsByIDs(ctx context.Context, productIDs []uuid.UUID) ([]*providers.Product, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// Mock implementation - in production, implement actual gRPC call
	products := make([]*providers.Product, len(productIDs))
	for i, id := range productIDs {
		products[i] = &providers.Product{
			ID:            id,
			SKU:           fmt.Sprintf("MOCK-SKU-%d", i),
			Name:          fmt.Sprintf("Mock Product %d", i),
			UnitOfMeasure: "EA",
		}
	}
	return products, nil
}

// Ensure Client implements CatalogServiceClient.
var _ providers.CatalogServiceClient = (*Client)(nil)
