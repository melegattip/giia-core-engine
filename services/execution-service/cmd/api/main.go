package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/melegattip/giia-core-engine/services/execution-service/internal/infrastructure/config"
	postgresRepo "github.com/melegattip/giia-core-engine/services/execution-service/internal/repository/postgres"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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

	db, err := initDatabase(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize repositories
	poRepo := postgresRepo.NewPurchaseOrderRepository(db)
	soRepo := postgresRepo.NewSalesOrderRepository(db)
	invTxnRepo := postgresRepo.NewInventoryTransactionRepository(db)
	invBalRepo := postgresRepo.NewInventoryBalanceRepository(db)
	alertRepo := postgresRepo.NewAlertRepository(db)

	// Log repository initialization (these will be used by use cases)
	_ = poRepo
	_ = soRepo
	_ = invTxnRepo
	_ = invBalRepo
	_ = alertRepo

	// Setup HTTP handlers
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "ok",
			"service": "execution-service",
			"time":    time.Now().UTC().Format(time.RFC3339),
		})
	})

	// Readiness check endpoint (checks database connection)
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		sqlDB, err := db.DB()
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "error",
				"error":  "database connection error",
			})
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := sqlDB.PingContext(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "error",
				"error":  "database ping failed",
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "ready",
			"service":  "execution-service",
			"database": "connected",
		})
	})

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:      mux,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	// Start server in goroutine
	serverErrors := make(chan error, 1)
	go func() {
		fmt.Printf("🚀 Execution Service starting on %s:%s\n", cfg.Server.Host, cfg.Server.Port)
		fmt.Printf("   Environment: %s\n", cfg.Server.Environment)
		fmt.Printf("   Health: http://%s:%s/health\n", cfg.Server.Host, cfg.Server.Port)
		fmt.Printf("   Ready: http://%s:%s/ready\n", cfg.Server.Host, cfg.Server.Port)
		serverErrors <- server.ListenAndServe()
	}()

	// Wait for shutdown signal
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)
	case sig := <-shutdown:
		fmt.Printf("\n📥 Received shutdown signal: %v\n", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			server.Close()
			return fmt.Errorf("graceful shutdown failed: %w", err)
		}

		// Close database connection
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}

		fmt.Println("✅ Execution Service shut down gracefully")
	}

	return nil
}

func initDatabase(cfg *config.Config) (*gorm.DB, error) {
	gormLogger := logger.Default.LogMode(logger.Silent)
	if cfg.Server.Environment == "development" {
		gormLogger = logger.Default.LogMode(logger.Info)
	}

	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.Database.ConnMaxLifetime) * time.Second)

	// Set schema
	if err := db.Exec(fmt.Sprintf("SET search_path TO %s", cfg.Database.Schema)).Error; err != nil {
		return nil, fmt.Errorf("failed to set search path: %w", err)
	}

	fmt.Printf("📦 Database connected: %s@%s/%s (schema: %s)\n",
		cfg.Database.User, cfg.Database.Host, cfg.Database.Name, cfg.Database.Schema)

	return db, nil
}
