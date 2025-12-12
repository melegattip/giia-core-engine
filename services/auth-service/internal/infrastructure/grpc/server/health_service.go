package server

import (
	"context"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/health/grpc_health_v1"
	"gorm.io/gorm"

	pkgLogger "github.com/giia/giia-core-engine/pkg/logger"
)

type HealthServiceServer struct {
	grpc_health_v1.UnimplementedHealthServer
	db          *gorm.DB
	redisClient *redis.Client
	logger      pkgLogger.Logger
}

func NewHealthServiceServer(db *gorm.DB, redisClient *redis.Client, logger pkgLogger.Logger) *HealthServiceServer {
	return &HealthServiceServer{
		db:          db,
		redisClient: redisClient,
		logger:      logger,
	}
}

func (s *HealthServiceServer) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	if err := s.db.Exec("SELECT 1").Error; err != nil {
		s.logger.Error(ctx, err, "Database health check failed", pkgLogger.Tags{})
		return &grpc_health_v1.HealthCheckResponse{
			Status: grpc_health_v1.HealthCheckResponse_NOT_SERVING,
		}, nil
	}

	if err := s.redisClient.Ping(ctx).Err(); err != nil {
		s.logger.Error(ctx, err, "Redis health check failed", pkgLogger.Tags{})
		return &grpc_health_v1.HealthCheckResponse{
			Status: grpc_health_v1.HealthCheckResponse_NOT_SERVING,
		}, nil
	}

	return &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	}, nil
}

func (s *HealthServiceServer) Watch(req *grpc_health_v1.HealthCheckRequest, stream grpc_health_v1.Health_WatchServer) error {
	return stream.Send(&grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	})
}
