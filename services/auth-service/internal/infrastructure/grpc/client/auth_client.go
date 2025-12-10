package client

import (
	"context"
	"fmt"
	"time"

	authv1 "github.com/giia/giia-core-engine/services/auth-service/api/proto/gen/go/auth/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type AuthClient struct {
	conn   *grpc.ClientConn
	client authv1.AuthServiceClient
}

type ClientConfig struct {
	Address        string
	Timeout        time.Duration
	MaxRecvMsgSize int
	MaxSendMsgSize int
}

func NewAuthClient(cfg *ClientConfig) (*AuthClient, error) {
	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second
	}
	if cfg.MaxRecvMsgSize == 0 {
		cfg.MaxRecvMsgSize = 1024 * 1024 * 4 // 4MB
	}
	if cfg.MaxSendMsgSize == 0 {
		cfg.MaxSendMsgSize = 1024 * 1024 * 4 // 4MB
	}

	conn, err := grpc.NewClient(
		cfg.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(cfg.MaxRecvMsgSize),
			grpc.MaxCallSendMsgSize(cfg.MaxSendMsgSize),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to auth service: %w", err)
	}

	client := authv1.NewAuthServiceClient(conn)

	return &AuthClient{
		conn:   conn,
		client: client,
	}, nil
}

func (c *AuthClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *AuthClient) ValidateToken(ctx context.Context, token string, requestID string) (*authv1.ValidateTokenResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if requestID != "" {
		md := metadata.Pairs("x-request-id", requestID)
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	req := &authv1.ValidateTokenRequest{
		Token: token,
	}

	return c.client.ValidateToken(ctx, req)
}

func (c *AuthClient) CheckPermission(ctx context.Context, userID, permission, requestID string) (*authv1.CheckPermissionResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if requestID != "" {
		md := metadata.Pairs("x-request-id", requestID)
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	req := &authv1.CheckPermissionRequest{
		UserId:     userID,
		Permission: permission,
	}

	return c.client.CheckPermission(ctx, req)
}

func (c *AuthClient) BatchCheckPermissions(ctx context.Context, userID string, permissions []string, requestID string) (*authv1.BatchCheckPermissionsResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if requestID != "" {
		md := metadata.Pairs("x-request-id", requestID)
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	req := &authv1.BatchCheckPermissionsRequest{
		UserId:      userID,
		Permissions: permissions,
	}

	return c.client.BatchCheckPermissions(ctx, req)
}

func (c *AuthClient) GetUser(ctx context.Context, userID, organizationID, requestID string) (*authv1.GetUserResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if requestID != "" {
		md := metadata.Pairs("x-request-id", requestID)
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	req := &authv1.GetUserRequest{
		UserId:         userID,
		OrganizationId: organizationID,
	}

	return c.client.GetUser(ctx, req)
}

type ConnectionPool struct {
	address string
	size    int
	clients []*AuthClient
	current int
}

func NewConnectionPool(address string, size int) (*ConnectionPool, error) {
	if size <= 0 {
		size = 10
	}

	pool := &ConnectionPool{
		address: address,
		size:    size,
		clients: make([]*AuthClient, size),
		current: 0,
	}

	for i := 0; i < size; i++ {
		client, err := NewAuthClient(&ClientConfig{
			Address: address,
		})
		if err != nil {
			pool.Close()
			return nil, fmt.Errorf("failed to create client %d: %w", i, err)
		}
		pool.clients[i] = client
	}

	return pool, nil
}

func (p *ConnectionPool) GetClient() *AuthClient {
	client := p.clients[p.current]
	p.current = (p.current + 1) % p.size
	return client
}

func (p *ConnectionPool) Close() error {
	var lastErr error
	for _, client := range p.clients {
		if client != nil {
			if err := client.Close(); err != nil {
				lastErr = err
			}
		}
	}
	return lastErr
}
