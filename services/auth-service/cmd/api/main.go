package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	pkgDatabase "github.com/melegattip/giia-core-engine/pkg/database"
	pkgErrors "github.com/melegattip/giia-core-engine/pkg/errors"
	pkgLogger "github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/auth-service/internal/infrastructure/config"
	grpcInit "github.com/melegattip/giia-core-engine/services/auth-service/internal/infrastructure/grpc/initialization"
	httpInit "github.com/melegattip/giia-core-engine/services/auth-service/internal/infrastructure/http/initialization"
	"github.com/melegattip/giia-core-engine/services/auth-service/pkg/database"
	"github.com/nats-io/nats.go"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Load configuration
	cfg := config.Load()

	// Log the database configuration
	log.Printf("🔧 [Database] Config: Host=%s, Port=%s, User=%s, DBName=%s, SSLMode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.DBName, cfg.Database.SSLMode)

	// Initialize database
	dbConfig := database.Config{
		DSN:             cfg.GetDatabaseDSN(),
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 2 * time.Minute,
	}

	log.Printf("🔧 [Database] DSN: %s", dbConfig.DSN)

	// Adjust connection pool settings based on pool mode
	if cfg.Database.PoolMode == "transaction" {
		// For transaction pooling, we can use more connections
		// since the pooler handles the actual database connections
		dbConfig.MaxOpenConns = 50
		dbConfig.MaxIdleConns = 10
		log.Printf("🔧 [Database] Using transaction pooling mode (MaxOpenConns: %d, MaxIdleConns: %d)",
			dbConfig.MaxOpenConns, dbConfig.MaxIdleConns)
	} else {
		// Use standard connection pool settings
		dbConfig.MaxOpenConns = 25
		dbConfig.MaxIdleConns = 5
		log.Printf("🔧 [Database] Using standard connection pool (MaxOpenConns: %d, MaxIdleConns: %d)",
			dbConfig.MaxOpenConns, dbConfig.MaxIdleConns)
	}

	db, err := database.NewConnection(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run database migrations
	if err := db.RunMigrations(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.GetRedisAddr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Test Redis connection
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Printf("⚠️  [Redis] Warning: Failed to connect to Redis: %v", err)
		log.Printf("⚠️  [Redis] gRPC server will not start without Redis")
	} else {
		log.Printf("✅ [Redis] Connected successfully to %s", cfg.GetRedisAddr())
	}

	// Initialize logger
	logger := pkgLogger.New("auth-service", "info")

	// Initialize GORM database for gRPC server
	gormDB, err := connectGormDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect GORM database: %v", err)
	}
	defer func() {
		sqlDB, _ := gormDB.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	// Initialize NATS connection (optional for event publishing)
	var natsConn *nats.Conn
	natsURL := getEnvOrDefault("NATS_URL", "")
	if natsURL != "" {
		natsConn, err = nats.Connect(natsURL)
		if err != nil {
			log.Printf("⚠️  [NATS] Warning: Failed to connect to NATS: %v", err)
			log.Printf("⚠️  [NATS] Event publishing will be disabled")
		} else {
			log.Printf("✅ [NATS] Connected successfully to %s", natsURL)
			defer natsConn.Close()
		}
	}

	// Initialize gRPC server
	grpcPort := getEnvOrDefault("GRPC_PORT", "9091")
	grpcContainer, err := grpcInit.InitializeGRPCServer(&grpcInit.GRPCConfig{
		Port:             fmt.Sprintf(":%s", grpcPort),
		JWTSecretKey:     cfg.JWT.SecretKey,
		JWTAccessExpiry:  cfg.JWT.AccessExpiry,
		JWTRefreshExpiry: cfg.JWT.RefreshExpiry,
		JWTIssuer:        cfg.JWT.Issuer,
		DB:               gormDB,
		RedisClient:      redisClient,
		Logger:           logger,
	})
	if err != nil {
		log.Fatalf("Failed to initialize gRPC server: %v", err)
	}

	// Initialize HTTP server
	httpPort := getEnvOrDefault("HTTP_PORT", "8080")
	baseURL := getEnvOrDefault("BASE_URL", fmt.Sprintf("http://localhost:%s", httpPort))
	httpContainer, err := httpInit.InitializeHTTPServer(&httpInit.HTTPConfig{
		Port:             httpPort,
		JWTSecretKey:     cfg.JWT.SecretKey,
		JWTAccessExpiry:  cfg.JWT.AccessExpiry,
		JWTRefreshExpiry: cfg.JWT.RefreshExpiry,
		JWTIssuer:        cfg.JWT.Issuer,
		SMTPHost:         cfg.Email.SMTPHost,
		SMTPPort:         fmt.Sprintf("%d", cfg.Email.SMTPPort),
		SMTPUsername:     cfg.Email.SMTPUser,
		SMTPPassword:     cfg.Email.SMTPPassword,
		SMTPFrom:         cfg.Email.FromEmail,
		BaseURL:          baseURL,
		DB:               gormDB,
		RedisClient:      redisClient,
		NATSConn:         natsConn,
		Logger:           logger,
	})
	if err != nil {
		log.Fatalf("Failed to initialize HTTP server: %v", err)
	}

	// Start gRPC server in a goroutine
	go func() {
		log.Printf("🚀 [gRPC Server] Starting on :%s", grpcPort)
		if err := grpcContainer.Server.Start(); err != nil {
			log.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	// Start HTTP server in a goroutine
	go func() {
		log.Printf("🚀 [HTTP Server] Starting on :%s", httpPort)
		if err := httpContainer.Server.Start(); err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 [Auth Service] Shutting down servers...")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := httpContainer.Server.Stop(shutdownCtx); err != nil {
		log.Printf("⚠️  [HTTP Server] Error during shutdown: %v", err)
	} else {
		log.Println("✅ [HTTP Server] Server shutdown gracefully")
	}

	// Shutdown gRPC server
	grpcContainer.Server.Stop()
	log.Println("✅ [gRPC Server] Server shutdown gracefully")

	// Close NATS connection
	if natsConn != nil {
		natsConn.Close()
		log.Println("✅ [NATS] Connection closed")
	}

	// Close Redis connection
	if err := redisClient.Close(); err != nil {
		log.Printf("⚠️  [Redis] Error closing connection: %v", err)
	} else {
		log.Println("✅ [Redis] Connection closed")
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func connectGormDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := cfg.GetDatabaseDSN()
	gormDB, err := pkgDatabase.ConnectWithDSN(context.Background(), dsn)
	if err != nil {
		return nil, pkgErrors.NewInternalServerError("failed to connect to database")
	}
	return gormDB, nil
}
