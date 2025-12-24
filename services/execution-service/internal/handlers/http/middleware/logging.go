// Package middleware provides HTTP middleware for the Execution Service.
package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// Logger defines the interface for logging.
type Logger interface {
	Info(ctx interface{}, msg string, tags map[string]interface{})
	Warn(ctx interface{}, msg string, tags map[string]interface{})
	Error(ctx interface{}, err error, msg string, tags map[string]interface{})
}

// Logging returns a middleware that logs HTTP requests.
func Logging(log Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			duration := time.Since(start)

			tags := map[string]interface{}{
				"method":     r.Method,
				"path":       r.URL.Path,
				"status":     ww.Status(),
				"duration":   duration.Milliseconds(),
				"request_id": middleware.GetReqID(r.Context()),
				"bytes":      ww.BytesWritten(),
			}

			if userID, ok := GetUserID(r.Context()); ok {
				tags["user_id"] = userID.String()
			}

			if orgID, ok := GetOrganizationID(r.Context()); ok {
				tags["organization_id"] = orgID.String()
			}

			if ww.Status() >= 500 {
				log.Error(r.Context(), nil, "HTTP request error", tags)
			} else if ww.Status() >= 400 {
				log.Warn(r.Context(), "HTTP request warning", tags)
			} else {
				log.Info(r.Context(), "HTTP request", tags)
			}
		})
	}
}
