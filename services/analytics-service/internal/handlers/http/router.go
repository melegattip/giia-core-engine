// Package http provides HTTP routing for the Analytics Service.
package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/melegattip/giia-core-engine/services/analytics-service/internal/handlers/http/handlers"
)

// RouterConfig holds the configuration for setting up the router.
type RouterConfig struct {
	KPIHandler  *handlers.KPIHandler
	ServiceName string
	Version     string
}

// NewRouter creates a new HTTP router with all routes and middleware.
func NewRouter(config *RouterConfig) http.Handler {
	mux := http.NewServeMux()

	// Health check endpoints (no auth required)
	mux.HandleFunc("GET /health", healthHandler(config.ServiceName, config.Version))
	mux.HandleFunc("GET /ready", readyHandler(config.ServiceName))
	mux.HandleFunc("GET /metrics", metricsHandler())

	// API routes
	mux.HandleFunc("GET /api/v1/analytics/days-in-inventory", config.KPIHandler.GetDaysInInventory)
	mux.HandleFunc("GET /api/v1/analytics/immobilized-inventory", config.KPIHandler.GetImmobilizedInventory)
	mux.HandleFunc("GET /api/v1/analytics/inventory-rotation", config.KPIHandler.GetInventoryRotation)
	mux.HandleFunc("GET /api/v1/analytics/buffer-analytics", config.KPIHandler.GetBufferAnalytics)
	mux.HandleFunc("GET /api/v1/analytics/snapshot", config.KPIHandler.GetSnapshot)
	mux.HandleFunc("POST /api/v1/analytics/sync-buffer", config.KPIHandler.SyncBufferData)

	// Apply middleware
	handler := recoveryMiddleware(mux)
	handler = loggingMiddleware(handler)
	handler = corsMiddleware(handler)

	return handler
}

// healthHandler returns a health check handler.
func healthHandler(serviceName, version string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "ok",
			"service": serviceName,
			"version": version,
			"time":    time.Now().UTC().Format(time.RFC3339),
		})
	}
}

// readyHandler returns a readiness check handler.
func readyHandler(serviceName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "ready",
			"service": serviceName,
		})
	}
}

// metricsHandler returns a metrics handler.
func metricsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("# Analytics Service Metrics\n"))
	}
}

// recoveryMiddleware catches panics and returns 500 errors.
func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error_code": "INTERNAL_ERROR",
					"message":    "internal server error",
				})
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware logs incoming requests.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)
		_ = duration // Could log: log.Printf("%s %s %v", r.Method, r.URL.Path, duration)
	})
}

// corsMiddleware adds CORS headers.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Organization-ID")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
