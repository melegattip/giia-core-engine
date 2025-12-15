package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/giia/giia-core-engine/pkg/events"
	"github.com/giia/giia-core-engine/pkg/logger"
	"github.com/giia/giia-core-engine/services/catalog-service/internal/core/domain"
	"github.com/giia/giia-core-engine/services/catalog-service/internal/core/usecases/product"
	eventAdapters "github.com/giia/giia-core-engine/services/catalog-service/internal/infrastructure/adapters/events"
	"github.com/giia/giia-core-engine/services/catalog-service/internal/infrastructure/config"
	httpEntrypoint "github.com/giia/giia-core-engine/services/catalog-service/internal/infrastructure/entrypoints/http"
	"github.com/giia/giia-core-engine/services/catalog-service/internal/infrastructure/entrypoints/http/handlers"
	"github.com/giia/giia-core-engine/services/catalog-service/internal/infrastructure/repositories"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	appLogger := logger.New("catalog-service", cfg.Logging.Level)

	db, err := initDatabase(cfg, appLogger)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	natsConn, err := events.ConnectWithDefaults(cfg.NATS.URL)
	if err != nil {
		return fmt.Errorf("failed to connect to NATS: %w", err)
	}
	defer natsConn.Close()

	eventPublisher, err := events.NewPublisher(natsConn)
	if err != nil {
		return fmt.Errorf("failed to create event publisher: %w", err)
	}

	productRepo := repositories.NewProductRepository(db)

	catalogPublisher := eventAdapters.NewCatalogEventPublisher(eventPublisher, appLogger)

	createProductUC := product.NewCreateProductUseCase(productRepo, catalogPublisher, appLogger)
	getProductUC := product.NewGetProductUseCase(productRepo, appLogger)
	updateProductUC := product.NewUpdateProductUseCase(productRepo, catalogPublisher, appLogger)
	deleteProductUC := product.NewDeleteProductUseCase(productRepo, catalogPublisher, appLogger)
	listProductsUC := product.NewListProductsUseCase(productRepo, appLogger)
	searchProductsUC := product.NewSearchProductsUseCase(productRepo, appLogger)

	productHandler := handlers.NewProductHandler(
		createProductUC,
		getProductUC,
		updateProductUC,
		deleteProductUC,
		listProductsUC,
		searchProductsUC,
		appLogger,
	)

	healthHandler := handlers.NewHealthHandler(db, appLogger)

	router := httpEntrypoint.NewRouter(&httpEntrypoint.RouterConfig{
		ProductHandler: productHandler,
		HealthHandler:  healthHandler,
		Logger:         appLogger,
	})

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		appLogger.Info(context.Background(), fmt.Sprintf("Starting Catalog Service on %s", server.Addr), logger.Tags{
			"environment": cfg.Server.Environment,
			"port":        cfg.Server.Port,
		})
		serverErrors <- server.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)
	case sig := <-shutdown:
		appLogger.Info(context.Background(), fmt.Sprintf("Received shutdown signal: %v", sig), nil)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			server.Close()
			return fmt.Errorf("graceful shutdown failed: %w", err)
		}

		appLogger.Info(context.Background(), "Catalog Service shut down gracefully", nil)
	}

	return nil
}

func initDatabase(cfg *config.Config, log logger.Logger) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Exec(fmt.Sprintf("SET search_path TO %s", cfg.Database.Schema)).Error; err != nil {
		return nil, fmt.Errorf("failed to set search path: %w", err)
	}

	if err := db.AutoMigrate(
		&domain.Product{},
		&domain.Supplier{},
		&domain.ProductSupplier{},
		&domain.BufferProfile{},
	); err != nil {
		return nil, fmt.Errorf("failed to run auto migrations: %w", err)
	}

	log.Info(context.Background(), "Database connected successfully", logger.Tags{
		"host":   cfg.Database.Host,
		"schema": cfg.Database.Schema,
	})

	return db, nil
}
