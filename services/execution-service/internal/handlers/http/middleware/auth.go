// Package middleware provides HTTP middleware for the Execution Service.
package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// ContextKey is a type for context keys to avoid collisions.
type ContextKey string

const (
	// UserIDKey is the context key for user ID.
	UserIDKey ContextKey = "user_id"
	// OrganizationIDKey is the context key for organization ID.
	OrganizationIDKey ContextKey = "organization_id"
	// EmailKey is the context key for user email.
	EmailKey ContextKey = "email"
)

// AuthClient defines the interface for token validation.
type AuthClient interface {
	ValidateToken(ctx context.Context, token string) (*TokenValidationResult, error)
}

// TokenValidationResult represents the result of token validation.
type TokenValidationResult struct {
	Valid          bool
	UserID         uuid.UUID
	OrganizationID uuid.UUID
	Email          string
	Reason         string
}

// AuthMiddleware provides JWT authentication middleware.
type AuthMiddleware struct {
	authClient AuthClient
}

// NewAuthMiddleware creates a new auth middleware.
func NewAuthMiddleware(authClient AuthClient) *AuthMiddleware {
	return &AuthMiddleware{
		authClient: authClient,
	}
}

// Authenticate validates JWT tokens and adds user context.
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			m.respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authorization header")
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			m.respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid authorization header format, expected 'Bearer <token>'")
			return
		}

		result, err := m.authClient.ValidateToken(r.Context(), token)
		if err != nil {
			m.respondError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "authentication service unavailable")
			return
		}

		if !result.Valid {
			m.respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid or expired token")
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, result.UserID)
		ctx = context.WithValue(ctx, OrganizationIDKey, result.OrganizationID)
		ctx = context.WithValue(ctx, EmailKey, result.Email)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *AuthMiddleware) respondError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error_code": code,
		"message":    message,
	})
}

// GetUserID retrieves the user ID from context.
func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(UserIDKey).(uuid.UUID)
	return userID, ok
}

// GetOrganizationID retrieves the organization ID from context.
func GetOrganizationID(ctx context.Context) (uuid.UUID, bool) {
	orgID, ok := ctx.Value(OrganizationIDKey).(uuid.UUID)
	return orgID, ok
}

// GetEmail retrieves the email from context.
func GetEmail(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(EmailKey).(string)
	return email, ok
}

// RequireAuth is a helper that extracts user and org IDs or returns an error.
func RequireAuth(ctx context.Context) (userID, orgID uuid.UUID, err error) {
	userID, ok := GetUserID(ctx)
	if !ok || userID == uuid.Nil {
		return uuid.Nil, uuid.Nil, ErrUserNotAuthenticated
	}

	orgID, ok = GetOrganizationID(ctx)
	if !ok || orgID == uuid.Nil {
		return uuid.Nil, uuid.Nil, ErrOrganizationNotFound
	}

	return userID, orgID, nil
}

// AuthError represents an authentication error.
type AuthError struct {
	Code    string
	Message string
}

func (e *AuthError) Error() string {
	return e.Message
}

var (
	// ErrUserNotAuthenticated is returned when user is not authenticated.
	ErrUserNotAuthenticated = &AuthError{Code: "UNAUTHORIZED", Message: "user not authenticated"}
	// ErrOrganizationNotFound is returned when organization is not found in context.
	ErrOrganizationNotFound = &AuthError{Code: "BAD_REQUEST", Message: "organization not found in context"}
)
