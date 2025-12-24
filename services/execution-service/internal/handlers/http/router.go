// Package http provides HTTP routing for the Execution Service.
package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/melegattip/giia-core-engine/services/execution-service/internal/handlers/http/handlers"
	"github.com/melegattip/giia-core-engine/services/execution-service/internal/handlers/http/middleware"
)

// RouterConfig holds the configuration for setting up the router.
type RouterConfig struct {
	PurchaseOrderHandler *handlers.PurchaseOrderHandler
	SalesOrderHandler    *handlers.SalesOrderHandler
	InventoryHandler     *handlers.InventoryHandler
	AuthMiddleware       *middleware.AuthMiddleware
	Logger               middleware.Logger
	ServiceName          string
	Version              string
}

// NewRouter creates a new Chi router with all routes and middleware.
func NewRouter(config *RouterConfig) *chi.Mux {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.Timeout(60 * time.Second))

	// Custom logging middleware
	if config.Logger != nil {
		r.Use(middleware.Logging(config.Logger))
	}

	// Tenant middleware (fallback for org ID from header)
	r.Use(middleware.TenantMiddleware())

	// Health check endpoints (no auth required)
	r.Get("/health", healthHandler(config.ServiceName, config.Version))
	r.Get("/ready", readyHandler(config.ServiceName))

	// API routes with authentication
	r.Route("/api/v1", func(r chi.Router) {
		// Apply auth middleware to all API routes
		if config.AuthMiddleware != nil {
			r.Use(config.AuthMiddleware.Authenticate)
		}

		// Purchase Orders
		r.Route("/purchase-orders", func(r chi.Router) {
			r.Post("/", config.PurchaseOrderHandler.Create)
			r.Get("/", config.PurchaseOrderHandler.List)
			r.Get("/{id}", config.PurchaseOrderHandler.Get)
			r.Post("/{id}/receive", config.PurchaseOrderHandler.Receive)
			r.Post("/{id}/cancel", config.PurchaseOrderHandler.Cancel)
		})

		// Sales Orders
		r.Route("/sales-orders", func(r chi.Router) {
			r.Post("/", config.SalesOrderHandler.Create)
			r.Get("/", config.SalesOrderHandler.List)
			r.Get("/{id}", config.SalesOrderHandler.Get)
			r.Post("/{id}/ship", config.SalesOrderHandler.Ship)
			r.Post("/{id}/cancel", config.SalesOrderHandler.Cancel)
		})

		// Inventory
		r.Route("/inventory", func(r chi.Router) {
			r.Get("/balances", config.InventoryHandler.GetBalances)
			r.Get("/transactions", config.InventoryHandler.GetTransactions)
		})
	})

	return r
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
		// In a real implementation, this would check database and other dependencies
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "ready",
			"service": serviceName,
		})
	}
}
