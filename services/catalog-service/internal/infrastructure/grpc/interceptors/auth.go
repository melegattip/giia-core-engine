package interceptors

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/catalog-service/internal/core/providers"
)

type AuthInterceptor struct {
	authClient providers.AuthClient
	logger     logger.Logger
}

func NewAuthInterceptor(authClient providers.AuthClient, logger logger.Logger) *AuthInterceptor {
	return &AuthInterceptor{
		authClient: authClient,
		logger:     logger,
	}
}

func (i *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			i.logger.Warn(ctx, "Missing metadata in gRPC request", nil)
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		authHeaders := md.Get("authorization")
		if len(authHeaders) == 0 {
			i.logger.Warn(ctx, "Missing authorization header in gRPC request", nil)
			return nil, status.Error(codes.Unauthenticated, "missing authorization token")
		}

		token := strings.TrimPrefix(authHeaders[0], "Bearer ")
		if token == authHeaders[0] {
			i.logger.Warn(ctx, "Invalid authorization header format", nil)
			return nil, status.Error(codes.Unauthenticated, "invalid authorization header format")
		}

		result, err := i.authClient.ValidateToken(ctx, token)
		if err != nil {
			i.logger.Error(ctx, err, "Failed to validate token", nil)
			return nil, status.Error(codes.Internal, "authentication service unavailable")
		}

		if !result.Valid {
			i.logger.Warn(ctx, "Invalid token", logger.Tags{
				"reason": result.Reason,
			})
			return nil, status.Error(codes.Unauthenticated, "invalid or expired token")
		}

		ctx = context.WithValue(ctx, "user_id", result.UserID)
		ctx = context.WithValue(ctx, "organization_id", result.OrganizationID)
		ctx = context.WithValue(ctx, "email", result.Email)

		i.logger.Info(ctx, "gRPC user authenticated successfully", logger.Tags{
			"user_id":         result.UserID.String(),
			"organization_id": result.OrganizationID.String(),
			"method":          info.FullMethod,
		})

		return handler(ctx, req)
	}
}
