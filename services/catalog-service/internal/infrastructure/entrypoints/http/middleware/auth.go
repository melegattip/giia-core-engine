package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/melegattip/giia-core-engine/pkg/errors"
	"github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/catalog-service/internal/core/providers"
)

type AuthMiddleware struct {
	authClient providers.AuthClient
	logger     logger.Logger
}

func NewAuthMiddleware(authClient providers.AuthClient, logger logger.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		authClient: authClient,
		logger:     logger,
	}
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			m.respondError(w, r, errors.NewUnauthorized("missing authorization header"))
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			m.respondError(w, r, errors.NewUnauthorized("invalid authorization header format, expected 'Bearer <token>'"))
			return
		}

		result, err := m.authClient.ValidateToken(r.Context(), token)
		if err != nil {
			m.logger.Error(r.Context(), err, "Auth service error during token validation", nil)
			m.respondError(w, r, errors.NewInternalServerError("authentication service unavailable"))
			return
		}

		if !result.Valid {
			m.logger.Warn(r.Context(), "Invalid token", logger.Tags{
				"reason": result.Reason,
			})
			m.respondError(w, r, errors.NewUnauthorized("invalid or expired token"))
			return
		}

		ctx := context.WithValue(r.Context(), "user_id", result.UserID)
		ctx = context.WithValue(ctx, "organization_id", result.OrganizationID)
		ctx = context.WithValue(ctx, "email", result.Email)

		m.logger.Info(r.Context(), "User authenticated successfully", logger.Tags{
			"user_id":         result.UserID.String(),
			"organization_id": result.OrganizationID.String(),
		})

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *AuthMiddleware) respondError(w http.ResponseWriter, r *http.Request, err error) {
	w.Header().Set("Content-Type", "application/json")

	statusCode := http.StatusInternalServerError
	if errors.IsUnauthorized(err) {
		statusCode = http.StatusUnauthorized
	}

	w.WriteHeader(statusCode)
	w.Write([]byte(`{"error":"` + err.Error() + `"}`))
}
