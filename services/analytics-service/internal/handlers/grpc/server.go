// Package grpc provides gRPC server setup for the Analytics Service.
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
	Service *AnalyticsService
}

// Server represents the gRPC server.
type Server struct {
	server  *grpc.Server
	config  *ServerConfig
	service *AnalyticsService
}

// NewServer creates a new gRPC server.
func NewServer(config *ServerConfig) *Server {
	server := grpc.NewServer(
		grpc.UnaryInterceptor(loggingInterceptor),
		grpc.MaxRecvMsgSize(10*1024*1024),
		grpc.MaxSendMsgSize(10*1024*1024),
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

	fmt.Printf("🚀 Analytics Service gRPC server starting on %s\n", addr)
	return s.server.Serve(listener)
}

// GracefulStop stops the gRPC server gracefully.
func (s *Server) GracefulStop() {
	s.server.GracefulStop()
}

// Stop stops the gRPC server immediately.
func (s *Server) Stop() {
	s.server.Stop()
}

// GetServer returns the underlying gRPC server (for registration).
func (s *Server) GetServer() *grpc.Server {
	return s.server
}

// GetAnalyticsService returns the analytics service.
func (s *Server) GetAnalyticsService() *AnalyticsService {
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
