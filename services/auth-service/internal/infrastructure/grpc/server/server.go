package server

import (
	"net"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"gorm.io/gorm"

	pkgLogger "github.com/giia/giia-core-engine/pkg/logger"
	authv1 "github.com/giia/giia-core-engine/services/auth-service/api/proto/gen/go/auth/v1"
	"github.com/giia/giia-core-engine/services/auth-service/internal/core/providers"
	"github.com/giia/giia-core-engine/services/auth-service/internal/core/usecases/auth"
	"github.com/giia/giia-core-engine/services/auth-service/internal/core/usecases/rbac"
	"github.com/giia/giia-core-engine/services/auth-service/internal/infrastructure/grpc/interceptors"
)

type GRPCServer struct {
	server        *grpc.Server
	listener      net.Listener
	authService   *AuthServiceServer
	healthService *HealthServiceServer
	logger        pkgLogger.Logger
}

func NewGRPCServer(
	port string,
	validateTokenUC *auth.ValidateTokenUseCase,
	checkPermissionUC *rbac.CheckPermissionUseCase,
	batchCheckUC *rbac.BatchCheckPermissionsUseCase,
	getUserPermissionsUC *rbac.GetUserPermissionsUseCase,
	userRepo providers.UserRepository,
	db *gorm.DB,
	redisClient *redis.Client,
	logger pkgLogger.Logger,
) (*GRPCServer, error) {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		return nil, err
	}

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptors.LoggingInterceptor(logger),
			interceptors.RecoveryInterceptor(logger),
			interceptors.MetricsInterceptor(),
		),
	)

	authService := NewAuthServiceServer(
		validateTokenUC,
		checkPermissionUC,
		batchCheckUC,
		getUserPermissionsUC,
		userRepo,
		logger,
	)

	healthService := NewHealthServiceServer(db, redisClient, logger)

	authv1.RegisterAuthServiceServer(server, authService)
	grpc_health_v1.RegisterHealthServer(server, healthService)
	reflection.Register(server)

	return &GRPCServer{
		server:        server,
		listener:      listener,
		authService:   authService,
		healthService: healthService,
		logger:        logger,
	}, nil
}

func (s *GRPCServer) Start() error {
	s.logger.Info(nil, "Starting gRPC server", pkgLogger.Tags{
		"address": s.listener.Addr().String(),
	})
	return s.server.Serve(s.listener)
}

func (s *GRPCServer) Stop() {
	s.logger.Info(nil, "Stopping gRPC server gracefully", pkgLogger.Tags{})
	s.server.GracefulStop()
}
