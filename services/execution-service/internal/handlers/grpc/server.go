// Package grpc provides gRPC server setup for the Execution Service.
package grpc

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// ServerConfig holds the configuration for the gRPC server.
type ServerConfig struct {
	Host    string
	Port    string
	Service *ExecutionService
}

// Server represents the gRPC server.
type Server struct {
	server  *grpc.Server
	config  *ServerConfig
	service *ExecutionService
}

// NewServer creates a new gRPC server.
func NewServer(config *ServerConfig) *Server {
	server := grpc.NewServer(
		grpc.UnaryInterceptor(loggingInterceptor),
	)

	s := &Server{
		server:  server,
		config:  config,
		service: config.Service,
	}

	// Register reflection for grpcurl
	reflection.Register(server)

	return s
}

// Start starts the gRPC server.
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%s", s.config.Host, s.config.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	fmt.Printf("🚀 Execution Service gRPC server starting on %s\n", addr)
	return s.server.Serve(listener)
}

// GracefulStop stops the gRPC server gracefully.
func (s *Server) GracefulStop() {
	s.server.GracefulStop()
}

// GetExecutionService returns the execution service.
func (s *Server) GetExecutionService() *ExecutionService {
	return s.service
}

// loggingInterceptor logs gRPC calls.
func loggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	resp, err := handler(ctx, req)
	if err != nil {
		fmt.Printf("gRPC error: method=%s error=%v\n", info.FullMethod, err)
	}
	return resp, err
}
