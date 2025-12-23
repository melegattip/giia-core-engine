package auth

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/melegattip/giia-core-engine/pkg/logger"
	authpb "github.com/melegattip/giia-core-engine/services/auth-service/api/proto/gen/go/auth/v1"
	"github.com/melegattip/giia-core-engine/services/catalog-service/internal/core/providers"
	"github.com/google/uuid"
)

type grpcAuthClient struct {
	client authpb.AuthServiceClient
	conn   *grpc.ClientConn
	logger logger.Logger
}

func NewGRPCAuthClient(authServiceURL string, log logger.Logger) (providers.AuthClient, error) {
	if authServiceURL == "" {
		return nil, fmt.Errorf("auth service URL is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		authServiceURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to auth service at %s: %w", authServiceURL, err)
	}

	client := authpb.NewAuthServiceClient(conn)

	log.Info(context.Background(), "Connected to Auth service successfully", logger.Tags{
		"auth_service_url": authServiceURL,
	})

	return &grpcAuthClient{
		client: client,
		conn:   conn,
		logger: log,
	}, nil
}

func (c *grpcAuthClient) ValidateToken(ctx context.Context, token string) (*providers.TokenValidationResult, error) {
	if token == "" {
		return &providers.TokenValidationResult{
			Valid:  false,
			Reason: "token is empty",
		}, nil
	}

	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := c.client.ValidateToken(reqCtx, &authpb.ValidateTokenRequest{
		Token: token,
	})
	if err != nil {
		c.logger.Error(ctx, err, "Failed to validate token with auth service", nil)
		return nil, fmt.Errorf("auth service error: %w", err)
	}

	if !resp.Valid {
		return &providers.TokenValidationResult{
			Valid:  false,
			Reason: resp.Reason,
		}, nil
	}

	if resp.User == nil {
		return &providers.TokenValidationResult{
			Valid:  false,
			Reason: "user information not available",
		}, nil
	}

	userID, err := uuid.Parse(resp.User.UserId)
	if err != nil {
		c.logger.Error(ctx, err, "Invalid user ID format from auth service", logger.Tags{
			"user_id": resp.User.UserId,
		})
		return &providers.TokenValidationResult{
			Valid:  false,
			Reason: "invalid user ID format",
		}, nil
	}

	orgID, err := uuid.Parse(resp.User.OrganizationId)
	if err != nil {
		c.logger.Error(ctx, err, "Invalid organization ID format from auth service", logger.Tags{
			"organization_id": resp.User.OrganizationId,
		})
		return &providers.TokenValidationResult{
			Valid:  false,
			Reason: "invalid organization ID format",
		}, nil
	}

	return &providers.TokenValidationResult{
		Valid:          true,
		UserID:         userID,
		OrganizationID: orgID,
		Email:          resp.User.Email,
		Reason:         "",
	}, nil
}

func (c *grpcAuthClient) CheckPermission(ctx context.Context, userID uuid.UUID, orgID uuid.UUID, permission string) (bool, error) {
	if userID == uuid.Nil {
		return false, fmt.Errorf("user ID is required")
	}
	if orgID == uuid.Nil {
		return false, fmt.Errorf("organization ID is required")
	}
	if permission == "" {
		return false, fmt.Errorf("permission is required")
	}

	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := c.client.CheckPermission(reqCtx, &authpb.CheckPermissionRequest{
		UserId:         userID.String(),
		OrganizationId: orgID.String(),
		Permission:     permission,
	})
	if err != nil {
		c.logger.Error(ctx, err, "Failed to check permission with auth service", logger.Tags{
			"user_id":    userID.String(),
			"permission": permission,
		})
		return false, fmt.Errorf("auth service error: %w", err)
	}

	return resp.Allowed, nil
}

func (c *grpcAuthClient) Close() error {
	if c.conn != nil {
		c.logger.Info(context.Background(), "Closing Auth service connection", nil)
		return c.conn.Close()
	}
	return nil
}
