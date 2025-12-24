package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	httpPkg "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/melegattip/giia-core-engine/services/analytics-service/internal/adapters/catalog"
	"github.com/melegattip/giia-core-engine/services/analytics-service/internal/adapters/ddmrp"
	"github.com/melegattip/giia-core-engine/services/analytics-service/internal/adapters/execution"
	"github.com/melegattip/giia-core-engine/services/analytics-service/internal/core/usecases/kpi"
	grpcHandlers "github.com/melegattip/giia-core-engine/services/analytics-service/internal/handlers/grpc"
	"github.com/melegattip/giia-core-engine/services/analytics-service/internal/handlers/http"
	"github.com/melegattip/giia-core-engine/services/analytics-service/internal/handlers/http/handlers"
	"github.com/melegattip/giia-core-engine/services/analytics-service/internal/infrastructure/persistence/repositories"
)

const (
	defaultGRPCPort      = "50053"
	defaultHTTPPort      = "8083"
	defaultDBURL         = "postgresql://postgres:postgres@localhost:5432/analytics_db?sslmode=disable"
	defaultCatalogAddr   = "localhost:50051"
	defaultDDMRPAddr     = "localhost:50052"
	defaultExecutionAddr = "localhost:50054"
	serviceName          = "analytics-service"
	serviceVersion       = "1.0.0"
)

type Server struct {
	grpcServer      *grpc.Server
	httpServer      *httpPkg.Server
	db              *sql.DB
	kpiRepo         *repositories.PostgresKPIRepository
	catalogClient   *catalog.Client
	ddmrpClient     *ddmrp.Client
	executionClient *execution.Client
}

func main() {
	log.Println("Starting Analytics Service...")

	server, err := NewServer()
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}
	defer server.Shutdown()

	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func NewServer() (*Server, error) {
	// Database connection
	dbURL := getEnv("DATABASE_URL", defaultDBURL)
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Printf("Warning: Database ping failed: %v", err)
		log.Println("Continuing with HTTP-only mode...")
	} else {
		log.Println("Database connection established")
	}

	// Initialize repository
	kpiRepo := repositories.NewPostgresKPIRepository(db)

	// Initialize service adapters
	catalogClient, err := catalog.NewClient(&catalog.ClientConfig{
		Address:    getEnv("CATALOG_SERVICE_ADDR", defaultCatalogAddr),
		Timeout:    5 * time.Second,
		MaxRetries: 3,
	})
	if err != nil {
		log.Printf("Warning: Failed to create catalog client: %v", err)
	}

	ddmrpClient, err := ddmrp.NewClient(&ddmrp.ClientConfig{
		Address:    getEnv("DDMRP_SERVICE_ADDR", defaultDDMRPAddr),
		Timeout:    5 * time.Second,
		MaxRetries: 3,
	})
	if err != nil {
		log.Printf("Warning: Failed to create DDMRP client: %v", err)
	}

	executionClient, err := execution.NewClient(&execution.ClientConfig{
		Address:    getEnv("EXECUTION_SERVICE_ADDR", defaultExecutionAddr),
		Timeout:    5 * time.Second,
		MaxRetries: 3,
	})
	if err != nil {
		log.Printf("Warning: Failed to create execution client: %v", err)
	}

	// Initialize use cases
	diiUseCase := kpi.NewCalculateDaysInInventoryUseCase(kpiRepo, catalogClient)
	immobilizedUseCase := kpi.NewCalculateImmobilizedInventoryUseCase(kpiRepo, catalogClient)
	rotationUseCase := kpi.NewCalculateInventoryRotationUseCase(kpiRepo, executionClient)
	syncBufferUseCase := kpi.NewSyncBufferAnalyticsUseCase(kpiRepo, ddmrpClient)

	// Initialize HTTP handlers
	kpiHandler := handlers.NewKPIHandler(
		kpiRepo,
		diiUseCase,
		immobilizedUseCase,
		rotationUseCase,
		syncBufferUseCase,
	)

	// Create HTTP router
	httpRouter := http.NewRouter(&http.RouterConfig{
		KPIHandler:  kpiHandler,
		ServiceName: serviceName,
		Version:     serviceVersion,
	})

	httpPort := getEnv("HTTP_PORT", defaultHTTPPort)
	httpServer := &httpPkg.Server{
		Addr:         ":" + httpPort,
		Handler:      httpRouter,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Initialize gRPC server
	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(10*1024*1024),
		grpc.MaxSendMsgSize(10*1024*1024),
	)

	// Initialize gRPC Analytics Service
	analyticsService := grpcHandlers.NewAnalyticsService(
		kpiRepo,
		diiUseCase,
		immobilizedUseCase,
		rotationUseCase,
		syncBufferUseCase,
	)

	// Register health service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus(serviceName, grpc_health_v1.HealthCheckResponse_SERVING)

	// Register reflection for grpcurl
	reflection.Register(grpcServer)

	// Note: In production, register the actual proto-generated service
	// analyticsv1.RegisterAnalyticsServiceServer(grpcServer, analyticsService)
	_ = analyticsService // TODO: Register with generated proto code

	return &Server{
		grpcServer:      grpcServer,
		httpServer:      httpServer,
		db:              db,
		kpiRepo:         kpiRepo,
		catalogClient:   catalogClient,
		ddmrpClient:     ddmrpClient,
		executionClient: executionClient,
	}, nil
}

func (s *Server) Start() error {
	grpcPort := getEnv("GRPC_PORT", defaultGRPCPort)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		return fmt.Errorf("failed to listen on gRPC port %s: %w", grpcPort, err)
	}

	log.Printf("Analytics Service gRPC server listening on port %s", grpcPort)
	log.Printf("Analytics Service HTTP server listening on port %s", s.httpServer.Addr)

	errChan := make(chan error, 2)

	go func() {
		if err := s.grpcServer.Serve(lis); err != nil {
			errChan <- fmt.Errorf("gRPC server error: %w", err)
		}
	}()

	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != httpPkg.ErrServerClosed {
			errChan <- fmt.Errorf("HTTP server error: %w", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errChan:
		return err
	case sig := <-sigChan:
		log.Printf("Received signal %v, shutting down gracefully...", sig)
		return nil
	}
}

func (s *Server) Shutdown() {
	log.Println("Shutting down Analytics Service...")

	if s.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.httpServer.Shutdown(ctx)
	}

	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}

	// Close service adapters
	if s.catalogClient != nil {
		s.catalogClient.Close()
	}
	if s.ddmrpClient != nil {
		s.ddmrpClient.Close()
	}
	if s.executionClient != nil {
		s.executionClient.Close()
	}

	if s.db != nil {
		s.db.Close()
	}

	log.Println("Analytics Service stopped")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
