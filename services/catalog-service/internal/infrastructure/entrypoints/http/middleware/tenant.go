package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

func TenantMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			orgIDStr := r.Header.Get("X-Organization-ID")
			if orgIDStr != "" {
				if orgID, err := uuid.Parse(orgIDStr); err == nil {
					ctx := context.WithValue(r.Context(), "organization_id", orgID)
					r = r.WithContext(ctx)
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
