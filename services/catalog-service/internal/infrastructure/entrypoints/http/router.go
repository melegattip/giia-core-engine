package http

import (
	"github.com/giia/giia-core-engine/pkg/logger"
	"github.com/giia/giia-core-engine/services/catalog-service/internal/infrastructure/entrypoints/http/handlers"
	"github.com/giia/giia-core-engine/services/catalog-service/internal/infrastructure/entrypoints/http/middleware"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

type RouterConfig struct {
	ProductHandler *handlers.ProductHandler
	HealthHandler  *handlers.HealthHandler
	Logger         logger.Logger
}

func NewRouter(config *RouterConfig) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Recoverer)
	r.Use(middleware.Logging(config.Logger))
	r.Use(middleware.TenantMiddleware())

	r.Get("/health", config.HealthHandler.Health)

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/products", func(r chi.Router) {
			r.Post("/", config.ProductHandler.Create)
			r.Get("/", config.ProductHandler.List)
			r.Get("/search", config.ProductHandler.Search)
			r.Get("/{id}", config.ProductHandler.Get)
			r.Put("/{id}", config.ProductHandler.Update)
			r.Delete("/{id}", config.ProductHandler.Delete)
		})
	})

	return r
}
