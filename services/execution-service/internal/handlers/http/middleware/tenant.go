// Package middleware provides HTTP middleware for the Execution Service.
package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// TenantMiddleware extracts organization ID from X-Organization-ID header.
// This is a fallback for when the organization ID is not in the JWT.
func TenantMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only set from header if not already set by auth middleware
			if _, ok := GetOrganizationID(r.Context()); !ok {
				orgIDStr := r.Header.Get("X-Organization-ID")
				if orgIDStr != "" {
					if orgID, err := uuid.Parse(orgIDStr); err == nil {
						ctx := context.WithValue(r.Context(), OrganizationIDKey, orgID)
						r = r.WithContext(ctx)
					}
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
