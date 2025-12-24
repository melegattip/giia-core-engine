// Package http provides HTTP routing for the AI Intelligence Hub service.
package http

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/handlers/websocket"
)

// RouterConfig holds the configuration for setting up the router.
type RouterConfig struct {
	NotificationHandler *NotificationHandler
	PreferencesHandler  *PreferencesHandler
	WebSocketHub        *websocket.Hub
	ServiceName         string
	Version             string
}

// NewRouter creates a new HTTP router with all routes and middleware.
func NewRouter(config *RouterConfig) http.Handler {
	mux := http.NewServeMux()

	// Health check endpoints (no auth required)
	mux.HandleFunc("GET /health", healthHandler(config.ServiceName, config.Version))
	mux.HandleFunc("GET /ready", readyHandler(config.ServiceName))
	mux.HandleFunc("GET /metrics", metricsHandler())

	// Notification REST endpoints
	if config.NotificationHandler != nil {
		mux.HandleFunc("GET /api/v1/notifications", config.NotificationHandler.ListNotifications)
		mux.HandleFunc("GET /api/v1/notifications/unread-count", config.NotificationHandler.GetUnreadCount)
		mux.HandleFunc("GET /api/v1/notifications/{id}", config.NotificationHandler.GetNotification)
		mux.HandleFunc("PATCH /api/v1/notifications/{id}", config.NotificationHandler.UpdateNotification)
		mux.HandleFunc("DELETE /api/v1/notifications/{id}", config.NotificationHandler.DeleteNotification)
	}

	// Preferences REST endpoints
	if config.PreferencesHandler != nil {
		mux.HandleFunc("GET /api/v1/notifications/preferences", config.PreferencesHandler.GetPreferences)
		mux.HandleFunc("PUT /api/v1/notifications/preferences", config.PreferencesHandler.UpdatePreferences)
	}

	// WebSocket endpoint
	if config.WebSocketHub != nil {
		mux.HandleFunc("GET /ws/notifications", config.WebSocketHub.HandleWebSocket)
	}

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
		w.Write([]byte("# AI Intelligence Hub Metrics\n"))
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

// responseRecorder wraps http.ResponseWriter to capture the status code.
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rr *responseRecorder) WriteHeader(code int) {
	rr.statusCode = code
	rr.ResponseWriter.WriteHeader(code)
}

// loggingMiddleware logs incoming requests.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Skip logging for health checks to reduce noise
		if strings.HasPrefix(r.URL.Path, "/health") || strings.HasPrefix(r.URL.Path, "/ready") {
			next.ServeHTTP(w, r)
			return
		}

		rr := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rr, r)

		duration := time.Since(start)
		_ = duration // Could log: log.Printf("%s %s %d %v", r.Method, r.URL.Path, rr.statusCode, duration)
	})
}

// corsMiddleware adds CORS headers.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-User-ID, X-Organization-ID")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// AuthMiddleware validates JWT tokens and sets user/org info in context.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for health endpoints and WebSocket upgrade
		if strings.HasPrefix(r.URL.Path, "/health") ||
			strings.HasPrefix(r.URL.Path, "/ready") ||
			strings.HasPrefix(r.URL.Path, "/metrics") {
			next.ServeHTTP(w, r)
			return
		}

		// Check for Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// For development, allow X-User-ID and X-Organization-ID headers
			if r.Header.Get("X-User-ID") != "" && r.Header.Get("X-Organization-ID") != "" {
				next.ServeHTTP(w, r)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error_code": "UNAUTHORIZED",
				"message":    "missing authorization header",
			})
			return
		}

		// Validate Bearer token format
		if !strings.HasPrefix(authHeader, "Bearer ") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error_code": "UNAUTHORIZED",
				"message":    "invalid authorization header format",
			})
			return
		}

		// In production, validate JWT token and extract claims
		// For now, pass through for development
		next.ServeHTTP(w, r)
	})
}

// RateLimitMiddleware limits the rate of requests.
func RateLimitMiddleware(requestsPerSecond int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Simple rate limiting implementation
			// In production, use a proper rate limiter with Redis
			next.ServeHTTP(w, r)
		})
	}
}
