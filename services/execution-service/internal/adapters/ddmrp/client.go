// Package ddmrp provides a gRPC client adapter for the DDMRP Engine Service.
package ddmrp

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/providers"
)

// Client implements the DDMRPServiceClient interface.
type Client struct {
	conn    *grpc.ClientConn
	address string
	timeout time.Duration
}

// ClientConfig holds the configuration for the DDMRP client.
type ClientConfig struct {
	Address string
	Timeout time.Duration
}

// NewClient creates a new DDMRP Engine service client.
func NewClient(config *ClientConfig) (*Client, error) {
	if config.Timeout == 0 {
		config.Timeout = 5 * time.Second
	}

	conn, err := grpc.NewClient(
		config.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ddmrp service: %w", err)
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

// GetBufferStatus retrieves the buffer status for a product.
func (c *Client) GetBufferStatus(ctx context.Context, organizationID, productID uuid.UUID) (*providers.BufferStatus, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// In a real implementation, this would call the DDMRP gRPC service
	// For now, we return a mock response to allow the service to compile and run
	// The actual gRPC call would look like:
	// resp, err := c.ddmrpClient.GetBuffer(ctx, &ddmrpv1.GetBufferRequest{
	//     ProductId:      productID.String(),
	//     OrganizationId: organizationID.String(),
	// })

	// Mock implementation - in production, implement actual gRPC call
	return &providers.BufferStatus{
		ProductID:       productID,
		Zone:            "green",
		NetFlowPosition: 100.0,
		TopOfGreen:      150.0,
		AlertLevel:      "none",
	}, nil
}

// UpdateNetFlowPosition triggers a recalculation of the Net Flow Position.
func (c *Client) UpdateNetFlowPosition(ctx context.Context, organizationID, productID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// In a real implementation, this would call the DDMRP gRPC service
	// For now, we just return nil to allow the service to compile and run
	// The actual gRPC call would look like:
	// _, err := c.ddmrpClient.UpdateNFP(ctx, &ddmrpv1.UpdateNFPRequest{
	//     ProductId:      productID.String(),
	//     OrganizationId: organizationID.String(),
	// })

	// Mock implementation - in production, implement actual gRPC call
	return nil
}

// GetProductsInRedZone retrieves all products in the red zone.
func (c *Client) GetProductsInRedZone(ctx context.Context, organizationID uuid.UUID) ([]*providers.BufferStatus, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// In a real implementation, this would call the DDMRP gRPC service
	// For now, we return an empty slice to allow the service to compile and run
	// The actual gRPC call would look like:
	// resp, err := c.ddmrpClient.CheckReplenishment(ctx, &ddmrpv1.CheckReplenishmentRequest{
	//     OrganizationId: organizationID.String(),
	// })

	// Mock implementation - in production, implement actual gRPC call
	return []*providers.BufferStatus{}, nil
}

// Ensure Client implements DDMRPServiceClient.
var _ providers.DDMRPServiceClient = (*Client)(nil)
