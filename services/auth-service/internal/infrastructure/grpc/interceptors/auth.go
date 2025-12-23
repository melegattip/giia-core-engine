package interceptors

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	pkgErrors "github.com/melegattip/giia-core-engine/pkg/errors"
	pkgLogger "github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/auth-service/internal/core/providers"
)

type contextKey string

const (
	contextKeyUserID         contextKey = "user_id"
	contextKeyOrganizationID contextKey = "organization_id"
	contextKeyEmail          contextKey = "email"
	contextKeyRoles          contextKey = "roles"
)

type AuthInterceptor struct {
	jwtManager providers.JWTManager
	logger     pkgLogger.Logger
}

func NewAuthInterceptor(jwtManager providers.JWTManager, logger pkgLogger.Logger) *AuthInterceptor {
	return &AuthInterceptor{
		jwtManager: jwtManager,
		logger:     logger,
	}
}

func (i *AuthInterceptor) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if isPublicMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		token, err := extractToken(ctx)
		if err != nil {
			i.logger.Warn(ctx, "Missing or invalid authorization header", pkgLogger.Tags{
				"method": info.FullMethod,
				"error":  err.Error(),
			})
			return nil, status.Error(codes.Unauthenticated, "missing authentication token")
		}

		claims, err := i.jwtManager.ValidateAccessToken(token)
		if err != nil {
			i.logger.Warn(ctx, "Token validation failed", pkgLogger.Tags{
				"method": info.FullMethod,
				"error":  err.Error(),
			})
			return nil, status.Error(codes.Unauthenticated, "invalid or expired token")
		}

		ctx = context.WithValue(ctx, contextKeyUserID, claims.UserID)
		ctx = context.WithValue(ctx, contextKeyOrganizationID, claims.OrganizationID)
		ctx = context.WithValue(ctx, contextKeyEmail, claims.Email)
		ctx = context.WithValue(ctx, contextKeyRoles, claims.Roles)

		i.logger.Info(ctx, "Request authenticated", pkgLogger.Tags{
			"method":          info.FullMethod,
			"user_id":         claims.UserID,
			"organization_id": claims.OrganizationID,
		})

		return handler(ctx, req)
	}
}

func (i *AuthInterceptor) StreamInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		if isPublicMethod(info.FullMethod) {
			return handler(srv, ss)
		}

		token, err := extractToken(ss.Context())
		if err != nil {
			i.logger.Warn(ss.Context(), "Missing or invalid authorization header", pkgLogger.Tags{
				"method": info.FullMethod,
				"error":  err.Error(),
			})
			return status.Error(codes.Unauthenticated, "missing authentication token")
		}

		claims, err := i.jwtManager.ValidateAccessToken(token)
		if err != nil {
			i.logger.Warn(ss.Context(), "Token validation failed", pkgLogger.Tags{
				"method": info.FullMethod,
				"error":  err.Error(),
			})
			return status.Error(codes.Unauthenticated, "invalid or expired token")
		}

		ctx := ss.Context()
		ctx = context.WithValue(ctx, contextKeyUserID, claims.UserID)
		ctx = context.WithValue(ctx, contextKeyOrganizationID, claims.OrganizationID)
		ctx = context.WithValue(ctx, contextKeyEmail, claims.Email)
		ctx = context.WithValue(ctx, contextKeyRoles, claims.Roles)

		wrappedStream := &authenticatedStream{
			ServerStream: ss,
			ctx:          ctx,
		}

		i.logger.Info(ctx, "Stream authenticated", pkgLogger.Tags{
			"method":          info.FullMethod,
			"user_id":         claims.UserID,
			"organization_id": claims.OrganizationID,
		})

		return handler(srv, wrappedStream)
	}
}

func extractToken(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", pkgErrors.NewUnauthorized("missing metadata")
	}

	values := md.Get("authorization")
	if len(values) == 0 {
		return "", pkgErrors.NewUnauthorized("missing authorization header")
	}

	token := values[0]
	if !strings.HasPrefix(token, "Bearer ") {
		return "", pkgErrors.NewUnauthorized("invalid authorization format")
	}

	return strings.TrimPrefix(token, "Bearer "), nil
}

func isPublicMethod(method string) bool {
	publicMethods := map[string]bool{
		"/auth.v1.AuthService/ValidateToken": true,
		"/grpc.health.v1.Health/Check":       true,
		"/grpc.health.v1.Health/Watch":       true,
	}
	return publicMethods[method]
}

type authenticatedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *authenticatedStream) Context() context.Context {
	return s.ctx
}
