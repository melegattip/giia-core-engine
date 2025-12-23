package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/giia/giia-core-engine/services/analytics-service/internal/infrastructure/persistence/repositories"
)

const (
	defaultGRPCPort = "50053"
	defaultHTTPPort = "8083"
	defaultDBURL    = "postgresql://postgres:postgres@localhost:5432/analytics_db?sslmode=disable"
)

type Server struct {
	grpcServer *grpc.Server
	httpServer *http.Server
	db         *sql.DB
	kpiRepo    *repositories.PostgresKPIRepository
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

	kpiRepo := repositories.NewPostgresKPIRepository(db)

	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(10*1024*1024),
		grpc.MaxSendMsgSize(10*1024*1024),
	)

	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	reflection.Register(grpcServer)

	httpPort := getEnv("HTTP_PORT", defaultHTTPPort)
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"analytics-service"}`))
	})
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("# Analytics Service Metrics\n"))
	})

	httpServer := &http.Server{
		Addr:         ":" + httpPort,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		grpcServer: grpcServer,
		httpServer: httpServer,
		db:         db,
		kpiRepo:    kpiRepo,
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
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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
