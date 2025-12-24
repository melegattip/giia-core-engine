// Package auth provides an authentication client adapter for the Auth Service.
package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/melegattip/giia-core-engine/services/execution-service/internal/handlers/http/middleware"
)

// Client implements the middleware.AuthClient interface.
type Client struct {
	conn    *grpc.ClientConn
	address string
	timeout time.Duration
}

// ClientConfig holds the configuration for the Auth client.
type ClientConfig struct {
	Address string
	Timeout time.Duration
}

// NewClient creates a new Auth service client.
func NewClient(config *ClientConfig) (*Client, error) {
	if config.Timeout == 0 {
		config.Timeout = 5 * time.Second
	}

	conn, err := grpc.NewClient(
		config.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to auth service: %w", err)
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

// ValidateToken validates a JWT token via the Auth service.
func (c *Client) ValidateToken(ctx context.Context, token string) (*middleware.TokenValidationResult, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// In a real implementation, this would call the Auth gRPC service
	// For now, we return a mock response to allow the service to compile and run
	// The actual gRPC call would look like:
	// resp, err := c.authClient.ValidateToken(ctx, &authv1.ValidateTokenRequest{
	//     Token: token,
	// })

	// Mock implementation - in production, implement actual gRPC call
	// For development/testing, we parse a mock token or return a valid response
	if token == "" {
		return &middleware.TokenValidationResult{
			Valid:  false,
			Reason: "empty token",
		}, nil
	}

	// For development, return a valid response with mock UUIDs
	// In production, this should validate against the Auth service
	return &middleware.TokenValidationResult{
		Valid:          true,
		UserID:         uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		OrganizationID: uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		Email:          "user@example.com",
	}, nil
}

// Ensure Client implements middleware.AuthClient.
var _ middleware.AuthClient = (*Client)(nil)
